package log

import (
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/util/testutil"
)

func TestCommonFormatter(t *testing.T) {

	ts := lo.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))

	t.Run("no_user", func(t *testing.T) {
		req := testutil.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		ctx := &Context{
			Request: req,
			Status:  http.StatusNoContent,
			Length:  5,
		}
		message, attrs := CommonLogFormatter(ctx)
		assert.Equal(t, "192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5", message)
		assert.Equal(t, []slog.Attr{slog.Group("details",
			slog.String("host", "192.0.2.1"),
			slog.String("username", "-"),
			slog.Time("time", ctx.Request.Now),
			slog.String("method", http.MethodGet),
			slog.String("uri", "/log"),
			slog.String("proto", "HTTP/1.1"),
			slog.Int("status", 204),
			slog.Int("length", 5),
		)}, attrs)
	})

	t.Run("user", func(t *testing.T) {
		req := testutil.NewTestRequest(http.MethodGet, "http://user@localhost/log", nil)
		req.Now = ts
		ctx := &Context{
			Request: req,
			Status:  http.StatusNoContent,
			Length:  5,
		}
		message, attrs := CommonLogFormatter(ctx)
		assert.Equal(t, "192.0.2.1 - user [23/Mar/2020:13:58:26 +0000] \"GET \"http://user@localhost/log\" HTTP/1.1\" 204 5", message)
		assert.Equal(t, []slog.Attr{slog.Group("details",
			slog.String("host", "192.0.2.1"),
			slog.String("username", "-"),
			slog.Time("time", ctx.Request.Now),
			slog.String("method", http.MethodGet),
			slog.String("uri", "/log"),
			slog.String("proto", "HTTP/1.1"),
			slog.Int("status", 204),
			slog.Int("length", 5),
		)}, attrs)
	})

	t.Run("inavlid_ipv6", func(t *testing.T) {
		req := testutil.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		ctx := &Context{
			Request: req,
			Status:  http.StatusNoContent,
			Length:  5,
		}
		ctx.Request.Request().RemoteAddr = "[::1"
		message, attrs := CommonLogFormatter(ctx)
		assert.Equal(t, "[::1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5", message)
		assert.Equal(t, []slog.Attr{slog.Group("details",
			slog.String("host", "[::1"),
			slog.String("username", "-"),
			slog.Time("time", ctx.Request.Now),
			slog.String("method", http.MethodGet),
			slog.String("uri", "/log"),
			slog.String("proto", "HTTP/1.1"),
			slog.Int("status", 204),
			slog.Int("length", 5),
		)}, attrs)
	})

	t.Run("http2", func(t *testing.T) {
		req := testutil.NewTestRequest(http.MethodConnect, "/log", nil)
		req.Now = ts
		ctx := &Context{
			Request: req,
			Status:  http.StatusNoContent,
			Length:  5,
		}
		ctx.Request.Request().Proto = "HTTP/2.0"
		ctx.Request.Request().ProtoMajor = 2
		message, attrs := CommonLogFormatter(ctx)
		assert.Equal(t, "192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"CONNECT \"example.com\" HTTP/2.0\" 204 5", message)
		assert.Equal(t, []slog.Attr{slog.Group("details",
			slog.String("host", "192.0.2.1"),
			slog.String("username", "-"),
			slog.Time("time", ctx.Request.Now),
			slog.String("method", http.MethodConnect),
			slog.String("uri", "/log"),
			slog.String("proto", "HTTP/2.0"),
			slog.Int("status", 204),
			slog.Int("length", 5),
		)}, attrs)
	})

	t.Run("no_request_uri", func(t *testing.T) {
		req := testutil.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		ctx := &Context{
			Request: req,
			Status:  http.StatusNoContent,
			Length:  5,
		}
		ctx.Request.Request().RequestURI = ""
		message, attrs := CommonLogFormatter(ctx)
		assert.Equal(t, "192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5", message)
		assert.Equal(t, []slog.Attr{slog.Group("details",
			slog.String("host", "192.0.2.1"),
			slog.String("username", "-"),
			slog.Time("time", ctx.Request.Now),
			slog.String("method", http.MethodGet),
			slog.String("uri", "/log"),
			slog.String("proto", "HTTP/1.1"),
			slog.Int("status", 204),
			slog.Int("length", 5),
		)}, attrs)
	})
}
func TestCombinedFormatter(t *testing.T) {
	ts := lo.Must(time.Parse("2006-01-02T15:04:05.000Z", "2020-03-23T13:58:26.371Z"))

	req := testutil.NewTestRequest(http.MethodGet, "/log", nil)
	req.Now = ts
	ctx := &Context{
		Request: req,
		Status:  http.StatusNoContent,
		Length:  5,
	}
	referrer := "http://example.com"
	userAgent := "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:74.0) Gecko/20100101 Firefox/74.0"
	ctx.Request.Header().Set("Referer", referrer)
	ctx.Request.Header().Set("User-Agent", userAgent)
	message, attrs := CombinedLogFormatter(ctx)
	assert.Equal(t, "192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5 \""+referrer+"\" \""+userAgent+"\"", message)
	assert.Equal(t, []slog.Attr{slog.Group("details",
		slog.String("host", "192.0.2.1"),
		slog.String("username", "-"),
		slog.Time("time", ctx.Request.Now),
		slog.String("method", http.MethodGet),
		slog.String("uri", "/log"),
		slog.String("proto", "HTTP/1.1"),
		slog.Int("status", 204),
		slog.Int("length", 5),
		slog.String("referrer", referrer),
		slog.String("userAgent", userAgent),
	)}, attrs)
}
