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

func TestRequestValidate(t *testing.T) {
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	request := createTestRequest(rawRequest)
	request.Data = map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}
	request.Rules = validation.RuleSet{
		"string": {"required", "string"},
		"number": {"required", "numeric", "min:10"},
	}
	errors := request.validate()
	assert.Nil(t, errors)

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world"))
	rawRequest.Header.Set("Content-Type", "application/json")
	request = createTestRequest(rawRequest)
	request.Data = map[string]interface{}{
		"string": "hello world",
	}
	request.Rules = validation.RuleSet{
		"string": {"required", "string"},
		"number": {"required", "numeric", "min:50"},
	}
	errors = request.validate()
	assert.NotNil(t, errors)
	assert.Equal(t, 2, len(errors["validationError"]["number"]))

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	request = createTestRequest(rawRequest)
	request.Rules = nil
	errors = request.validate()
	assert.Nil(t, errors)
}
