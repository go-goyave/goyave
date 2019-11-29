package goyave

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/validation"
	"github.com/stretchr/testify/suite"
)

type RouterTestSuite struct {
	suite.Suite
	middlewareExecuted bool
}

func createRouterTestRequest(url string) (*Request, *Response) {
	rawRequest := httptest.NewRequest("GET", url, nil)
	request := &Request{
		httpRequest: rawRequest,
		Params:      map[string]string{"resource": url},
	}
	response := &Response{
		ResponseWriter: httptest.NewRecorder(),
		empty:          true,
	}
	return request, response
}

func (suite *RouterTestSuite) routerTestMiddleware(handler Handler) Handler {
	return func(response *Response, request *Request) {
		suite.middlewareExecuted = true
		handler(response, request)
	}
}

func (suite *RouterTestSuite) SetupSuite() {
	config.Load()
}

func (suite *RouterTestSuite) TestNewRouter() {
	router := newRouter()
	suite.NotNil(router)
	suite.Equal(3, len(router.middleware))
}

func (suite *RouterTestSuite) TestRouterMiddleware() {
	router := newRouter()
	router.Middleware(suite.routerTestMiddleware)
	suite.Equal(4, len(router.middleware))
}

func (suite *RouterTestSuite) TestSubRouter() {
	router := newRouter()
	router.Middleware(suite.routerTestMiddleware)
	suite.Equal(4, len(router.middleware))

	subrouter := router.Subrouter("/sub")
	suite.Equal(4, len(subrouter.middleware))
}

func (suite *RouterTestSuite) TestCleanStaticPath() {
	suite.Equal("config/index.html", cleanStaticPath("config", "index.html"))
	suite.Equal("config/index.html", cleanStaticPath("config", ""))
	suite.Equal("config/defaults.json", cleanStaticPath("config", "defaults.json"))
	suite.Equal("resources/lang/en-US/locale.json", cleanStaticPath("resources", "lang/en-US/locale.json"))
	suite.Equal("resources/lang/en-US/locale.json", cleanStaticPath("resources", "/lang/en-US/locale.json"))
	suite.Equal("resources/lang/en-US/index.html", cleanStaticPath("resources", "lang/en-US"))
	suite.Equal("resources/lang/en-US/index.html", cleanStaticPath("resources", "lang/en-US/"))
	suite.Equal("resources/lang/index.html", cleanStaticPath("resources", "lang"))
	suite.Equal("resources/lang/index.html", cleanStaticPath("resources", "lang/"))
}

func (suite *RouterTestSuite) TestStaticHandler() {
	request, response := createRouterTestRequest("/config.test.json")
	handler := staticHandler("config", false)
	handler(response, request)
	result := response.ResponseWriter.(*httptest.ResponseRecorder).Result()
	suite.Equal(200, result.StatusCode)
	suite.Equal("application/octet-stream", result.Header.Get("Content-Type"))
	suite.Equal("inline", result.Header.Get("Content-Disposition"))

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}

	suite.True(len(body) > 0)

	request, response = createRouterTestRequest("/doesn'texist")
	handler = staticHandler("config", false)
	handler(response, request)
	result = response.ResponseWriter.(*httptest.ResponseRecorder).Result()
	suite.Equal(404, result.StatusCode)

	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}

	suite.Equal(0, len(body))

	request, response = createRouterTestRequest("/config.test.json")
	handler = staticHandler("config", true)
	handler(response, request)
	result = response.ResponseWriter.(*httptest.ResponseRecorder).Result()
	suite.Equal(200, result.StatusCode)
	suite.Equal("application/octet-stream", result.Header.Get("Content-Type"))
	suite.Equal("attachment; filename=\"config.test.json\"", result.Header.Get("Content-Disposition"))

	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}

	suite.True(len(body) > 0)
}

func (suite *RouterTestSuite) TestRequestHandler() {
	rawRequest := httptest.NewRequest("GET", "/uri", nil)
	writer := httptest.NewRecorder()
	router := newRouter()
	router.requestHandler(writer, rawRequest, func(response *Response, request *Request) {
		response.String(200, "Hello world")
	}, validation.RuleSet{})

	result := writer.Result()
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(200, result.StatusCode)
	suite.Equal("Hello world", string(body))

	writer = httptest.NewRecorder()
	router = newRouter()
	router.Middleware(suite.routerTestMiddleware)
	router.requestHandler(writer, rawRequest, func(response *Response, request *Request) {}, validation.RuleSet{})

	result = writer.Result()
	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(204, result.StatusCode)
	suite.Equal(0, len(body))
	suite.True(suite.middlewareExecuted)
	suite.middlewareExecuted = false
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
