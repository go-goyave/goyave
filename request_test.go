package goyave

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {

	t.Run("NewRequest", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/test", nil)
		r := NewRequest(httpReq)

		assert.Equal(t, httpReq, r.httpRequest)
		assert.False(t, r.Now.IsZero())
		assert.NotNil(t, r.Extra)
	})

	t.Run("Accessors", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("hello"))
		httpReq.Header.Set("X-Test", "value")
		httpReq.Header.Set("Referer", "referrer")
		httpReq.Header.Set("User-Agent", "useragent")
		httpReq.SetBasicAuth("username", "password")
		httpReq.AddCookie(&http.Cookie{Name: "test-cookie", Value: "value"})
		r := NewRequest(httpReq)

		assert.Equal(t, httpReq, r.Request())
		assert.Equal(t, http.MethodPost, r.Method())
		assert.Equal(t, "HTTP/1.1", r.Protocol())
		assert.Equal(t, "/test", r.URL().String())

		assert.Equal(t, httpReq.Header, r.Header())
		assert.Equal(t, "/test", r.URL().String())
		assert.Equal(t, int64(5), r.ContentLength())
		assert.Equal(t, "192.0.2.1:1234", r.RemoteAddress())

		cookies := r.Cookies()
		assert.Equal(t, cookies, r.cookies)
		assert.Equal(t, []*http.Cookie{{Name: "test-cookie", Value: "value"}}, cookies)

		assert.Equal(t, "referrer", r.Referrer())
		assert.Equal(t, "useragent", r.UserAgent())

		username, password, ok := r.BasicAuth()
		assert.Equal(t, "username", username)
		assert.Equal(t, "password", password)
		assert.True(t, ok)

		assert.NotNil(t, r.Body())
		httpReq = httptest.NewRequest(http.MethodGet, "/test", nil)
		assert.NotNil(t, NewRequest(httpReq).Body())
	})

	t.Run("BearerToken", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/test", nil)
		httpReq.Header.Set("Authorization", "Bearer  token  ")
		r := NewRequest(httpReq)

		token, ok := r.BearerToken()
		assert.Equal(t, "token", token)
		assert.True(t, ok)

		httpReq = httptest.NewRequest(http.MethodGet, "/test", nil)
		r = NewRequest(httpReq)

		token, ok = r.BearerToken()
		assert.Empty(t, token)
		assert.False(t, ok)
	})

	t.Run("Context", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/test", nil)
		r := NewRequest(httpReq)

		ctx := r.Context()
		if !assert.NotNil(t, ctx) {
			return
		}
		assert.Equal(t, httpReq.Context(), ctx)

		key := struct{}{}
		r2 := r.WithContext(context.WithValue(ctx, key, "value"))
		assert.Equal(t, r, r2)

		ctx2 := r.Context()
		assert.Equal(t, "value", ctx2.Value(key))
	})
}
