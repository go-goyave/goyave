package goyave

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/cors"
	"goyave.dev/goyave/v4/validation"

	_ "goyave.dev/goyave/v4/database/dialect/sqlite"
)

func TestMiddlewareHolder(t *testing.T) {

	m1 := &recoveryMiddlewareV5{}
	m2 := &languageMiddlewareV5{}
	holder := middlewareHolderV5{
		middleware: []MiddlewareV5{m1, m2},
	}
	expected := []MiddlewareV5{m1, m2}

	assert.Equal(t, expected, holder.GetMiddleware())
}

func TestHasMiddleware(t *testing.T) {

	t.Run("findMiddleware", func(t *testing.T) {
		m := &recoveryMiddlewareV5{}
		holder := []MiddlewareV5{m}

		assert.Equal(t, m, findMiddleware[*recoveryMiddlewareV5](holder))
		assert.Nil(t, findMiddleware[*languageMiddlewareV5](holder))
	})

	t.Run("routeHasMiddleware", func(t *testing.T) {
		route := &RouteV5{
			parent: &RouterV5{
				middlewareHolderV5: middlewareHolderV5{
					middleware: []MiddlewareV5{&languageMiddlewareV5{}},
				},
			},
			middlewareHolderV5: middlewareHolderV5{
				middleware: []MiddlewareV5{&recoveryMiddlewareV5{}},
			},
		}

		assert.True(t, routeHasMiddleware[*recoveryMiddlewareV5](route))
		assert.False(t, routeHasMiddleware[*languageMiddlewareV5](route))
	})

	t.Run("routerHasMiddleware", func(t *testing.T) {
		router := &RouterV5{
			parent: &RouterV5{
				middlewareHolderV5: middlewareHolderV5{
					middleware: []MiddlewareV5{&languageMiddlewareV5{}},
				},
			},
			middlewareHolderV5: middlewareHolderV5{
				middleware: []MiddlewareV5{&recoveryMiddlewareV5{}},
			},
		}

		assert.True(t, routerHasMiddleware[*recoveryMiddlewareV5](router))
		assert.True(t, routerHasMiddleware[*languageMiddlewareV5](router))
		assert.False(t, routerHasMiddleware[*corsMiddlewareV5](router))
	})
}

func TestRecoveryMiddleware(t *testing.T) {
	// TODO TestRecoveryMiddleware (after the error handling rework)

	t.Run("panic", func(t *testing.T) {
		server, err := NewWithConfig(config.LoadDefault())
		if err != nil {
			panic(err)
		}
		logBuffer := &bytes.Buffer{}
		server.ErrLogger = log.New(logBuffer, "", 0)
		middleware := &recoveryMiddlewareV5{}
		middleware.Init(server)

		panicErr := fmt.Errorf("test error")
		handler := middleware.Handle(func(_ *ResponseV5, _ *RequestV5) {
			panic(panicErr)
		})

		request := NewRequest(httptest.NewRequest(http.MethodGet, "/test", nil))
		response := NewResponse(server, request, httptest.NewRecorder())

		handler(response, request)

		assert.Equal(t, panicErr, request.Extra[ExtraError])
		assert.Equal(t, panicErr.Error()+"\n", logBuffer.String())
		assert.NotEmpty(t, request.Extra[ExtraStacktrace])
		assert.Equal(t, http.StatusInternalServerError, response.status)
	})

	t.Run("no_panic", func(t *testing.T) {
		server, err := NewWithConfig(config.LoadDefault())
		if err != nil {
			panic(err)
		}
		logBuffer := &bytes.Buffer{}
		server.ErrLogger = log.New(logBuffer, "", 0)
		middleware := &recoveryMiddlewareV5{}
		middleware.Init(server)

		handler := middleware.Handle(func(_ *ResponseV5, _ *RequestV5) {})

		request := NewRequest(httptest.NewRequest(http.MethodGet, "/test", nil))
		response := NewResponse(server, request, httptest.NewRecorder())

		handler(response, request)

		assert.Empty(t, logBuffer.String())
		assert.NotContains(t, request.Extra, ExtraError)
		assert.NotContains(t, request.Extra, ExtraStacktrace)
		assert.Equal(t, 0, response.status)
	})

	t.Run("nil_panic", func(t *testing.T) {
		server, err := NewWithConfig(config.LoadDefault())
		if err != nil {
			panic(err)
		}
		logBuffer := &bytes.Buffer{}
		server.ErrLogger = log.New(logBuffer, "", 0)
		middleware := &recoveryMiddlewareV5{}
		middleware.Init(server)

		handler := middleware.Handle(func(_ *ResponseV5, _ *RequestV5) {
			panic(nil)
		})

		request := NewRequest(httptest.NewRequest(http.MethodGet, "/test", nil))
		response := NewResponse(server, request, httptest.NewRecorder())

		handler(response, request)

		assert.Nil(t, request.Extra[ExtraError])
		assert.Contains(t, request.Extra, ExtraError)
		assert.Equal(t, "<nil>\n", logBuffer.String())
		assert.NotEmpty(t, request.Extra[ExtraStacktrace])
		assert.Equal(t, http.StatusInternalServerError, response.status)
	})
}

