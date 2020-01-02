package goyave

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/cors"
	"github.com/System-Glitch/goyave/v2/validation"
	"github.com/gorilla/mux"
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
	suite.Equal("application/json", result.Header.Get("Content-Type"))
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
	suite.Equal(200, result.StatusCode) // Not written yet
	suite.Equal(404, response.GetStatus())

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
	suite.Equal("application/json", result.Header.Get("Content-Type"))
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

	writer = httptest.NewRecorder()
	router = newRouter()
	router.requestHandler(writer, rawRequest, func(response *Response, request *Request) {
		response.Status(http.StatusNotFound)
	}, validation.RuleSet{})

	result = writer.Result()
	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(404, result.StatusCode)
	suite.Equal("{\"error\":\""+http.StatusText(404)+"\"}\n", string(body))
}

func (suite *RouterTestSuite) TestCORS() {
	router := newRouter()
	suite.Nil(router.corsOptions)

	router.CORS(cors.Default())

	suite.NotNil(router.corsOptions)
	suite.True(router.hasCORSMiddleware)

	route := router.route("GET", "/cors", helloHandler, nil)

	var match mux.RouteMatch
	suite.True(route.Match(httptest.NewRequest("OPTIONS", "/cors", nil), &match))
	suite.True(route.Match(httptest.NewRequest("GET", "/cors", nil), &match))

	writer := httptest.NewRecorder()
	router.Middleware(func(handler Handler) Handler {
		return func(response *Response, request *Request) {
			suite.NotNil(request.corsOptions)
			suite.NotNil(request.CORSOptions())
			handler(response, request)
		}
	})
	rawRequest := httptest.NewRequest("GET", "/cors", nil)
	router.requestHandler(writer, rawRequest, func(response *Response, request *Request) {}, validation.RuleSet{})
}

func (suite *RouterTestSuite) TestPanicStatusHandler() {
	request, response := createRouterTestRequest("/uri")
	response.err = "random error"
	panicStatusHandler(response, request)
	result := response.ResponseWriter.(*httptest.ResponseRecorder).Result()
	suite.Equal(500, result.StatusCode)
}

func (suite *RouterTestSuite) TestErrorStatusHandler() {
	request, response := createRouterTestRequest("/uri")
	response.Status(404)
	errorStatusHandler(response, request)
	result := response.ResponseWriter.(*httptest.ResponseRecorder).Result()
	suite.Equal(404, result.StatusCode)
	suite.Equal("application/json", result.Header.Get("Content-Type"))

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal("{\"error\":\""+http.StatusText(404)+"\"}\n", string(body))
}

func (suite *RouterTestSuite) TestStatusHandlers() {
	rawRequest := httptest.NewRequest("GET", "/uri", nil)
	writer := httptest.NewRecorder()
	router := newRouter()
	router.StatusHandler(func(response *Response, request *Request) {
		response.String(http.StatusInternalServerError, "An unexpected panic occurred.")
	}, http.StatusInternalServerError)
	router.requestHandler(writer, rawRequest, func(response *Response, request *Request) {
		panic("Panic")
	}, validation.RuleSet{})

	result := writer.Result()
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(500, result.StatusCode)
	suite.Equal("An unexpected panic occurred.", string(body))

	// On subrouters
	subrouter := router.Subrouter("/sub")
	writer = httptest.NewRecorder()
	router = newRouter()
	subrouter.requestHandler(writer, rawRequest, func(response *Response, request *Request) {
		panic("Panic")
	}, validation.RuleSet{})

	result = writer.Result()
	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(500, result.StatusCode)
	suite.Equal("An unexpected panic occurred.", string(body))

	// Multiple statuses
	writer = httptest.NewRecorder()
	subrouter.StatusHandler(func(response *Response, request *Request) {
		response.String(response.GetStatus(), http.StatusText(response.GetStatus()))
	}, 400, 404)
	subrouter.requestHandler(writer, rawRequest, func(response *Response, request *Request) {
		response.Status(400)
	}, validation.RuleSet{})

	result = writer.Result()
	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(400, result.StatusCode)
	suite.Equal(http.StatusText(400), string(body))

	writer = httptest.NewRecorder()
	subrouter.requestHandler(writer, rawRequest, func(response *Response, request *Request) {
		response.Status(404)
	}, validation.RuleSet{})

	result = writer.Result()
	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(404, result.StatusCode)
	suite.Equal(http.StatusText(404), string(body))
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
