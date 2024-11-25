package goyave

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/slog"
	errorutil "goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
)

func newTestReponse() (*Response, *httptest.ResponseRecorder) {
	server, err := New(Options{Config: config.LoadDefault()})
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
	flushErr   error
	prewritten []byte
	closed     bool
	flushed    bool
}

func (r *testChainedWriter) PreWrite(b []byte) {
	r.prewritten = b
}

func (r *testChainedWriter) Flush() error {
	r.flushed = true
	return r.flushErr
}

func (r *testChainedWriter) Close() error {
	r.closed = true
	return nil
}

type testChainedWriterHTTPFlusher struct {
	*testChainedWriter
}

func (r *testChainedWriterHTTPFlusher) Flush() {
	r.flushed = true
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
		require.NoError(t, res.Body.Close())
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("WriteHeader", func(t *testing.T) {
		resp, recorder := newTestReponse()
		resp.WriteHeader(http.StatusNoContent)

		res := recorder.Result()
		require.NoError(t, res.Body.Close())
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
		require.NoError(t, res.Body.Close())
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
		cases := []struct {
			setup           func(resp *Response)
			desc            string
			filename        string
			wantLength      string
			wantContentType string
			wantDisposition string
			wantBody        []byte
			wantStatus      int
		}{
			{
				desc:            "content-type_detection",
				filename:        "resources/test_file.txt",
				wantLength:      "25",
				wantContentType: "text/plain; charset=utf-8",
				wantDisposition: "inline",
				wantStatus:      http.StatusOK,
				wantBody:        append([]byte{0xef, 0xbb, 0xbf}, []byte("utf-8 with BOM content")...), // utf-8 BOM + text content
			},
			{
				desc: "content-type_provided",
				setup: func(resp *Response) {
					resp.Header().Set("Content-Type", "provided")
				},
				filename:        "resources/test_file.txt",
				wantLength:      "25",
				wantContentType: "provided",
				wantDisposition: "inline",
				wantStatus:      http.StatusOK,
				wantBody:        append([]byte{0xef, 0xbb, 0xbf}, []byte("utf-8 with BOM content")...), // utf-8 BOM + text content
			},
			{
				desc:            "empty_txt_file",
				filename:        "resources/empty.txt",
				wantLength:      "0",
				wantContentType: "text/plain",
				wantDisposition: "inline",
				wantStatus:      0, // Will be set to 200 in finalization step, which isn't executed here.
				wantBody:        []byte{},
			},
			{
				desc:            "empty_file",
				filename:        "resources/empty",
				wantLength:      "0",
				wantContentType: "application/octet-stream",
				wantDisposition: "inline",
				wantStatus:      0, // Will be set to 200 in finalization step, which isn't executed here.
				wantBody:        []byte{},
			},
			{
				desc:            "not_found",
				filename:        "not_a_file",
				wantLength:      "",
				wantContentType: "",
				wantDisposition: "",
				wantStatus:      http.StatusNotFound,
				wantBody:        []byte{},
			},
			{
				desc:            "directory",
				filename:        ".",
				wantLength:      "",
				wantContentType: "",
				wantDisposition: "",
				wantStatus:      http.StatusNotFound,
				wantBody:        []byte{},
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				resp, recorder := newTestReponse()
				if c.setup != nil {
					c.setup(resp)
				}
				resp.File(&osfs.FS{}, c.filename)
				res := recorder.Result()

				assert.Equal(t, c.wantStatus, resp.status)
				assert.Equal(t, c.wantDisposition, res.Header.Get("Content-Disposition"))
				assert.Equal(t, c.wantLength, res.Header.Get("Content-Length"))
				assert.Equal(t, c.wantContentType, res.Header.Get("Content-Type"))

				body, err := io.ReadAll(res.Body)
				assert.NoError(t, res.Body.Close())
				require.NoError(t, err)
				assert.Equal(t, c.wantBody, body)
			})
		}
	})

	t.Run("Download", func(t *testing.T) {
		cases := []struct {
			setup           func(resp *Response)
			desc            string
			path            string
			filename        string
			wantLength      string
			wantContentType string
			wantDisposition string
			wantBody        []byte
			wantStatus      int
		}{
			{
				desc:            "content-type_detection",
				path:            "resources/test_file.txt",
				filename:        "name.txt",
				wantLength:      "25",
				wantContentType: "text/plain; charset=utf-8",
				wantDisposition: "attachment; filename=\"name.txt\"",
				wantStatus:      http.StatusOK,
				wantBody:        append([]byte{0xef, 0xbb, 0xbf}, []byte("utf-8 with BOM content")...), // utf-8 BOM + text content
			},
			{
				desc: "content-type_provided",
				setup: func(resp *Response) {
					resp.Header().Set("Content-Type", "provided")
				},
				path:            "resources/test_file.txt",
				filename:        "name.txt",
				wantLength:      "25",
				wantContentType: "provided",
				wantDisposition: "attachment; filename=\"name.txt\"",
				wantStatus:      http.StatusOK,
				wantBody:        append([]byte{0xef, 0xbb, 0xbf}, []byte("utf-8 with BOM content")...), // utf-8 BOM + text content
			},
			{
				desc:            "empty_txt_file",
				path:            "resources/empty.txt",
				filename:        "name.txt",
				wantLength:      "0",
				wantContentType: "text/plain",
				wantDisposition: "attachment; filename=\"name.txt\"",
				wantStatus:      0, // Will be set to 200 in finalization step, which isn't executed here.
				wantBody:        []byte{},
			},
			{
				desc:            "empty_file",
				path:            "resources/empty",
				filename:        "empty",
				wantLength:      "0",
				wantContentType: "application/octet-stream",
				wantDisposition: "attachment; filename=\"empty\"",
				wantStatus:      0, // Will be set to 200 in finalization step, which isn't executed here.
				wantBody:        []byte{},
			},
			{
				desc:            "empty_json_file", // The content-type should be detected from the file extension as a fallback
				path:            "resources/empty.json",
				filename:        "name.json",
				wantLength:      "0",
				wantContentType: "application/json",
				wantDisposition: "attachment; filename=\"name.json\"",
				wantStatus:      0, // Will be set to 200 in finalization step, which isn't executed here.
				wantBody:        []byte{},
			},
			{
				desc:            "not_found",
				path:            "not_a_file",
				filename:        "name.txt",
				wantLength:      "",
				wantContentType: "",
				wantDisposition: "",
				wantStatus:      http.StatusNotFound,
				wantBody:        []byte{},
			},
			{
				desc:            "directory",
				path:            ".",
				filename:        "name.txt",
				wantLength:      "",
				wantContentType: "",
				wantDisposition: "",
				wantStatus:      http.StatusNotFound,
				wantBody:        []byte{},
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				resp, recorder := newTestReponse()
				if c.setup != nil {
					c.setup(resp)
				}
				resp.Download(&osfs.FS{}, c.path, c.filename)
				res := recorder.Result()

				assert.Equal(t, c.wantStatus, resp.status)
				assert.Equal(t, c.wantDisposition, res.Header.Get("Content-Disposition"))
				assert.Equal(t, c.wantLength, res.Header.Get("Content-Length"))
				assert.Equal(t, c.wantContentType, res.Header.Get("Content-Type"))

				body, err := io.ReadAll(res.Body)
				assert.NoError(t, res.Body.Close())
				require.NoError(t, err)
				assert.Equal(t, c.wantBody, body)
			})
		}
	})

	t.Run("JSON", func(t *testing.T) {
		resp, recorder := newTestReponse()
		resp.JSON(http.StatusOK, map[string]any{"hello": "world"})

		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, res.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"))
		assert.Equal(t, "{\"hello\":\"world\"}\n", string(body))
	})

	t.Run("JSON_error", func(t *testing.T) {
		resp, _ := newTestReponse()
		assert.Panics(t, func() {
			resp.JSON(http.StatusOK, make(chan struct{}))
		})
	})

	t.Run("String", func(t *testing.T) {
		resp, recorder := newTestReponse()
		resp.String(http.StatusOK, "hello world")

		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, res.Body.Close())
		require.NoError(t, err)
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
		require.NoError(t, res.Body.Close())
		cookies := res.Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "cookie-name", cookies[0].Name)
		assert.Equal(t, "test", cookies[0].Value)
	})

	t.Run("Write", func(t *testing.T) {
		resp, recorder := newTestReponse()
		_, _ = resp.Write([]byte("hello world"))

		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, res.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.status) // Ensures PreWrite has been called
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "hello world", string(body))
	})

	t.Run("PreWrite_called_once", func(t *testing.T) {
		resp, recorder := newTestReponse()
		newWriter := &testChainedWriter{
			ResponseRecorder: recorder,
		}
		resp.SetWriter(newWriter)
		_, _ = resp.Write([]byte("hello world\n"))
		_, _ = resp.Write([]byte("how are you today?"))

		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, res.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.status)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, "hello world\nhow are you today?", string(body))
		assert.Equal(t, "hello world\n", string(newWriter.prewritten))
	})

	t.Run("Hijack", func(t *testing.T) {
		resp, _ := newTestReponse()
		resp.responseWriter = &hijackableRecorder{httptest.NewRecorder()}

		assert.False(t, resp.hijacked)
		assert.False(t, resp.Hijacked())

		c, b, err := resp.Hijack()
		require.NoError(t, err)
		assert.NotNil(t, c)
		assert.NotNil(t, b)
		assert.True(t, resp.hijacked)
		assert.True(t, resp.Hijacked())

		t.Run("not_hijackable", func(t *testing.T) {
			resp, _ := newTestReponse()

			c, b, err := resp.Hijack()
			require.ErrorIs(t, err, ErrNotHijackable)
			assert.Nil(t, c)
			assert.Nil(t, b)
		})

		t.Run("error_on_hijack", func(t *testing.T) {
			resp, _ := newTestReponse()
			resp.server.config.Set("app.debug", true)
			resp.server.Logger = slog.New(slog.NewHandler(false, &bytes.Buffer{}))
			recorder := httptest.NewRecorder()
			resp.responseWriter = &hijackableRecorder{recorder}

			_, _, err := resp.Hijack()
			require.NoError(t, err)

			resp.Error(fmt.Errorf("test error"))
			res := recorder.Result()
			defer func() {
				assert.NoError(t, res.Body.Close())
			}()
			assert.Equal(t, http.StatusInternalServerError, resp.status)

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			// The connection was hijacked so errors shouldn't be written to the response
			assert.Empty(t, body)
		})
	})

	t.Run("SetWriter", func(t *testing.T) {
		resp, _ := newTestReponse()
		newWriter := &bytes.Buffer{}
		resp.SetWriter(newWriter)
		assert.Equal(t, newWriter, resp.Writer())
	})

	t.Run("SetWriter_composable", func(t *testing.T) {
		type composableWriter struct {
			Component
			bytes.Buffer
		}

		resp, _ := newTestReponse()
		newWriter := &composableWriter{}
		resp.SetWriter(newWriter)
		assert.Equal(t, newWriter, resp.Writer())
		assert.Equal(t, resp.server, newWriter.server)
	})

	t.Run("Chained_writer", func(t *testing.T) {
		resp, _ := newTestReponse()
		newWriter := &testChainedWriter{}
		resp.SetWriter(newWriter)

		resp.PreWrite([]byte{1, 2, 3})
		assert.Equal(t, []byte{1, 2, 3}, newWriter.prewritten)
		require.NoError(t, resp.close())
		assert.True(t, newWriter.closed)
	})

	t.Run("Error_no_debug", func(t *testing.T) {
		resp, _ := newTestReponse()
		logBuffer := &bytes.Buffer{}
		resp.server.Logger = slog.New(slog.NewHandler(false, logBuffer))
		resp.server.config.Set("app.debug", false)
		err := fmt.Errorf("custom error")
		resp.Error(err)

		e := resp.GetError()
		if !assert.NotNil(t, e) {
			return
		}
		assert.Equal(t, []error{err}, e.Unwrap())
		assert.Equal(t, http.StatusInternalServerError, resp.status)
		assert.Regexp(t,
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"%s","trace":%s}\n`,
				regexp.QuoteMeta(e.Error()), regexp.QuoteMeta(string(lo.Must(json.Marshal(e.StackFrames().String())))),
			),
			logBuffer.String(),
		)
	})

	t.Run("Error_no_debug_nil", func(t *testing.T) {
		resp, _ := newTestReponse()
		logBuffer := &bytes.Buffer{}
		resp.server.Logger = slog.New(slog.NewHandler(false, logBuffer))
		resp.server.config.Set("app.debug", false)
		resp.Error(nil)

		e := resp.GetError()
		assert.Nil(t, e)
		assert.Equal(t, http.StatusInternalServerError, resp.status)
	})

	t.Run("Error_with_debug", func(t *testing.T) {
		cases := []struct {
			expectedLog     func(e *errorutil.Error) *regexp.Regexp
			err             any
			expectedMessage string
		}{
			{err: fmt.Errorf("custom error"), expectedMessage: `"custom error"`, expectedLog: func(e *errorutil.Error) *regexp.Regexp {
				return regexp.MustCompile(
					fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"%s","trace":%s}\n`,
						regexp.QuoteMeta(e.Error()), regexp.QuoteMeta(string(lo.Must(json.Marshal(e.StackFrames().String())))),
					))
			}},
			{err: map[string]any{"key": "value"}, expectedMessage: `{"key":"value"}`, expectedLog: func(e *errorutil.Error) *regexp.Regexp {
				return regexp.MustCompile(
					fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"%s","trace":%s,"reason":{"key":"value"}}\n`,
						regexp.QuoteMeta(e.Error()), regexp.QuoteMeta(string(lo.Must(json.Marshal(e.StackFrames().String())))),
					))
			}},
			{err: []error{fmt.Errorf("custom error 1"), fmt.Errorf("custom error 2")}, expectedMessage: `["custom error 1","custom error 2"]`, expectedLog: func(e *errorutil.Error) *regexp.Regexp {
				reasons := e.Unwrap()
				stacktrace := regexp.QuoteMeta(string(lo.Must(json.Marshal(e.StackFrames().String()))))
				return regexp.MustCompile(
					fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"%s","trace":%s}\n{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"%s","trace":%s}\n`,
						regexp.QuoteMeta(reasons[0].Error()), stacktrace,
						regexp.QuoteMeta(reasons[1].Error()), stacktrace,
					),
				)
			}},
			{err: nil, expectedMessage: `null`, expectedLog: func(_ *errorutil.Error) *regexp.Regexp {
				return regexp.MustCompile(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"<nil>"}\n`)
			}},
		}

		for _, c := range cases {
			resp, recorder := newTestReponse()
			logBuffer := &bytes.Buffer{}
			resp.server.Logger = slog.New(slog.NewHandler(false, logBuffer))
			resp.server.config.Set("app.debug", true)
			resp.Error(c.err)

			e := resp.GetError()
			switch c.err {
			case nil:
				if !assert.Nil(t, e) {
					return
				}
			default:
				if !assert.NotNil(t, e) {
					return
				}
			}
			assert.Equal(t, http.StatusInternalServerError, resp.status)

			res := recorder.Result()
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, res.Body.Close())
			require.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, resp.status)
			assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
			assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"))
			assert.Equal(t, "{\"error\":"+c.expectedMessage+"}\n", string(body))

			assert.Regexp(t, c.expectedLog(e), logBuffer.String())
		}
	})

	t.Run("Error_with_debug_and_custom_status", func(t *testing.T) {
		resp, recorder := newTestReponse()
		logBuffer := &bytes.Buffer{}
		resp.server.Logger = slog.New(slog.NewHandler(false, logBuffer))
		resp.server.config.Set("app.debug", true)
		err := fmt.Errorf("custom error")
		resp.Status(http.StatusForbidden)
		resp.Error(err)

		e := resp.GetError()
		if !assert.NotNil(t, e) {
			return
		}
		assert.Equal(t, []error{err}, e.Unwrap())
		assert.Equal(t, http.StatusForbidden, resp.status)

		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, res.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.status)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
		assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"))
		assert.Equal(t, "{\"error\":\"custom error\"}\n", string(body))

		assert.Regexp(t,
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"%s","trace":%s}\n`,
				regexp.QuoteMeta(e.Error()), regexp.QuoteMeta(string(lo.Must(json.Marshal(e.StackFrames().String())))),
			),
			logBuffer.String(),
		)
	})

	t.Run("Error_with_debug_and_not_empty", func(t *testing.T) {
		resp, recorder := newTestReponse()
		logBuffer := &bytes.Buffer{}
		resp.server.Logger = slog.New(slog.NewHandler(false, logBuffer))
		resp.server.config.Set("app.debug", true)
		err := fmt.Errorf("custom error")
		resp.String(http.StatusForbidden, "forbidden")
		resp.Error(err)

		e := resp.GetError()
		if !assert.NotNil(t, e) {
			return
		}
		assert.Equal(t, []error{err}, e.Unwrap())
		assert.Equal(t, http.StatusForbidden, resp.status)

		res := recorder.Result()
		body, err := io.ReadAll(res.Body)
		assert.NoError(t, res.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.status)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
		assert.Equal(t, "forbidden", string(body))

		assert.Regexp(t,
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"%s","trace":%s}\n`,
				regexp.QuoteMeta(e.Error()), regexp.QuoteMeta(string(lo.Must(json.Marshal(e.StackFrames().String())))),
			),
			logBuffer.String(),
		)
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
			resp.server.Logger = slog.New(slog.NewHandler(false, logBuffer))
			assert.True(t, resp.WriteDBError(fmt.Errorf("random db error")))

			res := recorder.Result()
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, res.Body.Close())
			require.NoError(t, err)
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

	t.Run("Flush", func(t *testing.T) {
		resp, recorder := newTestReponse()
		newWriter := &testChainedWriter{
			ResponseRecorder: recorder,
		}
		resp.SetWriter(newWriter)
		resp.Flush()
		assert.True(t, newWriter.flushed)
		assert.True(t, resp.wroteHeader)
		assert.Equal(t, []byte{}, newWriter.prewritten) // PreWrite called

		// PreWrite called once
		newWriter.prewritten = nil
		resp.Flush()
		assert.Nil(t, newWriter.prewritten)

		t.Run("error", func(t *testing.T) {
			resp, recorder := newTestReponse()
			logBuffer := &bytes.Buffer{}
			resp.server.Logger = slog.New(slog.NewHandler(false, logBuffer))
			newWriter := &testChainedWriter{
				ResponseRecorder: recorder,
				flushErr:         fmt.Errorf("test error"),
			}
			resp.SetWriter(newWriter)
			resp.Flush()
			assert.NotEmpty(t, logBuffer.String())
		})

		t.Run("http.Flusher", func(t *testing.T) {
			resp, recorder := newTestReponse()
			logBuffer := &bytes.Buffer{}
			resp.server.Logger = slog.New(slog.NewHandler(false, logBuffer))
			newWriter := &testChainedWriterHTTPFlusher{
				testChainedWriter: &testChainedWriter{
					ResponseRecorder: recorder,
				},
			}
			resp.SetWriter(newWriter)
			resp.Flush()
			assert.True(t, newWriter.flushed)
			assert.Empty(t, logBuffer.String())
		})
	})
}

type httpFlusher struct {
	io.Writer
	flushed bool
}

func (f *httpFlusher) Flush() {
	f.flushed = true
}

func TestCommonWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	chainedWriter := &testChainedWriter{
		ResponseRecorder: recorder,
	}
	wr := NewCommonWriter(chainedWriter)
	assert.Equal(t, chainedWriter, wr.wr)

	wr.PreWrite([]byte("hello"))
	assert.Equal(t, []byte("hello"), chainedWriter.prewritten)

	_, err := wr.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.NoError(t, wr.Flush())
	assert.True(t, chainedWriter.flushed)
	assert.NoError(t, wr.Close())
	assert.True(t, chainedWriter.closed)

	res := chainedWriter.Result()
	body, err := io.ReadAll(res.Body)
	assert.NoError(t, res.Body.Close())
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), body)

	t.Run("http.flusher_and_nocloser", func(t *testing.T) {
		chainedWriter := &httpFlusher{}
		wr := NewCommonWriter(chainedWriter)
		assert.NoError(t, wr.Flush())
		assert.True(t, chainedWriter.flushed)
		assert.NoError(t, wr.Close())
	})
}