func TestLanguageMiddleware(t *testing.T) {
	server, err := NewWithConfig(config.LoadDefault())
	if err != nil {
		panic(err)
	}
	middleware := &languageMiddlewareV5{}
	middleware.Init(server)

	cases := []struct {
		desc     string
		lang     string
		expected string
	}{
		{desc: "no_default", lang: "en-UK", expected: "en-UK"},
		{desc: "default_provided", lang: "en-US", expected: "en-US"},
		{desc: "default_not_provided", lang: "en-US", expected: "en-US"},
		{desc: "priority", lang: "en-US;q=0.9, en-UK", expected: "en-UK"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			executed := false
			handler := middleware.Handle(func(resp *ResponseV5, req *RequestV5) {
				assert.Equal(t, c.expected, req.Lang.Name())
				executed = true
			})

			request := NewRequest(httptest.NewRequest(http.MethodGet, "/test", nil))
			if c.lang != "" {
				request.Header().Set("Accept-Language", c.lang)
			}
			response := NewResponse(server, request, httptest.NewRecorder())

			handler(response, request)
			assert.True(t, executed)
		})
	}
}

type testValidator struct {
	validation.BaseValidator
	validateFunc func(c *testValidator, ctx *validation.Context) bool
}

func (v *testValidator) Validate(ctx *validation.Context) bool {
	return v.validateFunc(v, ctx)
}

func (v *testValidator) Name() string {
	return "test_validator"
}

