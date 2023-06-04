package log

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/util/testutil"
)

type testWriter struct {
	io.Writer
	closed     bool
	preWritten bool
}

func (w *testWriter) PreWrite(b []byte) {
	w.preWritten = true
	if pr, ok := w.Writer.(goyave.PreWriter); ok {
		pr.PreWrite(b)
	}
}

func (w *testWriter) Close() error {
	w.closed = true
	return nil
}

func TestWriter(t *testing.T) {

	t.Run("Write", func(t *testing.T) {
		ts := lo.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))
		server := testutil.NewTestServerWithConfig(t, config.LoadDefault(), nil)
		req := server.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		resp, recorder := server.NewTestResponse(req)

		buffer := bytes.NewBufferString("")
		server.AccessLogger.SetOutput(buffer)

		writer := NewWriter(server.Server, resp, req, CommonLogFormatter)
		resp.SetWriter(writer)

		i, err := resp.Write([]byte("body response"))
		assert.Equal(t, 13, i)
		assert.Equal(t, 13, writer.length)
		assert.NoError(t, err)

		assert.NoError(t, writer.Close())
		httpResponse := recorder.Result()
		_ = httpResponse.Body.Close()
		assert.Equal(t, http.StatusOK, httpResponse.StatusCode)
		assert.Equal(t, "192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 200 13\n", buffer.String())
	})

	t.Run("child_writer_prewrite_and_close", func(t *testing.T) {
		ts := lo.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))
		server := testutil.NewTestServerWithConfig(t, config.LoadDefault(), nil)
		req := server.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		resp, recorder := server.NewTestResponse(req)

		buffer := bytes.NewBufferString("")
		server.AccessLogger.SetOutput(buffer)

		child := &testWriter{
			preWritten: false,
			closed:     false,
			Writer:     resp.Writer(),
		}
		resp.SetWriter(child)
		writer := NewWriter(server.Server, resp, req, CommonLogFormatter)
		resp.SetWriter(writer)

		i, err := resp.Write([]byte("body response"))
		assert.True(t, child.preWritten)
		assert.Equal(t, 13, i)
		assert.Equal(t, 13, writer.length)
		assert.NoError(t, err)

		assert.NoError(t, writer.Close())
		httpResponse := recorder.Result()
		_ = httpResponse.Body.Close()
		assert.Equal(t, http.StatusOK, httpResponse.StatusCode)
		assert.Equal(t, "192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 200 13\n", buffer.String())
		assert.True(t, child.closed)
	})
}

func TestMiddleware(t *testing.T) {

	t.Run("Common", func(t *testing.T) {
		ts := lo.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))
		server := testutil.NewTestServerWithConfig(t, config.LoadDefault(), nil)
		buffer := bytes.NewBufferString("")
		server.AccessLogger.SetOutput(buffer)

		req := server.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		httpResponse := server.TestMiddleware(CommonLogMiddleware(), req, func(r *goyave.ResponseV5, _ *goyave.RequestV5) {
			r.String(http.StatusOK, "hello world")
		})
		_ = httpResponse.Body.Close()
		assert.Equal(t, http.StatusOK, httpResponse.StatusCode)
		assert.Equal(t, "192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 200 11\n", buffer.String())
	})

	t.Run("Combined", func(t *testing.T) {
		ts := lo.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))
		server := testutil.NewTestServerWithConfig(t, config.LoadDefault(), nil)
		buffer := bytes.NewBufferString("")
		server.AccessLogger.SetOutput(buffer)

		req := server.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts

		referrer := "http://example.com"
		userAgent := "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:74.0) Gecko/20100101 Firefox/74.0"
		req.Header().Set("Referer", referrer)
		req.Header().Set("User-Agent", userAgent)

		httpResponse := server.TestMiddleware(CombinedLogMiddleware(), req, func(r *goyave.ResponseV5, _ *goyave.RequestV5) {
			r.String(http.StatusOK, "hello world")
		})
		_ = httpResponse.Body.Close()
		assert.Equal(t, http.StatusOK, httpResponse.StatusCode)
		assert.Equal(t, "192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 200 11 \""+referrer+"\" \""+userAgent+"\"\n", buffer.String())
	})
}
