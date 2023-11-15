package log

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/testutil"
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
		cfg := config.LoadDefault()
		cfg.Set("app.debug", false)
		buffer := bytes.NewBufferString("")
		server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg, Logger: slog.New(slog.NewHandler(false, buffer))})
		req := server.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		resp, recorder := server.NewTestResponse(req)

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

		assert.Regexp(t, regexp.MustCompile(
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"INFO","source":{"function":".+","file":".+","line":\d+},"msg":"%s","details":{"host":"192\.0\.2\.1","username":"-","time":"2020-03-23T13:58:26\.371Z","method":"GET","uri":"/log","proto":"HTTP/1\.1","status":200,"length":13}}\n`,
				regexp.QuoteMeta(`192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 200 13`),
			)),
			buffer.String(),
		)
	})

	t.Run("Write_dev_mode", func(t *testing.T) {
		ts := lo.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))
		cfg := config.LoadDefault()
		cfg.Set("app.debug", false)
		buffer := bytes.NewBufferString("")
		server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg, Logger: slog.New(slog.NewHandler(false, buffer))})
		req := server.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		resp, recorder := server.NewTestResponse(req)

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

		assert.Regexp(t, regexp.MustCompile(
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"INFO","source":{"function":".+","file":".+","line":\d+},"msg":"%s","details":{"host":"192\.0\.2\.1","username":"-","time":"2020-03-23T13:58:26\.371Z","method":"GET","uri":"/log","proto":"HTTP/1\.1","status":200,"length":13}}\n`,
				regexp.QuoteMeta(`192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 200 13`),
			)),
			buffer.String(),
		)
	})

	t.Run("child_writer_prewrite_and_close", func(t *testing.T) {
		ts := lo.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))
		cfg := config.LoadDefault()
		cfg.Set("app.debug", false)
		buffer := bytes.NewBufferString("")
		server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg, Logger: slog.New(slog.NewHandler(false, buffer))})
		req := server.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		resp, recorder := server.NewTestResponse(req)

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

		assert.Regexp(t, regexp.MustCompile(
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"INFO","source":{"function":".+","file":".+","line":\d+},"msg":"%s","details":{"host":"192\.0\.2\.1","username":"-","time":"2020-03-23T13:58:26\.371Z","method":"GET","uri":"/log","proto":"HTTP/1\.1","status":200,"length":13}}\n`,
				regexp.QuoteMeta(`192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 200 13`),
			)),
			buffer.String(),
		)
		assert.True(t, child.closed)
	})

	t.Run("child_writer_prewrite_and_close_dev_mode", func(t *testing.T) {
		ts := lo.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))
		cfg := config.LoadDefault()
		cfg.Set("app.debug", true)
		buffer := bytes.NewBufferString("")
		server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg, Logger: slog.New(slog.NewHandler(false, buffer))})
		req := server.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		resp, recorder := server.NewTestResponse(req)

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

		assert.Regexp(t, regexp.MustCompile(
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"INFO","source":{"function":".+","file":".+","line":\d+},"msg":"%s"}\n`, // Same thing but details are omitted
				regexp.QuoteMeta(`192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 200 13`),
			)),
			buffer.String(),
		)
		assert.True(t, child.closed)
	})
}

