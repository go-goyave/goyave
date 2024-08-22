package goyave

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"runtime"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/cors"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/validation"

	_ "goyave.dev/goyave/v5/database/dialect/sqlite"
)

func TestMiddlewareHolder(t *testing.T) {
	m1 := &recoveryMiddleware{}
	m2 := &languageMiddleware{}
	holder := middlewareHolder{
		middleware: []Middleware{m1, m2},
	}
	expected := []Middleware{m1, m2}

	assert.Equal(t, expected, holder.GetMiddleware())
}

func TestHasMiddleware(t *testing.T) {
	t.Run("findMiddleware", func(t *testing.T) {
		m := &recoveryMiddleware{}
		holder := []Middleware{m}

		assert.Equal(t, m, findMiddleware[*recoveryMiddleware](holder))
		assert.Nil(t, findMiddleware[*languageMiddleware](holder))
	})

	t.Run("routeHasMiddleware", func(t *testing.T) {
		route := &Route{
			parent: &Router{
				middlewareHolder: middlewareHolder{
					middleware: []Middleware{&languageMiddleware{}},
				},
			},
			middlewareHolder: middlewareHolder{
				middleware: []Middleware{&recoveryMiddleware{}},
			},
		}

		assert.True(t, routeHasMiddleware[*recoveryMiddleware](route))
		assert.False(t, routeHasMiddleware[*languageMiddleware](route))
	})

	t.Run("routerHasMiddleware", func(t *testing.T) {
		router := &Router{
			parent: &Router{
				globalMiddleware: &middlewareHolder{
					middleware: []Middleware{&testMiddleware{}},
				},
				middlewareHolder: middlewareHolder{
					middleware: []Middleware{&languageMiddleware{}},
				},
			},
			globalMiddleware: &middlewareHolder{},
			middlewareHolder: middlewareHolder{
				middleware: []Middleware{&recoveryMiddleware{}},
			},
		}

		assert.True(t, routerHasMiddleware[*recoveryMiddleware](router))
		assert.True(t, routerHasMiddleware[*languageMiddleware](router))
		assert.True(t, routerHasMiddleware[*testMiddleware](router))
		assert.False(t, routerHasMiddleware[*corsMiddleware](router))
	})
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Run("panic", func(t *testing.T) {
		logBuffer := &bytes.Buffer{}
		server, err := New(Options{Config: config.LoadDefault(), Logger: slog.New(slog.NewHandler(false, logBuffer))})
		if err != nil {
			panic(err)
		}
		middleware := &recoveryMiddleware{}
		middleware.Init(server)

		panicErr := fmt.Errorf("test error")
		handler := middleware.Handle(func(_ *Response, _ *Request) {
			panic(panicErr)
		})

		request := NewRequest(httptest.NewRequest(http.MethodGet, "/test", nil))
		response := NewResponse(server, request, httptest.NewRecorder())

		handler(response, request)

		returnedErr := response.GetError()
		if !assert.NotNil(t, returnedErr) {
			return
		}
		assert.Equal(t, []error{panicErr}, returnedErr.Unwrap())
		assert.Regexp(t, regexp.MustCompile(
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"%s","trace":%s}\n`,
				regexp.QuoteMeta(returnedErr.Error()), regexp.QuoteMeta(string(lo.Must(json.Marshal(returnedErr.StackFrames().String())))),
			)),
			logBuffer.String(),
		)
		assert.Equal(t, http.StatusInternalServerError, response.status)
	})

	t.Run("no_panic", func(t *testing.T) {
		logBuffer := &bytes.Buffer{}
		server, err := New(Options{Config: config.LoadDefault(), Logger: slog.New(slog.NewHandler(false, logBuffer))})
		if err != nil {
			panic(err)
		}
		middleware := &recoveryMiddleware{}
		middleware.Init(server)

		handler := middleware.Handle(func(_ *Response, _ *Request) {})

		request := NewRequest(httptest.NewRequest(http.MethodGet, "/test", nil))
		response := NewResponse(server, request, httptest.NewRecorder())

		handler(response, request)

		assert.Empty(t, logBuffer.String())
		assert.Nil(t, response.err)
		assert.Equal(t, 0, response.status)
	})

	t.Run("nil_panic", func(t *testing.T) {
		logBuffer := &bytes.Buffer{}
		server, err := New(Options{Config: config.LoadDefault(), Logger: slog.New(slog.NewHandler(false, logBuffer))})
		if err != nil {
			panic(err)
		}
		middleware := &recoveryMiddleware{}
		middleware.Init(server)

		handler := middleware.Handle(func(_ *Response, _ *Request) {
			//nolint:govet
			panic(nil)
		})

		request := NewRequest(httptest.NewRequest(http.MethodGet, "/test", nil))
		response := NewResponse(server, request, httptest.NewRecorder())

		handler(response, request)

		returnedErr := response.GetError()
		if !assert.NotNil(t, returnedErr) {
			return
		}
		assert.Equal(t, []error{&runtime.PanicNilError{}}, returnedErr.Unwrap())
		assert.Regexp(t, regexp.MustCompile(
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"%s","trace":%s}\n`,
				regexp.QuoteMeta(returnedErr.Error()), regexp.QuoteMeta(string(lo.Must(json.Marshal(returnedErr.StackFrames().String())))),
			)),
			logBuffer.String(),
		)
		assert.Equal(t, http.StatusInternalServerError, response.status)
	})

	t.Run("panic_status_override", func(t *testing.T) {
		// Even if the response status is already set, the recovery middleware
		// should always force it to 500.
		logBuffer := &bytes.Buffer{}
		server, err := New(Options{Config: config.LoadDefault(), Logger: slog.New(slog.NewHandler(false, logBuffer))})
		if err != nil {
			panic(err)
		}
		middleware := &recoveryMiddleware{}
		middleware.Init(server)

		handler := middleware.Handle(func(r *Response, _ *Request) {
			r.JSON(http.StatusOK, make(chan struct{})) // Unsupported type for JSON encoding, status is set to 200 before writing
		})

		request := NewRequest(httptest.NewRequest(http.MethodGet, "/test", nil))
		response := NewResponse(server, request, httptest.NewRecorder())

		handler(response, request)

		returnedErr := response.GetError()
		if !assert.NotNil(t, returnedErr) {
			return
		}
		assert.Len(t, returnedErr.Unwrap(), 1)
		assert.Regexp(t, regexp.MustCompile(
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"%s","trace":%s}\n`,
				regexp.QuoteMeta(returnedErr.Error()), regexp.QuoteMeta(string(lo.Must(json.Marshal(returnedErr.StackFrames().String())))),
			)),
			logBuffer.String(),
		)
		assert.Equal(t, http.StatusInternalServerError, response.status)
	})
}

func TestLanguageMiddleware(t *testing.T) {
	server, err := New(Options{Config: config.LoadDefault()})
	require.NoError(t, err)
	middleware := &languageMiddleware{}
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
		t.Run(c.desc, func(t *testing.T) {
			executed := false
			handler := middleware.Handle(func(_ *Response, req *Request) {
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
		next              func(*Response, *Request)
		queryRules        func(*Request) validation.RuleSet
		bodyRules         func(*Request) validation.RuleSet
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
			queryRules: func(_ *Request) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Int(), validation.Min(5)}}}
			},
			query:        map[string]any{"param": "6"},
			expectBody:   "OK",
			expectPass:   true,
			expectStatus: http.StatusOK,
			next: func(_ *Response, r *Request) {
				assert.Equal(t, map[string]any{"param": 6}, r.Query)
			},
		},
		{
			desc: "query_nok",
			queryRules: func(_ *Request) validation.RuleSet {
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
			queryRules: func(_ *Request) validation.RuleSet {
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
			queryRules: func(request *Request) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), &testValidator{
					validateFunc: func(v *testValidator, ctx *validation.Context) bool {
						assert.Equal(t, request, ctx.Extra[validation.ExtraRequest{}])
						assert.NotNil(t, v.DB())
						assert.NotNil(t, v.Config())
						assert.NotNil(t, v.Logger())
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
			queryRules: func(_ *Request) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Array()}}}
			},
			query:        map[string]any{"param": "v"},
			expectPass:   true,
			expectStatus: http.StatusOK,
			expectBody:   "OK",
			next: func(_ *Response, r *Request) {
				assert.Equal(t, map[string]any{"param": []string{"v"}}, r.Query)
			},
		},
		{
			desc: "body_ok",
			bodyRules: func(_ *Request) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Int(), validation.Min(5)}}}
			},
			data:         map[string]any{"param": "6"},
			expectBody:   "OK",
			expectPass:   true,
			expectStatus: http.StatusOK,
			next: func(_ *Response, r *Request) {
				assert.Equal(t, map[string]any{"param": 6}, r.Data)
			},
		},
		{
			desc: "body_nok",
			bodyRules: func(_ *Request) validation.RuleSet {
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
			bodyRules: func(_ *Request) validation.RuleSet {
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
			bodyRules: func(request *Request) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), &testValidator{
					validateFunc: func(v *testValidator, ctx *validation.Context) bool {
						assert.Equal(t, request, ctx.Extra[validation.ExtraRequest{}])
						assert.NotNil(t, v.Config())
						assert.NotNil(t, v.DB())
						assert.NotNil(t, v.Logger())
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
			bodyRules: func(_ *Request) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Array()}}}
			},
			data:         map[string]any{"param": "v"},
			expectPass:   true,
			expectStatus: http.StatusOK,
			expectBody:   "OK",
			next: func(_ *Response, r *Request) {
				assert.Equal(t, map[string]any{"param": []string{"v"}}, r.Data)
			},
		},
		{
			desc: "body_dont_convert_single_value_arrays",
			bodyRules: func(_ *Request) validation.RuleSet {
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
			queryRules: func(_ *Request) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Int(), validation.Min(5)}}}
			},
			bodyRules: func(_ *Request) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), validation.Int(), validation.Min(5)}}}
			},
			query:        map[string]any{"param": "6"},
			data:         map[string]any{"param": "7"},
			expectBody:   "OK",
			expectPass:   true,
			expectStatus: http.StatusOK,
			next: func(_ *Response, r *Request) {
				assert.Equal(t, map[string]any{"param": 7}, r.Data)
				assert.Equal(t, map[string]any{"param": 6}, r.Query)
			},
		},
		{
			desc: "query_and_body_error",
			queryRules: func(_ *Request) validation.RuleSet {
				return validation.RuleSet{{Path: "param", Rules: validation.List{validation.Required(), &testValidator{
					validateFunc: func(_ *testValidator, ctx *validation.Context) bool {
						ctx.AddError(fmt.Errorf("test error 1"))
						return true
					},
				}}}}
			},
			bodyRules: func(_ *Request) validation.RuleSet {
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
		t.Run(c.desc, func(t *testing.T) {
			cfg := config.LoadDefault()
			if c.hasDB {
				cfg.Set("database.connection", "sqlite3")
				cfg.Set("database.name", fmt.Sprintf("test_validation_middleware_%s.db", c.desc))
				cfg.Set("database.options", "mode=memory")
			}
			buffer := &bytes.Buffer{}
			server, err := New(Options{Config: cfg, Logger: slog.New(slog.NewHandler(false, buffer))})
			if err != nil {
				panic(err)
			}
			defer func() {
				assert.NoError(t, server.CloseDB())
			}()

			m := &validateRequestMiddleware{
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
			m.Handle(func(r *Response, req *Request) {
				pass = true
				if c.next != nil {
					c.next(r, req)
				}
				r.String(http.StatusOK, "OK")
			})(response, request)

			res := recorder.Result()
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, res.Body.Close())
			assert.NoError(t, err)
			assert.Equal(t, c.expectPass, pass)
			if c.expectPass {
				assert.Equal(t, "OK", string(body))
			}
			assert.Equal(t, c.expectStatus, response.status)

			if c.expectQueryErrors == nil {
				assert.NotContains(t, request.Extra, ExtraQueryValidationError{})
			} else {
				assert.Equal(t, c.expectQueryErrors, request.Extra[ExtraQueryValidationError{}])
			}
			if c.expectBodyErrors == nil {
				assert.NotContains(t, request.Extra, ExtraValidationError{})
			} else {
				assert.Equal(t, c.expectBodyErrors, request.Extra[ExtraValidationError{}])
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	cases := []struct {
		options            func() *cors.Options
		req                func() *Request
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
			req: func() *Request {
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
			req: func() *Request {
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
			req: func() *Request {
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
			req: func() *Request {
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
			req: func() *Request {
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
		t.Run(c.desc, func(t *testing.T) {
			middleware := &corsMiddleware{}
			handler := middleware.Handle(func(resp *Response, _ *Request) {
				if c.respBody != "" {
					resp.String(c.respStatus, c.respBody)
				} else {
					resp.WriteHeader(c.respStatus)
				}
			})

			request := c.req()
			request.Route = &Route{
				Meta: map[string]any{
					MetaCORS: c.options(),
				},
			}
			recorder := httptest.NewRecorder()
			response := NewResponse(nil, request, recorder)
			match := &routeMatch{
				route: request.Route,
			}

			handler(response, request)
			require.NoError(t, (&Router{}).finalize(match, response, request))
			resp := recorder.Result()
			assert.Equal(t, c.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, c.expectedHeaders, resp.Header)
			defer func() {
				assert.NoError(t, resp.Body.Close())
			}()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, c.expectedBody, string(body))
		})
	}
}