func TestValidateMiddleware(t *testing.T) {

	cases := []struct {
		next              func(*ResponseV5, *RequestV5)
		queryRules        func(*RequestV5) validation.RuleSet
		bodyRules         func(*RequestV5) validation.RuleSet
		headers           map[string]string
		query             map[string]any
		data              any
		expectQueryErrors *validation.Errors
		expectBodyErrors  *validation.Errors
		desc              string
		expectBody        string
		hasDB             bool
		expectPass        bool
		expectStatus      int
	}{
		{
			desc: "query_ok",
			queryRules: func(_ *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Int(), validation.Min(5)}}}
			},
			query:        map[string]any{"param": "6"},
			expectBody:   "OK",
			expectPass:   true,
			expectStatus: http.StatusOK,
			next: func(_ *ResponseV5, r *RequestV5) {
				assert.Equal(t, map[string]any{"param": 6}, r.Query)
			},
		},
		{
			desc: "query_nok",
			queryRules: func(_ *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Min(5)}}}
			},
			query:        map[string]any{"param": "v"},
			expectPass:   false,
			expectStatus: http.StatusUnprocessableEntity,
			expectQueryErrors: &validation.Errors{Fields: validation.FieldsErrors{
				"param": &validation.Errors{Errors: []string{"The param must be at least 5 characters."}},
			}},
		},
		{
			desc: "query_error",
			queryRules: func(_ *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), &testValidator{
					validateFunc: func(_ *testValidator, ctx *validation.Context) bool {
						ctx.AddError(fmt.Errorf("test error 1"), fmt.Errorf("test error 2"))
						return true
					},
				}}}}
			},
			query:        map[string]any{"param": "v"},
			expectPass:   false,
			expectStatus: http.StatusInternalServerError,
			expectBody:   "{\"error\": [\"test error 1\",\"test error 2\"]}",
		},
		{
			desc:  "query_validation_options",
			hasDB: true,
			queryRules: func(request *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), &testValidator{
					validateFunc: func(v *testValidator, ctx *validation.Context) bool {
						assert.Equal(t, request, ctx.Extra[validation.ExtraRequest])
						assert.NotNil(t, v.DB())
						assert.NotNil(t, v.Config())
						assert.NotNil(t, v.Logger())
						assert.NotNil(t, v.ErrLogger())
						assert.NotNil(t, v.Lang())
						return false
					},
				}}}}
			},
			query:        map[string]any{"param": "v"},
			expectPass:   false,
			expectStatus: http.StatusUnprocessableEntity,
			expectQueryErrors: &validation.Errors{Fields: validation.FieldsErrors{
				"param": &validation.Errors{Errors: []string{"validation.rules.test_validator"}},
			}},
		},
		{
			desc: "query_convert_single_value_arrays",
			queryRules: func(request *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Array()}}}
			},
			query:        map[string]any{"param": "v"},
			expectPass:   true,
			expectStatus: http.StatusOK,
			expectBody:   "OK",
			next: func(_ *ResponseV5, r *RequestV5) {
				assert.Equal(t, map[string]any{"param": []string{"v"}}, r.Query)
			},
		},
		{
			desc: "body_ok",
			bodyRules: func(_ *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Int(), validation.Min(5)}}}
			},
			data:         map[string]any{"param": "6"},
			expectBody:   "OK",
			expectPass:   true,
			expectStatus: http.StatusOK,
			next: func(_ *ResponseV5, r *RequestV5) {
				assert.Equal(t, map[string]any{"param": 6}, r.Data)
			},
		},
		{
			desc: "body_nok",
			bodyRules: func(_ *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Min(5)}}}
			},
			data:         map[string]any{"param": "v"},
			expectPass:   false,
			expectStatus: http.StatusUnprocessableEntity,
			expectBodyErrors: &validation.Errors{Fields: validation.FieldsErrors{
				"param": &validation.Errors{Errors: []string{"The param must be at least 5 characters."}},
			}},
		},
		{
			desc: "body_error",
			bodyRules: func(_ *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), &testValidator{
					validateFunc: func(_ *testValidator, ctx *validation.Context) bool {
						ctx.AddError(fmt.Errorf("test error 1"), fmt.Errorf("test error 2"))
						return true
					},
				}}}}
			},
			data:         map[string]any{"param": "v"},
			expectPass:   false,
			expectStatus: http.StatusInternalServerError,
			expectBody:   "{\"error\": [\"test error 1\",\"test error 2\"]}",
		},
		{
			desc:  "body_validation_options",
			hasDB: true,
			bodyRules: func(request *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), &testValidator{
					validateFunc: func(v *testValidator, ctx *validation.Context) bool {
						assert.Equal(t, request, ctx.Extra[validation.ExtraRequest])
						assert.NotNil(t, v.Config())
						assert.NotNil(t, v.DB())
						assert.NotNil(t, v.Logger())
						assert.NotNil(t, v.ErrLogger())
						assert.NotNil(t, v.Lang())
						return false
					},
				}}}}
			},
			data:         map[string]any{"param": "v"},
			expectPass:   false,
			expectStatus: http.StatusUnprocessableEntity,
			expectBodyErrors: &validation.Errors{Fields: validation.FieldsErrors{
				"param": &validation.Errors{Errors: []string{"validation.rules.test_validator"}},
			}},
		},
		{
			desc: "body_convert_single_value_arrays",
			bodyRules: func(request *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Array()}}}
			},
			data:         map[string]any{"param": "v"},
			expectPass:   true,
			expectStatus: http.StatusOK,
			expectBody:   "OK",
			next: func(_ *ResponseV5, r *RequestV5) {
				assert.Equal(t, map[string]any{"param": []string{"v"}}, r.Data)
			},
		},
		{
			desc: "body_dont_convert_single_value_arrays",
			bodyRules: func(request *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Array()}}}
			},
			headers:      map[string]string{"Content-Type": "application/json; charset=utf-8"},
			data:         map[string]any{"param": "v"},
			expectPass:   false,
			expectStatus: http.StatusUnprocessableEntity,
			expectBodyErrors: &validation.Errors{Fields: validation.FieldsErrors{
				"param": &validation.Errors{Errors: []string{"The param must be an array."}},
			}},
		},
		{
			desc: "query_and_body_ok",
			queryRules: func(_ *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Int(), validation.Min(5)}}}
			},
			bodyRules: func(_ *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Int(), validation.Min(5)}}}
			},
			query:        map[string]any{"param": "6"},
			data:         map[string]any{"param": "7"},
			expectBody:   "OK",
			expectPass:   true,
			expectStatus: http.StatusOK,
			next: func(_ *ResponseV5, r *RequestV5) {
				assert.Equal(t, map[string]any{"param": 7}, r.Data)
				assert.Equal(t, map[string]any{"param": 6}, r.Query)
			},
		},
		{
			desc: "query_and_body_error",
			queryRules: func(_ *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), &testValidator{
					validateFunc: func(_ *testValidator, ctx *validation.Context) bool {
						ctx.AddError(fmt.Errorf("test error 1"))
						return true
					},
				}}}}
			},
			bodyRules: func(_ *RequestV5) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), &testValidator{
					validateFunc: func(_ *testValidator, ctx *validation.Context) bool {
						ctx.AddError(fmt.Errorf("test error 2"))
						return true
					},
				}}}}
			},
			data:         map[string]any{"param": "v"},
			query:        map[string]any{"param": "v"},
			expectPass:   false,
			expectStatus: http.StatusInternalServerError,
			expectBody:   "{\"error\": [\"test error 1\",\"test error 2\"]}",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			cfg := config.LoadDefault()
			if c.hasDB {
				cfg.Set("database.connection", "sqlite3")
				cfg.Set("database.name", fmt.Sprintf("test_validation_middleware_%s.db", c.desc))
				cfg.Set("database.options", "mode=memory")
			}
			server, err := NewWithConfig(cfg)
			if err != nil {
				panic(err)
			}
			defer func() {
				assert.NoError(t, server.CloseDB())
			}()
			buffer := &bytes.Buffer{}
			server.ErrLogger = log.New(buffer, "", 0)

			m := &validateRequestMiddlewareV5{
				QueryRules: c.queryRules,
				BodyRules:  c.bodyRules,
			}
			m.Init(server)

			request := NewRequest(httptest.NewRequest(http.MethodGet, "/test", nil))
			request.Lang = server.Lang.GetDefault()
			request.Query = c.query
			request.Data = c.data
			if c.headers != nil {
				for h, v := range c.headers {
					request.httpRequest.Header.Set(h, v)
				}
			}
			recorder := httptest.NewRecorder()
			response := NewResponse(server, request, recorder)

			pass := false
			m.Handle(func(r *ResponseV5, req *RequestV5) {
				pass = true
				if c.next != nil {
					c.next(r, req)
				}
				r.String(http.StatusOK, "OK")
			})(response, request)

			res := recorder.Result()
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, res.Body.Close())
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, c.expectPass, pass)
			if c.expectPass {
				assert.Equal(t, "OK", string(body))
			}
			assert.Equal(t, c.expectStatus, response.status)

			if c.expectQueryErrors == nil {
				assert.NotContains(t, request.Extra, ExtraQueryValidationError)
			} else {
				assert.Equal(t, c.expectQueryErrors, request.Extra[ExtraQueryValidationError])
			}
			if c.expectBodyErrors == nil {
				assert.NotContains(t, request.Extra, ExtraValidationError)
			} else {
				assert.Equal(t, c.expectBodyErrors, request.Extra[ExtraValidationError])
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	cases := []struct {
		options            func() *cors.Options
		req                func() *RequestV5
		expectedHeaders    http.Header
		desc               string
		respBody           string
		expectedBody       string
		respStatus         int
		expectedStatusCode int
	}{
		{
			desc:    "no_options",
			options: func() *cors.Options { return nil },
			req: func() *RequestV5 {
				return NewRequest(httptest.NewRequest(http.MethodGet, "/test", nil))
			},
			respStatus:         http.StatusOK,
			respBody:           "hello world",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "hello world",
			expectedHeaders:    http.Header{},
		},
		{
			desc:    "preflight",
			options: cors.Default,
			req: func() *RequestV5 {
				req := NewRequest(httptest.NewRequest(http.MethodOptions, "/test", nil))
				req.Header().Set("Origin", "https://google.com")
				req.Header().Set("Access-Control-Request-Method", http.MethodGet)
				return req
			},
			respStatus:         http.StatusOK,
			respBody:           "hello world",
			expectedStatusCode: http.StatusNoContent,
			expectedBody:       "",
			expectedHeaders: http.Header{
				"Access-Control-Allow-Headers": []string{"Origin, Accept, Content-Type, X-Requested-With, Authorization"},
				"Access-Control-Allow-Methods": []string{"HEAD, GET, POST, PUT, PATCH, DELETE"},
				"Access-Control-Allow-Origin":  []string{"*"},
				"Access-Control-Max-Age":       []string{"43200"},
			},
		},
		{
			desc: "preflight_passthrough",
			options: func() *cors.Options {
				o := cors.Default()
				o.OptionsPassthrough = true
				return o
			},
			req: func() *RequestV5 {
				req := NewRequest(httptest.NewRequest(http.MethodOptions, "/test", nil))
				req.Header().Set("Access-Control-Request-Method", http.MethodGet)
				return req
			},
			respStatus:         http.StatusOK,
			respBody:           "hello world",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "hello world",
			expectedHeaders: http.Header{
				"Access-Control-Allow-Headers": []string{"Origin, Accept, Content-Type, X-Requested-With, Authorization"},
				"Access-Control-Allow-Methods": []string{"HEAD, GET, POST, PUT, PATCH, DELETE"},
				"Access-Control-Allow-Origin":  []string{"*"},
				"Access-Control-Max-Age":       []string{"43200"},
			},
		},
		{
			desc:    "preflight_without_Access-Control-Request-Method",
			options: cors.Default,
			req: func() *RequestV5 {
				return NewRequest(httptest.NewRequest(http.MethodOptions, "/test", nil))
			},
			respStatus:         http.StatusOK,
			respBody:           "hello world",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "",
			expectedHeaders:    http.Header{},
		},
		{
			desc:    "actual_request",
			options: cors.Default,
			req: func() *RequestV5 {
				return NewRequest(httptest.NewRequest(http.MethodGet, "/test", nil))
			},
			respStatus:         http.StatusOK,
			respBody:           "hello world",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "hello world",
			expectedHeaders: http.Header{
				"Access-Control-Allow-Origin": []string{"*"},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			middleware := &corsMiddlewareV5{}
			handler := middleware.Handle(func(resp *ResponseV5, req *RequestV5) {
				if c.respBody != "" {
					resp.String(c.respStatus, c.respBody)
				} else {
					resp.WriteHeader(c.respStatus)
				}
			})

			request := c.req()
			request.Route = &RouteV5{
				Meta: map[string]any{
					MetaCORS: c.options(),
				},
			}
			recorder := httptest.NewRecorder()
			response := NewResponse(nil, request, recorder)

			handler(response, request)
			assert.NoError(t, (&RouterV5{}).finalize(response, request))
			resp := recorder.Result()
			assert.Equal(t, c.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, c.expectedHeaders, resp.Header)
			defer func() {
				_ = resp.Body.Close()
			}()
			body, err := io.ReadAll(resp.Body)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, c.expectedBody, string(body))
		})
	}
}