func TestMiddleware(t *testing.T) {

	t.Run("Common", func(t *testing.T) {
		ts := lo.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))
		cfg := config.LoadDefault()
		cfg.Set("app.debug", false)
		buffer := bytes.NewBufferString("")
		server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg, Logger: slog.New(slog.NewHandler(false, buffer))})

		req := server.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		httpResponse := server.TestMiddleware(CommonLogMiddleware(), req, func(r *goyave.Response, _ *goyave.Request) {
			r.String(http.StatusOK, "hello world")
		})
		_ = httpResponse.Body.Close()
		assert.Equal(t, http.StatusOK, httpResponse.StatusCode)
		assert.Regexp(t, regexp.MustCompile(
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"INFO","source":{"function":".+","file":".+","line":\d+},"msg":"%s","details":{"host":"192\.0\.2\.1","username":"-","time":"2020-03-23T13:58:26\.371Z","method":"GET","uri":"/log","proto":"HTTP/1\.1","status":200,"length":11}}\n`,
				regexp.QuoteMeta(`192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 200 11`),
			)),
			buffer.String(),
		)
	})

	t.Run("Common_dev_mode", func(t *testing.T) {
		ts := lo.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))
		cfg := config.LoadDefault()
		cfg.Set("app.debug", true)
		buffer := bytes.NewBufferString("")
		server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg, Logger: slog.New(slog.NewHandler(false, buffer))})

		req := server.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		httpResponse := server.TestMiddleware(CommonLogMiddleware(), req, func(r *goyave.Response, _ *goyave.Request) {
			r.String(http.StatusOK, "hello world")
		})
		_ = httpResponse.Body.Close()
		assert.Equal(t, http.StatusOK, httpResponse.StatusCode)
		assert.Regexp(t, regexp.MustCompile(
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"INFO","source":{"function":".+","file":".+","line":\d+},"msg":"%s"}\n`, // Same thing but details are omitted
				regexp.QuoteMeta(`192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 200 11`),
			)),
			buffer.String(),
		)
	})

	t.Run("Combined", func(t *testing.T) {
		ts := lo.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))
		cfg := config.LoadDefault()
		cfg.Set("app.debug", false)
		server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg})
		buffer := bytes.NewBufferString("")
		server.Logger = slog.New(slog.NewHandler(false, buffer))

		req := server.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts

		referrer := "http://example.com"
		userAgent := "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:74.0) Gecko/20100101 Firefox/74.0"
		req.Header().Set("Referer", referrer)
		req.Header().Set("User-Agent", userAgent)

		httpResponse := server.TestMiddleware(CombinedLogMiddleware(), req, func(r *goyave.Response, _ *goyave.Request) {
			r.String(http.StatusOK, "hello world")
		})
		_ = httpResponse.Body.Close()
		assert.Equal(t, http.StatusOK, httpResponse.StatusCode)

		assert.Regexp(t, regexp.MustCompile(
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"INFO","source":{"function":".+","file":".+","line":\d+},"msg":"%s","details":{"host":"192\.0\.2\.1","username":"-","time":"2020-03-23T13:58:26\.371Z","method":"GET","uri":"/log","proto":"HTTP/1\.1","status":200,"length":11,"referrer":"%s","userAgent":"%s"}}\n`,
				regexp.QuoteMeta(fmt.Sprintf(`192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 200 11 \"%s\" \"%s\"`, referrer, userAgent)),
				regexp.QuoteMeta(referrer),
				regexp.QuoteMeta(userAgent),
			)),
			buffer.String(),
		)
	})

	t.Run("Combined_dev_mode", func(t *testing.T) {
		ts := lo.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))
		cfg := config.LoadDefault()
		cfg.Set("app.debug", true)
		buffer := bytes.NewBufferString("")
		server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: cfg, Logger: slog.New(slog.NewHandler(false, buffer))})

		req := server.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts

		referrer := "http://example.com"
		userAgent := "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:74.0) Gecko/20100101 Firefox/74.0"
		req.Header().Set("Referer", referrer)
		req.Header().Set("User-Agent", userAgent)

		httpResponse := server.TestMiddleware(CombinedLogMiddleware(), req, func(r *goyave.Response, _ *goyave.Request) {
			r.String(http.StatusOK, "hello world")
		})
		_ = httpResponse.Body.Close()
		assert.Equal(t, http.StatusOK, httpResponse.StatusCode)

		assert.Regexp(t, regexp.MustCompile(
			fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"INFO","source":{"function":".+","file":".+","line":\d+},"msg":"%s"}\n`, // Same thing but details are omitted
				regexp.QuoteMeta(fmt.Sprintf(`192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 200 11 \"%s\" \"%s\"`, referrer, userAgent)),
			)),
			buffer.String(),
		)
	})
}
