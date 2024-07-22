package goyave

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/validation"
)

func prepareStatusHandlerTest() (*Request, *Response, *httptest.ResponseRecorder) {
	server, err := New(Options{Config: config.LoadDefault()})
	if err != nil {
		panic(err)
	}

	httpReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	req := NewRequest(httpReq)
	req.Lang = server.Lang.GetDefault()

	recorder := httptest.NewRecorder()
	resp := NewResponse(server, req, recorder)
	return req, resp, recorder
}

func TestPanicStatusHandler(t *testing.T) {
	t.Run("no_debug", func(t *testing.T) {
		req, resp, recorder := prepareStatusHandlerTest()
		resp.server.config.Set("app.debug", false)
		handler := &PanicStatusHandler{}
		handler.Init(resp.server)

		resp.err = errors.New("test error").(*errors.Error)
		handler.Handle(resp, req)
		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, res.Body.Close())
		require.NoError(t, err)

		assert.Equal(t, `{"error":"Internal Server Error"}`+"\n", string(body))
	})

	t.Run("debug", func(t *testing.T) {
		req, resp, recorder := prepareStatusHandlerTest()
		resp.server.config.Set("app.debug", true)
		logBuffer := &bytes.Buffer{}
		resp.server.Logger = slog.New(slog.NewHandler(false, logBuffer))
		handler := &PanicStatusHandler{}
		handler.Init(resp.server)

		resp.err = errors.New("test error").(*errors.Error)
		handler.Handle(resp, req)
		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, res.Body.Close())
		require.NoError(t, err)

		assert.Equal(t, `{"error":"test error"}`+"\n", string(body))

		// Error and stacktrace already printed by the recovery middleware or `response.Error`
		// (those are not executed in this test, thus leaving the log buffer empty)
		assert.Empty(t, logBuffer.String())
	})

	t.Run("nil_error", func(t *testing.T) {
		req, resp, recorder := prepareStatusHandlerTest()
		resp.server.config.Set("app.debug", true)
		logBuffer := &bytes.Buffer{}
		resp.server.Logger = slog.New(slog.NewHandler(false, logBuffer))
		handler := &PanicStatusHandler{}
		handler.Init(resp.server)

		handler.Handle(resp, req)
		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, res.Body.Close())
		require.NoError(t, err)

		assert.Equal(t, `{"error":null}`+"\n", string(body))

		// Error and stacktrace are not printed to console because recovery middleware
		// is not executed (no error raised, we just set the response status to 500 for example)
		assert.Empty(t, logBuffer.String())
	})
}

func TestErrorStatusHandler(t *testing.T) {
	req, resp, recorder := prepareStatusHandlerTest()
	handler := &ErrorStatusHandler{}
	handler.Init(resp.server)

	resp.Status(http.StatusNotFound)

	handler.Handle(resp, req)

	res := recorder.Result()
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, res.Body.Close())
	require.NoError(t, err)

	assert.Equal(t, `{"error":"Not Found"}`+"\n", string(body))
}

func TestValidationStatusHandler(t *testing.T) {
	req, resp, recorder := prepareStatusHandlerTest()
	handler := &ValidationStatusHandler{}
	handler.Init(resp.server)

	req.Extra[ExtraValidationError{}] = &validation.Errors{
		Errors: []string{"The body is required"},
		Fields: validation.FieldsErrors{
			"field": &validation.Errors{Errors: []string{"The field is required"}},
		},
	}
	req.Extra[ExtraQueryValidationError{}] = &validation.Errors{
		Fields: validation.FieldsErrors{
			"query": &validation.Errors{Errors: []string{"The query is required"}},
		},
	}

	handler.Handle(resp, req)

	res := recorder.Result()
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, res.Body.Close())
	require.NoError(t, err)

	assert.Equal(t, `{"error":{"body":{"fields":{"field":{"errors":["The field is required"]}},"errors":["The body is required"]},"query":{"fields":{"query":{"errors":["The query is required"]}}}}}`+"\n", string(body))
}

func TestParseErrorStatusHandler(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectedMessage string
		expectedStatus  int
	}{
		{
			name:            "InvalidJSONBody",
			err:             ErrInvalidJSONBody,
			expectedMessage: "The request Content-Type indicates JSON, but the request body is empty or invalid.",
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:            "InvalidQuery",
			err:             ErrInvalidQuery,
			expectedMessage: "Failed to parse query string due to invalid syntax or unexpected input format.",
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:            "InvalidContentForType",
			err:             ErrInvalidContentForType,
			expectedMessage: "The request content does not match its type. E.g. invalid multipart/form-data or a problem with the file upload.",
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:            "ErrorInRequestBody",
			err:             ErrErrorInRequestBody,
			expectedMessage: "Failed to read request body due to connection issues, timeouts, size mismatches, or corrupted data.",
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name:            "OtherError",
			err:             errors.New("some.other.error"),
			expectedMessage: "some.other.error",
			expectedStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, resp, recorder := prepareStatusHandlerTest()

			handler := &ParseErrorStatusHandler{}
			handler.Init(resp.server)

			req.Extra[ExtraParseError{}] = tt.err
			resp.Status(tt.expectedStatus)

			handler.Handle(resp, req)

			res := recorder.Result()
			body, err := io.ReadAll(res.Body)

			assert.NoError(t, res.Body.Close())
			require.NoError(t, err)

			expectedResponse := fmt.Sprintf(`{"error":"%s"}`, tt.expectedMessage) + "\n"
			assert.Equal(t, expectedResponse, string(body))
			assert.Equal(t, tt.expectedStatus, res.StatusCode)
		})
	}
}

func TestParseErrorStatusHandlerWithoutExtra(t *testing.T) {
	req, resp, recorder := prepareStatusHandlerTest()

	handler := &ParseErrorStatusHandler{}
	handler.Init(resp.server)

	resp.Status(http.StatusBadRequest)

	handler.Handle(resp, req)

	res := recorder.Result()
	body, err := io.ReadAll(res.Body)

	assert.NoError(t, res.Body.Close())
	require.NoError(t, err)

	expectedResponse := fmt.Sprintf(`{"error":"%s"}`, "Bad Request") + "\n"
	assert.Equal(t, expectedResponse, string(body))
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
