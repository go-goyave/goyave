package goyave

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"goyave.dev/goyave/v4/cors"
	"goyave.dev/goyave/v4/util/fsutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4/validation"
)

func createTestRequest(rawRequest *http.Request) *Request {
	return &Request{
		httpRequest: rawRequest,
		Rules:       &validation.Rules{},
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
	assert.Equal(t, "/test-route", request.URI().Path)
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
	cookies := request.Cookies()
	assert.Equal(t, 1, len(cookies))
	assert.Equal(t, "cookie-name", cookies[0].Name)
	assert.Equal(t, "test", cookies[0].Value)
}

func TestRequestValidate(t *testing.T) {
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	request := createTestRequest(rawRequest)

	validation.AddRule("test_extra_request", &validation.RuleDefinition{
		Function: func(ctx *validation.Context) bool {
			r, ok := ctx.Extra["request"].(*Request)
			assert.True(t, ok)
			assert.Equal(t, request, r)
			return true
		},
	})

	request.Data = map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}
	request.Rules = validation.RuleSet{
		"string": validation.List{"required", "string", "test_extra_request"},
		"number": validation.List{"required", "numeric", "min:10"},
	}.AsRules()
	errors := request.validate()
	assert.Nil(t, errors)

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world"))
	rawRequest.Header.Set("Content-Type", "application/json")
	request = createTestRequest(rawRequest)
	request.Data = map[string]interface{}{
		"string": "hello world",
	}

	request.Rules = &validation.Rules{
		Fields: validation.FieldMap{
			"string": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "string"},
				},
			},
			"number": &validation.Field{
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "numeric"},
					{Name: "min", Params: []string{"50"}},
				},
			},
		},
	}
	errors = request.validate()
	assert.NotNil(t, errors)
	assert.Equal(t, 2, len(errors["number"].Errors))

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	request = createTestRequest(rawRequest)
	request.Rules = nil
	errors = request.validate()
	assert.Nil(t, errors)
}

func TestRequestAccessors(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}

	uid, err := uuid.Parse("3bbcee75-cecc-5b56-8031-b6641c1ed1f1")
	if err != nil {
		panic(err)
	}

	date, err := time.Parse("2006-01-02", "2019-11-21")
	if err != nil {
		panic(err)
	}

	url, err := url.ParseRequestURI("https://google.com")
	if err != nil {
		panic(err)
	}

	rawRequest := httptest.NewRequest("POST", "/test-route", nil)
	request := createTestRequest(rawRequest)
	request.Data = map[string]interface{}{
		"string":   "hello world",
		"integer":  42,
		"numeric":  42.3,
		"bool":     true,
		"file":     []fsutil.File{{MIMEType: "image/png"}},
		"timezone": loc,
		"ip":       net.ParseIP("127.0.0.1"),
		"uuid":     uid,
		"date":     date,
		"url":      url,
		"object": map[string]interface{}{
			"hello": "world",
		},
	}

	assert.Equal(t, "hello world", request.String("string"))
	assert.Equal(t, 42, request.Integer("integer"))
	assert.Equal(t, 42.3, request.Numeric("numeric"))
	assert.Equal(t, rawRequest, request.Request())
	assert.True(t, request.Bool("bool"))

	files := request.File("file")
	assert.Len(t, files, 1)
	assert.Equal(t, "image/png", files[0].MIMEType)

	assert.Equal(t, "America/New_York", request.Timezone("timezone").String())
	assert.Equal(t, "127.0.0.1", request.IP("ip").String())
	assert.Equal(t, "3bbcee75-cecc-5b56-8031-b6641c1ed1f1", request.UUID("uuid").String())
	assert.Equal(t, "2019-11-21 00:00:00 +0000 UTC", request.Date("date").String())
	assert.Equal(t, "https://google.com", request.URL("url").String())
	assert.Equal(t, request.Data["object"], request.Object("object"))

	assert.Panics(t, func() { request.String("integer") })
	assert.Panics(t, func() { request.Integer("string") })
	assert.Panics(t, func() { request.Numeric("string") })
	assert.Panics(t, func() { request.Bool("string") })
	assert.Panics(t, func() { request.File("string") })
	assert.Panics(t, func() { request.Timezone("string") })
	assert.Panics(t, func() { request.IP("string") })
	assert.Panics(t, func() { request.UUID("string") })
	assert.Panics(t, func() { request.Date("string") })
	assert.Panics(t, func() { request.URL("string") })
	assert.Panics(t, func() { request.String("doesn't exist") })
	assert.Panics(t, func() { request.Object("doesn't exist") })
}

func TestRequestHas(t *testing.T) {
	request := createTestRequest(httptest.NewRequest("POST", "/test-route", nil))
	request.Data = map[string]interface{}{
		"string": "hello world",
	}

	assert.True(t, request.Has("string"))
	assert.False(t, request.Has("not_in_request"))
}

func TestRequestCors(t *testing.T) {
	request := createTestRequest(httptest.NewRequest("POST", "/test-route", nil))

	assert.Nil(t, request.CORSOptions())

	request.corsOptions = cors.Default()
	options := request.CORSOptions()
	assert.NotNil(t, options)

	// Check cannot alter config
	options.MaxAge = time.Second
	assert.NotEqual(t, request.corsOptions.MaxAge, options.MaxAge)
}

func TestGetBearerToken(t *testing.T) {
	request := createTestRequest(httptest.NewRequest("POST", "/test-route", nil))
	request.Header().Set("Authorization", "NotBearer 123456789")
	token, ok := request.BearerToken()
	assert.Empty(t, token)
	assert.False(t, ok)

	request.Header().Set("Authorization", "Bearer123456789")
	token, ok = request.BearerToken()
	assert.Empty(t, token)
	assert.False(t, ok)

	request.Header().Set("Authorization", "Bearer 123456789")
	token, ok = request.BearerToken()
	assert.Equal(t, "123456789", token)
	assert.True(t, ok)
}

func TestToStruct(t *testing.T) {
	type UserInsertRequest struct {
		Username string
		Email    string
	}

	request := createTestRequest(httptest.NewRequest("POST", "/test-route", nil))
	request.Data = map[string]interface{}{
		"username": "johndoe",
		"email":    "johndoe@example.org",
	}

	userInsertRequest := UserInsertRequest{}

	if err := request.ToStruct(&userInsertRequest); err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, "johndoe", userInsertRequest.Username)
	assert.Equal(t, "johndoe@example.org", userInsertRequest.Email)
}
