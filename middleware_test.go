package goyave

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/System-Glitch/goyave/config"
	"github.com/System-Glitch/goyave/lang"
	"github.com/System-Glitch/goyave/validation"
	"github.com/stretchr/testify/suite"
)

type MiddlewareTestSuite struct {
	suite.Suite
}

func (suite *MiddlewareTestSuite) SetupSuite() {
	config.LoadConfig()
	lang.LoadDefault()
}

func testMiddleware(middleware Middleware, rawRequest *http.Request, handler func(*Response, *Request)) *http.Response {
	request := &Request{
		httpRequest: rawRequest,
		Rules:       validation.RuleSet{},
		Params:      map[string]string{},
	}
	response := &Response{
		writer: httptest.NewRecorder(),
		empty:  true,
	}
	middleware(handler)(response, request)

	return response.writer.(*httptest.ResponseRecorder).Result()
}

func (suite *MiddlewareTestSuite) TestRecoveryMiddlewarePanicDebug() {
	config.Set("debug", true)
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	resp := testMiddleware(recoveryMiddleware, rawRequest, func(response *Response, r *Request) {
		panic(fmt.Errorf("error message"))
	})
	body, _ := ioutil.ReadAll(resp.Body)
	suite.Equal(500, resp.StatusCode)
	suite.Equal("{\"error\":\"error message\"}\n", string(body))
}

func (suite *MiddlewareTestSuite) TestRecoveryMiddlewarePanicNoDebug() {
	config.Set("debug", false)
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	resp := testMiddleware(recoveryMiddleware, rawRequest, func(response *Response, r *Request) {
		panic(fmt.Errorf("error message"))
	})
	body, _ := ioutil.ReadAll(resp.Body)
	suite.Equal(500, resp.StatusCode)
	suite.Equal("", string(body))
	config.Set("debug", true)
}

func (suite *MiddlewareTestSuite) TestRecoveryMiddlewareNoPanic() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	resp := testMiddleware(recoveryMiddleware, rawRequest, func(response *Response, r *Request) {
		response.String(200, "message")
	})
	body, _ := ioutil.ReadAll(resp.Body)
	suite.Equal(200, resp.StatusCode)
	suite.Equal("message", string(body))
}

func (suite *MiddlewareTestSuite) TestLanguageMiddleware() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	rawRequest.Header.Set("Accept-Language", "en-US")
	testMiddleware(languageMiddleware, rawRequest, func(response *Response, r *Request) {
		suite.Equal("en-US", r.Lang)
	})

	rawRequest = httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	testMiddleware(languageMiddleware, rawRequest, func(response *Response, r *Request) {
		suite.Equal("en-US", r.Lang)
	})
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}
