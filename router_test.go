package goyave

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/cors"
)

type RouterTestSuite struct {
	TestSuite
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

func (suite *RouterTestSuite) createOrderedTestMiddleware(result *string, str string) Middleware {
	return func(next Handler) Handler {
		return func(response *Response, r *Request) {
			*result += str
			next(response, r)
		}
	}
}

func (suite *RouterTestSuite) TearDownTest() {
	regexCache = nil
}

func (suite *RouterTestSuite) TestNewRouter() {
	router := newRouter()
	suite.NotNil(router)
	suite.Equal(3, len(router.middleware))
}

func (suite *RouterTestSuite) TestRouterRegisterRoute() {
	router := newRouter()
	route := router.Route("GET", "/uri", func(resp *Response, r *Request) {}, nil)
	suite.Contains(router.routes, route)
	suite.Equal(router, route.parent)

	route = router.Route("GET", "/", func(resp *Response, r *Request) {}, nil)
	suite.Equal("", route.uri)
	suite.Equal(router, route.parent)

	route = router.Route("GET|POST", "/", func(resp *Response, r *Request) {}, nil)
	suite.Equal([]string{"GET", "POST"}, route.methods)
	suite.Equal(router, route.parent)
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
	suite.Contains(router.subrouters, subrouter)
	suite.Equal(0, len(subrouter.middleware)) // Middleware inherited, not copied
	suite.Equal(len(router.statusHandlers), len(subrouter.statusHandlers))

	router = newRouter()
	subrouter = router.Subrouter("/")
	suite.Equal("", subrouter.prefix)
}

func (suite *RouterTestSuite) TestCleanStaticPath() {
	suite.Equal("config/index.html", cleanStaticPath("config", "index.html"))
	suite.Equal("config/index.html", cleanStaticPath("config", ""))
	suite.Equal("config/defaults.json", cleanStaticPath("config", "defaults.json"))
	suite.Equal("resources/lang/en-US/locale.json", cleanStaticPath("resources", "lang/en-US/locale.json"))
	suite.Equal("resources/lang/en-US/locale.json", cleanStaticPath("resources", "/lang/en-US/locale.json"))
	suite.Equal("resources/img/logo/index.html", cleanStaticPath("resources", "img/logo"))
	suite.Equal("resources/img/logo/index.html", cleanStaticPath("resources", "img/logo/"))
	suite.Equal("resources/img/index.html", cleanStaticPath("resources", "img"))
	suite.Equal("resources/img/index.html", cleanStaticPath("resources", "img/"))
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

	route := &Route{}
	var tmp *Route
	route.handler = func(response *Response, request *Request) {
		tmp = request.Route()
		response.String(200, "Hello world")
	}
	match := &routeMatch{route: route}
	router.requestHandler(match, writer, rawRequest)
	suite.Equal(route, tmp)

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

	match = &routeMatch{
		route: &Route{
			handler: func(response *Response, request *Request) {},
			parent:  router,
		},
	}
	router.requestHandler(match, writer, rawRequest)

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
	match = &routeMatch{
		route: &Route{
			handler: func(response *Response, request *Request) {
				response.Status(http.StatusNotFound)
			},
		},
	}
	router.requestHandler(match, writer, rawRequest)

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

	route := router.registerRoute("GET", "/cors", helloHandler, nil)
	suite.Equal([]string{"GET", "OPTIONS"}, route.methods)

	match := routeMatch{currentPath: "/cors"}
	suite.True(route.match(httptest.NewRequest("OPTIONS", "/cors", nil), &match))
	match = routeMatch{currentPath: "/cors"}
	suite.True(route.match(httptest.NewRequest("GET", "/cors", nil), &match))

	writer := httptest.NewRecorder()
	router.Middleware(func(handler Handler) Handler {
		return func(response *Response, request *Request) {
			suite.NotNil(request.corsOptions)
			suite.NotNil(request.CORSOptions())
			handler(response, request)
		}
	})
	rawRequest := httptest.NewRequest("GET", "/cors", nil)

	match = routeMatch{
		route: &Route{
			handler: func(response *Response, request *Request) {},
		},
	}
	router.requestHandler(&match, writer, rawRequest)
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

	match := &routeMatch{
		route: &Route{
			handler: func(response *Response, request *Request) {
				panic("Panic")
			},
			parent: router,
		},
	}
	router.requestHandler(match, writer, rawRequest)

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

	subrouter.requestHandler(match, writer, rawRequest)

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

	match = &routeMatch{
		route: &Route{
			handler: func(response *Response, request *Request) {
				response.Status(400)
			},
		},
	}
	subrouter.requestHandler(match, writer, rawRequest)

	result = writer.Result()
	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(400, result.StatusCode)
	suite.Equal(http.StatusText(400), string(body))

	writer = httptest.NewRecorder()

	match = &routeMatch{
		route: &Route{
			handler: func(response *Response, request *Request) {
				response.Status(404)
			},
		},
	}
	subrouter.requestHandler(match, writer, rawRequest)

	result = writer.Result()
	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(404, result.StatusCode)
	suite.Equal(http.StatusText(404), string(body))
}

func (suite *RouterTestSuite) TestRouteNoMatch() {
	rawRequest := httptest.NewRequest("GET", "/uri", nil)
	writer := httptest.NewRecorder()
	router := newRouter()

	match := &routeMatch{route: notFoundRoute}
	router.requestHandler(match, writer, rawRequest)
	result := writer.Result()
	suite.Equal(http.StatusNotFound, result.StatusCode)

	writer = httptest.NewRecorder()
	match = &routeMatch{route: methodNotAllowedRoute}
	router.requestHandler(match, writer, rawRequest)
	result = writer.Result()
	suite.Equal(http.StatusMethodNotAllowed, result.StatusCode)
}

func (suite *RouterTestSuite) TestNamedRoutes() {
	r := newRouter()
	route := r.Route("GET", "/uri", func(resp *Response, r *Request) {}, nil)
	route.Name("get-uri")
	suite.Equal(route, r.namedRoutes["get-uri"])
	suite.Equal(route, r.GetRoute("get-uri"))

	subrouter := r.Subrouter("/sub")
	suite.Equal(route, subrouter.GetRoute("get-uri"))

	route2 := r.Route("GET", "/other-route", func(resp *Response, r *Request) {}, nil)
	suite.Panics(func() {
		route2.Name("get-uri")
	})
	suite.Empty(route2.GetName())

	// Global router
	router = r
	suite.Equal(route, GetRoute("get-uri"))
	router = nil
}

func (suite *RouterTestSuite) TestMiddleware() {
	// Test the middleware execution order
	result := ""
	middleware := make([]Middleware, 0, 4)
	for i := 0; i < 4; i++ {
		middleware = append(middleware, suite.createOrderedTestMiddleware(&result, strconv.Itoa(i+1)))
	}
	router := newRouter()
	router.Middleware(middleware[0])

	subrouter := router.Subrouter("/")
	subrouter.Middleware(middleware[1])

	handler := func(response *Response, r *Request) {
		result += "5"
	}
	route := subrouter.Route("GET", "/hello", handler, nil, middleware[2], middleware[3])

	rawRequest := httptest.NewRequest("GET", "/hello", nil)
	match := routeMatch{
		route:       route,
		currentPath: rawRequest.URL.Path,
	}
	router.requestHandler(&match, httptest.NewRecorder(), rawRequest)

	suite.Equal("12345", result)
}

func (suite *RouterTestSuite) TestCoreMiddleware() {
	// Ensure core middleware is executed on Not Found and Method Not Allowed
	router := newRouter()

	match := &routeMatch{
		route: newRoute(func(response *Response, request *Request) {
			panic("Test panic") // Test recover middleware is executed
		}),
	}

	writer := httptest.NewRecorder()
	router.requestHandler(match, writer, httptest.NewRequest("GET", "/uri", nil))

	result := writer.Result()
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(500, result.StatusCode)
	suite.Equal("{\"error\":\"Internal Server Error\"}\n", string(body))

	lang := ""
	param := ""
	match = &routeMatch{
		route: newRoute(func(response *Response, request *Request) {
			// Test lang and parse request
			lang = request.Lang
			param = request.String("param")
		}),
	}

	writer = httptest.NewRecorder()
	router.requestHandler(match, writer, httptest.NewRequest("GET", "/uri?param=param", nil))
	suite.Equal("en-US", lang)
	suite.Equal("param", param)

	// Custom middleware shouldn't be executed
	strResult := ""
	testMiddleware := suite.createOrderedTestMiddleware(&strResult, "1")
	router.Middleware(testMiddleware)

	match = &routeMatch{route: notFoundRoute}
	router.requestHandler(match, httptest.NewRecorder(), httptest.NewRequest("GET", "/uri", nil))
	suite.Empty(strResult)

	match = &routeMatch{route: methodNotAllowedRoute}
	router.requestHandler(match, httptest.NewRecorder(), httptest.NewRequest("GET", "/uri", nil))
	suite.Empty(strResult)

}

func (suite *RouterTestSuite) TestMiddlewareHolder() {
	result := ""
	testMiddleware := suite.createOrderedTestMiddleware(&result, "1")
	secondTestMiddleware := suite.createOrderedTestMiddleware(&result, "2")

	holder := &middlewareHolder{[]Middleware{testMiddleware, secondTestMiddleware}}
	handler := holder.applyMiddleware(func(response *Response, r *Request) {
		result += "3"
	})
	handler(suite.CreateTestResponse(httptest.NewRecorder()), suite.CreateTestRequest(nil))
	suite.Equal("123", result)
}

func (suite *RouterTestSuite) TestTrimCurrentPath() {
	routeMatch := routeMatch{currentPath: "/product/55"}
	routeMatch.trimCurrentPath("/product")
	suite.Equal("/55", routeMatch.currentPath)
}

func (suite *RouterTestSuite) TestMatch() {
	handler := func(response *Response, r *Request) {
		response.String(http.StatusOK, "Hello")
	}

	router := newRouter()
	router.Route("GET|POST", "/hello", handler, nil).Name("hello")
	router.Route("PUT", "/hello", handler, nil).Name("hello.put")
	router.Route("GET", "/hello/sub", handler, nil).Name("hello.sub")

	productRouter := router.Subrouter("/product")
	productRouter.Route("GET", "/", handler, nil).Name("product.index")
	productRouter.Route("GET", "/{id:[0-9]+}", handler, nil).Name("product.show")
	productRouter.Route("GET", "/{id:[0-9]+}/details", handler, nil).Name("product.show.details")

	userRouter := router.Subrouter("/user")
	userRouter.Route("GET", "/", handler, nil).Name("user.index")
	userRouter.Route("GET", "/{id:[0-9]+}", handler, nil).Name("user.show")

	router.Subrouter("/empty")

	match := routeMatch{currentPath: "/hello"}
	suite.True(router.match(httptest.NewRequest("GET", "/hello", nil), &match))
	suite.Equal(router.GetRoute("hello"), match.route)

	match = routeMatch{currentPath: "/hello/sub"}
	suite.True(router.match(httptest.NewRequest("GET", "/hello/sub", nil), &match))
	suite.Equal(router.GetRoute("hello.sub"), match.route)

	match = routeMatch{currentPath: "/product"}
	suite.True(router.match(httptest.NewRequest("GET", "/product", nil), &match))
	suite.Equal(router.GetRoute("product.index"), match.route)

	match = routeMatch{currentPath: "/product/5"}
	suite.True(router.match(httptest.NewRequest("GET", "/product/5", nil), &match))
	suite.Equal(router.GetRoute("product.show"), match.route)
	suite.Equal("5", match.parameters["id"])

	match = routeMatch{currentPath: "/product/5/details"}
	suite.True(router.match(httptest.NewRequest("GET", "/product/5/details", nil), &match))
	suite.Equal(router.GetRoute("product.show.details"), match.route)
	suite.Equal("5", match.parameters["id"])

	match = routeMatch{currentPath: "/user"}
	suite.True(router.match(httptest.NewRequest("GET", "/user", nil), &match))
	suite.Equal(router.GetRoute("user.index"), match.route)

	match = routeMatch{currentPath: "/user/42"}
	suite.True(router.match(httptest.NewRequest("GET", "/user/42", nil), &match))
	suite.Equal(router.GetRoute("user.show"), match.route)
	suite.Equal("42", match.parameters["id"])

	match = routeMatch{currentPath: "/product/notaroute"}
	suite.False(router.match(httptest.NewRequest("GET", "/product/notaroute", nil), &match))
	suite.Equal(notFoundRoute, match.route)

	match = routeMatch{currentPath: "/empty"}
	suite.False(router.match(httptest.NewRequest("GET", "/empty", nil), &match))
	suite.Equal(notFoundRoute, match.route)

	match = routeMatch{currentPath: "/product"}
	suite.True(router.match(httptest.NewRequest("DELETE", "/product", nil), &match))
	suite.Equal(methodNotAllowedRoute, match.route)

	// ------------

	paramSubrouter := router.Subrouter("/{param}")
	route := paramSubrouter.Route("GET", "/{subparam}", handler, nil).Name("param.name")
	match = routeMatch{currentPath: "/name/surname"}
	suite.True(router.match(httptest.NewRequest("GET", "/name/surname", nil), &match))
	suite.Equal(route, match.route)
	suite.Equal("name", match.parameters["param"])
	suite.Equal("surname", match.parameters["subparam"])

	// ------------

	match = routeMatch{currentPath: "/user/42"}
	suite.False(productRouter.match(httptest.NewRequest("GET", "/user/42", nil), &match))
	match = routeMatch{currentPath: "/product/42"}
	suite.True(productRouter.match(httptest.NewRequest("GET", "/product/42", nil), &match))
	suite.Equal(router.GetRoute("product.show"), match.route)
	suite.Equal("42", match.parameters["id"])

	match = routeMatch{currentPath: "/user/42/extra"}
	suite.False(userRouter.match(httptest.NewRequest("GET", "/user/42/extra", nil), &match))
}

func (suite *RouterTestSuite) TestScheme() {
	// From HTTP to HTTPS
	config.Set("protocol", "https")
	router := newRouter()

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "http://localhost:443/test?param=1", nil))
	result := recorder.Result()
	body, err := ioutil.ReadAll(result.Body)
	suite.Nil(err)

	suite.Equal(http.StatusPermanentRedirect, result.StatusCode)
	suite.Equal("<a href=\"https://127.0.0.1:1236/test?param=1\">Permanent Redirect</a>.\n\n", string(body))

	// From HTTPS to HTTP
	config.Set("protocol", "http")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "https://localhost:80/test?param=1", nil))
	result = recorder.Result()
	body, err = ioutil.ReadAll(result.Body)
	suite.Nil(err)

	suite.Equal(http.StatusPermanentRedirect, result.StatusCode)
	suite.Equal("<a href=\"http://127.0.0.1:1235/test?param=1\">Permanent Redirect</a>.\n\n", string(body))

	// Only URI
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest("GET", "/test?param=1", nil))
	result = recorder.Result()
	body, err = ioutil.ReadAll(result.Body)
	suite.Nil(err)

	suite.Equal(http.StatusNotFound, result.StatusCode)
	suite.Equal("{\"error\":\"Not Found\"}\n", string(body))
}

func (suite *RouterTestSuite) TestConflictingRoutes() {
	// Test subrouter has priority over routes
	handler := func(response *Response, request *Request) {
		response.Status(200)
	}
	router := newRouter()

	subrouter := router.Subrouter("/product")
	routeSub := subrouter.Route("GET", "/{id:[0-9]+}", handler, nil)

	router.Route("GET", "/product/{id:[0-9]+}", handler, nil)

	req := httptest.NewRequest("GET", "/product/2", nil)
	match := routeMatch{currentPath: req.URL.Path}
	router.match(req, &match)

	suite.Equal(routeSub, match.route)

	// Test when route not in subrouter but first segment matches
	// Should not match
	router.Route("GET", "/product/test", handler, nil)

	req = httptest.NewRequest("GET", "/product/test", nil)
	match = routeMatch{currentPath: req.URL.Path}
	router.match(req, &match)

	suite.Equal(notFoundRoute, match.route)
}

func TestRouterTestSuite(t *testing.T) {
	RunTest(t, new(RouterTestSuite))
}
