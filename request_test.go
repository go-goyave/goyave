package goyave

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/System-Glitch/goyave/validation"
	"github.com/stretchr/testify/assert"
)

func createTestRequest(rawRequest *http.Request) *Request {
	return &Request{
		httpRequest: rawRequest,
		Rules:       validation.RuleSet{},
		Params:      map[string]string{},
	}
}
func TestRequestContentLength(t *testing.T) {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	request := createTestRequest(rawRequest)
	assert.Equal(t, int64(4), request.ContentLength())
}

func TestRequestMethod(t *testing.T) {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	request := createTestRequest(rawRequest)
	assert.Equal(t, "GET", request.Method())

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("body"))
	request = createTestRequest(rawRequest)
	assert.Equal(t, "POST", request.Method())
}

func TestRequestRemoteAddress(t *testing.T) {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	request := createTestRequest(rawRequest)
	assert.Equal(t, "192.0.2.1:1234", request.RemoteAddress())
}

func TestRequestProtocol(t *testing.T) {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	request := createTestRequest(rawRequest)
	assert.Equal(t, "HTTP/1.1", request.Protocol())
}

func TestRequestURL(t *testing.T) {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	request := createTestRequest(rawRequest)
	assert.Equal(t, "/test-route", request.URL().Path)
}

func TestRequestReferrer(t *testing.T) {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	rawRequest.Header.Set("Referer", "https://www.google.com")
	request := createTestRequest(rawRequest)
	assert.Equal(t, "https://www.google.com", request.Referrer())
}

func TestRequestUserAgent(t *testing.T) {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	rawRequest.Header.Set("User-Agent", "goyave/version")
	request := createTestRequest(rawRequest)
	assert.Equal(t, "goyave/version", request.UserAgent())
}

func TestRequestCookies(t *testing.T) {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	rawRequest.AddCookie(&http.Cookie{
		Name:  "cookie-name",
		Value: "test",
	})
	request := createTestRequest(rawRequest)
	cookies := request.Cookies("cookie-name")
	assert.Equal(t, 1, len(cookies))
	assert.Equal(t, "test", cookies[0].Value)
}

// TODO test request validate
