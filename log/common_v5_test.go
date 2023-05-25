package log

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4/util/testutil"
	"goyave.dev/goyave/v4/util/typeutil"
)

func TestCommonFormatter(t *testing.T) {

	ts := typeutil.Must(time.Parse(time.RFC3339, "2020-03-23T13:58:26.371Z"))

	t.Run("no_user", func(t *testing.T) {
		req := testutil.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		ctx := &LogContext{
			Request: req,
			Status:  http.StatusNoContent,
			Length:  5,
		}
		assert.Equal(t, "192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5", CommonLogFormatterV5(ctx))
	})

	t.Run("user", func(t *testing.T) {
		req := testutil.NewTestRequest(http.MethodGet, "http://user@localhost/log", nil)
		req.Now = ts
		ctx := &LogContext{
			Request: req,
			Status:  http.StatusNoContent,
			Length:  5,
		}
		assert.Equal(t, "192.0.2.1 - user [23/Mar/2020:13:58:26 +0000] \"GET \"http://user@localhost/log\" HTTP/1.1\" 204 5", CommonLogFormatterV5(ctx))
	})

	t.Run("inavlid_ipv6", func(t *testing.T) {
		req := testutil.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		ctx := &LogContext{
			Request: req,
			Status:  http.StatusNoContent,
			Length:  5,
		}
		ctx.Request.Request().RemoteAddr = "[::1"
		assert.Equal(t, "[::1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5", CommonLogFormatterV5(ctx))
	})

	t.Run("http2", func(t *testing.T) {
		req := testutil.NewTestRequest(http.MethodConnect, "/log", nil)
		req.Now = ts
		ctx := &LogContext{
			Request: req,
			Status:  http.StatusNoContent,
			Length:  5,
		}
		ctx.Request.Request().Proto = "HTTP/2.0"
		ctx.Request.Request().ProtoMajor = 2
		assert.Equal(t, "192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"CONNECT \"example.com\" HTTP/2.0\" 204 5", CommonLogFormatterV5(ctx))
	})

	t.Run("no_request_uri", func(t *testing.T) {
		req := testutil.NewTestRequest(http.MethodGet, "/log", nil)
		req.Now = ts
		ctx := &LogContext{
			Request: req,
			Status:  http.StatusNoContent,
			Length:  5,
		}
		ctx.Request.Request().RequestURI = ""
		assert.Equal(t, "192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5", CommonLogFormatterV5(ctx))
	})
}
func TestCombinedFormatter(t *testing.T) {
	ts := typeutil.Must(time.Parse("2006-01-02T15:04:05.000Z", "2020-03-23T13:58:26.371Z"))

	req := testutil.NewTestRequest(http.MethodGet, "/log", nil)
	req.Now = ts
	ctx := &LogContext{
		Request: req,
		Status:  http.StatusNoContent,
		Length:  5,
	}
	referrer := "http://example.com"
	userAgent := "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:74.0) Gecko/20100101 Firefox/74.0"
	ctx.Request.Header().Set("Referer", referrer)
	ctx.Request.Header().Set("User-Agent", userAgent)
	assert.Equal(t, "192.0.2.1 - - [23/Mar/2020:13:58:26 +0000] \"GET \"/log\" HTTP/1.1\" 204 5 \""+referrer+"\" \""+userAgent+"\"", CombinedLogFormatterV5(ctx))
}
