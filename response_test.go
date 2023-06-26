package goyave

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"goyave.dev/goyave/v4/config"
)

func newTestReponse() (*ResponseV5, *httptest.ResponseRecorder) {
	server, err := NewWithConfig(config.LoadDefault())
	if err != nil {
		panic(err)
	}
	httpReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	req := NewRequest(httpReq)
	recorder := httptest.NewRecorder()
	resp := NewResponse(server, req, recorder)
	return resp, recorder
}

type hijackableRecorder struct {
	*httptest.ResponseRecorder
}

func (r *hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn := &net.TCPConn{}
	return conn, bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

type testChainedWriter struct {
	*httptest.ResponseRecorder
	closed     bool
	prewritten []byte
}

func (r *testChainedWriter) PreWrite(b []byte) {
	r.prewritten = b
}

func (r *testChainedWriter) Close() error {
	r.closed = true
	return nil
}

func TestResponse(t *testing.T) {

	t.Run("NewResponse", func(t *testing.T) {
		resp, _ := newTestReponse()
		assert.NotNil(t, resp.server)
		assert.NotNil(t, resp.request)
		assert.NotNil(t, resp.writer)
		assert.NotNil(t, resp.responseWriter)
		assert.Equal(t, resp.writer, resp.responseWriter)
		assert.True(t, resp.empty)
		assert.Equal(t, 0, resp.status)
		assert.False(t, resp.wroteHeader)
	})

	t.Run("Status", func(t *testing.T) {
		// The status header should not be written right away when
		// defining the status.
		resp, recorder := newTestReponse()
		resp.Status(http.StatusForbidden)
		assert.Equal(t, http.StatusForbidden, resp.status)
		assert.Equal(t, http.StatusForbidden, resp.GetStatus())
		assert.False(t, resp.wroteHeader)

		// Can't override status once defined
		resp.Status(http.StatusOK)
		assert.Equal(t, http.StatusForbidden, resp.status)

		// Header not written
		res := recorder.Result()
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("WriteHeader", func(t *testing.T) {
		resp, recorder := newTestReponse()
		resp.WriteHeader(http.StatusNoContent)

		res := recorder.Result()
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusNoContent, resp.status)
		assert.True(t, resp.wroteHeader)
		assert.True(t, resp.IsHeaderWritten())
		assert.Equal(t, http.StatusNoContent, res.StatusCode)

		// Cannot rewrite header
		resp.WriteHeader(http.StatusForbidden)
		assert.Equal(t, http.StatusNoContent, resp.status)
	})

	t.Run("Header", func(t *testing.T) {
		resp, recorder := newTestReponse()
		resp.Header().Set("X-Test", "value")
		resp.WriteHeader(http.StatusOK)

		res := recorder.Result()
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "value", res.Header.Get("X-Test"))
	})

	t.Run("IsEmpty", func(t *testing.T) {
		resp, _ := newTestReponse()
		resp.Status(http.StatusOK)
		assert.True(t, resp.IsEmpty())
		resp.WriteHeader(http.StatusOK)
		assert.True(t, resp.IsEmpty())
		_, _ = resp.Write([]byte("hello"))
		assert.False(t, resp.IsEmpty())
	})

	t.Run("File", func(t *testing.T) {
		resp, recorder := newTestReponse()

		resp.File("resources/test_file.txt")
		res := recorder.Result()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "inline", res.Header.Get("Content-Disposition"))
		assert.Equal(t, "25", res.Header.Get("Content-Length"))
		assert.Equal(t, "text/plain; charset=utf-8", res.Header.Get("Content-Type"))

		body, err := io.ReadAll(res.Body)
		assert.NoError(t, res.Body.Close())
		if !assert.NoError(t, err) {
			return
		}

		// utf-8 BOM + text content
		assert.Equal(t, append([]byte{0xef, 0xbb, 0xbf}, []byte("utf-8 with BOM content")...), body)

		t.Run("not_found", func(t *testing.T) {
			resp, _ := newTestReponse()
			resp.File("not_a_file")
			assert.Equal(t, http.StatusNotFound, resp.status)
		})
	})

	t.Run("Download", func(t *testing.T) {
		resp, recorder := newTestReponse()

		resp.Download("resources/test_file.txt", "test_file.txt")
		res := recorder.Result()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, fmt.Sprintf("attachment; filename=\"test_file.txt\""), res.Header.Get("Content-Disposition"))
		assert.Equal(t, "25", res.Header.Get("Content-Length"))
		assert.Equal(t, "text/plain; charset=utf-8", res.Header.Get("Content-Type"))

		body, err := io.ReadAll(res.Body)
		assert.NoError(t, res.Body.Close())
		if !assert.NoError(t, err) {
			return
		}

		// utf-8 BOM + text content
		assert.Equal(t, append([]byte{0xef, 0xbb, 0xbf}, []byte("utf-8 with BOM content")...), body)

		t.Run("not_found", func(t *testing.T) {
			resp, _ := newTestReponse()
			resp.Download("not_a_file", "file.txt")
			assert.Equal(t, http.StatusNotFound, resp.status)
		})
	})

	t.Run("JSON", func(t *testing.T) {
		resp, recorder := newTestReponse()
		resp.JSON(http.StatusOK, map[string]any{"hello": "world"})

		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"))
		assert.Equal(t, "{\"hello\":\"world\"}\n", string(body))
	})

	t.Run("String", func(t *testing.T) {
		resp, recorder := newTestReponse()
		resp.String(http.StatusOK, "hello world")

		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "hello world", string(body))
	})

	t.Run("Cookie", func(t *testing.T) {
		resp, recorder := newTestReponse()
		resp.Cookie(&http.Cookie{
			Name:  "cookie-name",
			Value: "test",
		})

		res := recorder.Result()
		assert.NoError(t, res.Body.Close())
		cookies := res.Cookies()
		if !assert.Equal(t, 1, len(cookies)) {
			return
		}
		assert.Equal(t, "cookie-name", cookies[0].Name)
		assert.Equal(t, "test", cookies[0].Value)
	})

	t.Run("Write", func(t *testing.T) {
		resp, recorder := newTestReponse()
		_, _ = resp.Write([]byte("hello world"))

		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusOK, resp.status) // Ensures PreWrite has been called
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "hello world", string(body))
	})

	t.Run("Hijack", func(t *testing.T) {
		resp, _ := newTestReponse()
		resp.responseWriter = &hijackableRecorder{httptest.NewRecorder()}

		assert.False(t, resp.hijacked)
		assert.False(t, resp.Hijacked())

		c, b, err := resp.Hijack()
		if !assert.NoError(t, err) {
			return
		}
		assert.Nil(t, err)
		assert.NotNil(t, c)
		assert.NotNil(t, b)
		assert.True(t, resp.hijacked)
		assert.True(t, resp.Hijacked())

		t.Run("not_hijackable", func(t *testing.T) {
			resp, _ := newTestReponse()

			c, b, err := resp.Hijack()
			assert.True(t, errors.Is(err, ErrNotHijackable))
			assert.Nil(t, c)
			assert.Nil(t, b)
		})

		t.Run("error_on_hijack", func(t *testing.T) {
			resp, _ := newTestReponse()
			resp.server.config.Set("app.debug", true)
			resp.server.ErrLogger = log.New(&bytes.Buffer{}, "", 0)
			recorder := httptest.NewRecorder()
			resp.responseWriter = &hijackableRecorder{recorder}

			_, _, err := resp.Hijack()
			if !assert.NoError(t, err) {
				return
			}

			resp.Error(fmt.Errorf("test error"))
			res := recorder.Result()
			assert.Equal(t, http.StatusInternalServerError, resp.status)

			body, err := io.ReadAll(res.Body)
			assert.Nil(t, err)

			// The connection was hijacked so errors shouldn't be written to the response
			assert.Empty(t, body)
			assert.NoError(t, res.Body.Close())
		})
	})

	t.Run("SetWriter", func(t *testing.T) {
		resp, _ := newTestReponse()
		newWriter := &bytes.Buffer{}
		resp.SetWriter(newWriter)
		assert.Equal(t, newWriter, resp.Writer())
	})

	t.Run("Chained_writer", func(t *testing.T) {
		resp, _ := newTestReponse()
		newWriter := &testChainedWriter{}
		resp.SetWriter(newWriter)

		resp.PreWrite([]byte{1, 2, 3})
		assert.Equal(t, []byte{1, 2, 3}, newWriter.prewritten)
		assert.NoError(t, resp.close())
		assert.True(t, newWriter.closed)
	})

	t.Run("Error_no_debug", func(t *testing.T) {
		resp, _ := newTestReponse()
		logBuffer := &bytes.Buffer{}
		resp.server.ErrLogger = log.New(logBuffer, "", 0)
		resp.server.config.Set("app.debug", false)
		err := fmt.Errorf("custom error")
		resp.Error(err)

		assert.Equal(t, err, resp.request.Extra[ExtraError])
		assert.Equal(t, http.StatusInternalServerError, resp.status)
		assert.Equal(t, "custom error\n", logBuffer.String())
	})

	t.Run("Error_with_debug", func(t *testing.T) {
		cases := []struct {
			err             any
			expectedMessage string
			expectedLog     string
		}{
			{err: fmt.Errorf("custom error"), expectedMessage: `"custom error"`, expectedLog: "custom error"},
			{err: map[string]any{"key": "value"}, expectedMessage: `{"key":"value"}`, expectedLog: "map[key:value]"},
			{err: []error{fmt.Errorf("custom error 1"), fmt.Errorf("custom error 2")}, expectedMessage: `["custom error 1","custom error 2"]`, expectedLog: `[custom error 1 custom error 2]`},
		}

		for _, c := range cases {
			resp, recorder := newTestReponse()
			logBuffer := &bytes.Buffer{}
			resp.server.ErrLogger = log.New(logBuffer, "", 0)
			resp.server.config.Set("app.debug", true)
			resp.Error(c.err)

			assert.Equal(t, c.err, resp.request.Extra[ExtraError])
			assert.Equal(t, http.StatusInternalServerError, resp.status)

			res := recorder.Result()
			body, err := io.ReadAll(res.Body)
			if !assert.NoError(t, err) {
				return
			}
			assert.NoError(t, res.Body.Close())
			assert.Equal(t, http.StatusInternalServerError, resp.status)
			assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
			assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"))
			assert.Equal(t, "{\"error\":"+c.expectedMessage+"}\n", string(body))
			assert.True(t, strings.HasPrefix(logBuffer.String(), c.expectedLog+"\ngoroutine ")) // Error and stacktrace printed to ErrLogger
		}
	})

	t.Run("Error_with_debug_and_custom_status", func(t *testing.T) {
		resp, recorder := newTestReponse()
		logBuffer := &bytes.Buffer{}
		resp.server.ErrLogger = log.New(logBuffer, "", 0)
		resp.server.config.Set("app.debug", true)
		err := fmt.Errorf("custom error")
		resp.Status(http.StatusForbidden)
		resp.Error(err)

		assert.Equal(t, err, resp.request.Extra[ExtraError])
		assert.Equal(t, http.StatusForbidden, resp.status)

		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusForbidden, resp.status)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
		assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"))
		assert.Equal(t, "{\"error\":\"custom error\"}\n", string(body))

		assert.True(t, strings.HasPrefix(logBuffer.String(), "custom error\ngoroutine ")) // Error and stacktrace printed to ErrLogger
	})

	t.Run("Error_with_debug_and_not_empty", func(t *testing.T) {
		resp, recorder := newTestReponse()
		logBuffer := &bytes.Buffer{}
		resp.server.ErrLogger = log.New(logBuffer, "", 0)
		resp.server.config.Set("app.debug", true)
		err := fmt.Errorf("custom error")
		resp.String(http.StatusForbidden, "forbidden")
		resp.Error(err)

		assert.Equal(t, err, resp.request.Extra[ExtraError])
		assert.Equal(t, http.StatusForbidden, resp.status)

		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		if !assert.NoError(t, err) {
			return
		}
		assert.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusForbidden, resp.status)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
		assert.Equal(t, "forbidden", string(body))

		assert.True(t, strings.HasPrefix(logBuffer.String(), "custom error\ngoroutine ")) // Error and stacktrace printed to ErrLogger
	})

	t.Run("WriteDBError", func(t *testing.T) {

		t.Run("ErrRecordNotFound", func(t *testing.T) {
			resp, _ := newTestReponse()
			assert.True(t, resp.WriteDBError(fmt.Errorf("%w", gorm.ErrRecordNotFound)))
			assert.Equal(t, http.StatusNotFound, resp.status)
		})

		t.Run("DBError", func(t *testing.T) {
			resp, recorder := newTestReponse()
			logBuffer := &bytes.Buffer{}
			resp.server.ErrLogger = log.New(logBuffer, "", 0)
			assert.True(t, resp.WriteDBError(fmt.Errorf("random db error")))

			res := recorder.Result()
			body, err := io.ReadAll(res.Body)
			if !assert.NoError(t, err) {
				return
			}
			assert.NoError(t, res.Body.Close())
			assert.Equal(t, http.StatusInternalServerError, resp.status)
			assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
			assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"))
			assert.Equal(t, "{\"error\":\"random db error\"}\n", string(body))
		})

		t.Run("no_error", func(t *testing.T) {
			resp, _ := newTestReponse()
			assert.False(t, resp.WriteDBError(nil))
		})

	})

}
