package goyave

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"goyave.dev/goyave/v3/validation"
)

type RouteTestSuite struct {
	TestSuite
}

func (suite *RouteTestSuite) SetupTest() {
	regexCache = make(map[string]*regexp.Regexp, 5)
}

func (suite *RouteTestSuite) TearDownTest() {
	regexCache = nil
}

func (suite *RouteTestSuite) TestNewRoute() {
	route := newRoute(func(resp *Response, r *Request) {})
	suite.NotNil(route)
	suite.NotNil(route.handler)
}

func (suite *RouteTestSuite) TestMakeParameters() {
	route := newRoute(func(resp *Response, r *Request) {})
	route.compileParameters("/product/{id:[0-9]+}", true)
	suite.Equal([]string{"id"}, route.parameters)
	suite.NotNil(route.regex)
	suite.True(route.regex.MatchString("/product/666"))
	suite.False(route.regex.MatchString("/product/"))
	suite.False(route.regex.MatchString("/product/qwerty"))
}

func (suite *RouteTestSuite) TestMatch() {
	handler := func(resp *Response, r *Request) {
		resp.String(http.StatusOK, "Success")
	}
	route := &Route{
		name:            "test-route",
		uri:             "/product/{id:[0-9]+}",
		methods:         []string{"GET", "POST"},
		parent:          nil,
		handler:         handler,
		validationRules: nil,
	}
	route.compileParameters(route.uri, true)

	rawRequest := httptest.NewRequest("GET", "/product/33", nil)
	match := routeMatch{currentPath: rawRequest.URL.Path}
	suite.True(route.match(rawRequest, &match))
	suite.Equal("33", match.parameters["id"])

	rawRequest = httptest.NewRequest("POST", "/product/33", nil)
	match = routeMatch{currentPath: rawRequest.URL.Path}
	suite.True(route.match(rawRequest, &match))
	suite.Equal("33", match.parameters["id"])

	rawRequest = httptest.NewRequest("PUT", "/product/33", nil)
	match = routeMatch{currentPath: rawRequest.URL.Path}
	suite.False(route.match(rawRequest, &match))
	suite.Equal(errMatchMethodNotAllowed, match.err)

	// Test error has not been overridden
	rawRequest = httptest.NewRequest("PUT", "/product/test", nil)
	suite.False(route.match(rawRequest, &match))
	suite.Equal(errMatchMethodNotAllowed, match.err)

	match = routeMatch{currentPath: rawRequest.URL.Path}
	suite.False(route.match(rawRequest, &match))
	suite.Equal(errMatchNotFound, match.err)

	route = &Route{
		name:            "test-route",
		uri:             "/product/{id:[0-9]+}/{name}",
		methods:         []string{"GET"},
		parent:          nil,
		handler:         handler,
		validationRules: nil,
	}
	route.compileParameters(route.uri, true)
	rawRequest = httptest.NewRequest("GET", "/product/666/test", nil)
	match = routeMatch{currentPath: rawRequest.URL.Path}
	suite.True(route.match(rawRequest, &match))
	suite.Equal("666", match.parameters["id"])
	suite.Equal("test", match.parameters["name"])

	route = &Route{
		name:            "test-route",
		uri:             "/categories/{category}/{sort:(?:asc|desc|new)}",
		methods:         []string{"GET"},
		parent:          nil,
		handler:         handler,
		validationRules: nil,
	}
	route.compileParameters(route.uri, true)
	rawRequest = httptest.NewRequest("GET", "/categories/lawn-mower/asc", nil)
	match = routeMatch{currentPath: rawRequest.URL.Path}
	suite.True(route.match(rawRequest, &match))
	suite.Equal("lawn-mower", match.parameters["category"])
	suite.Equal("asc", match.parameters["sort"])

	rawRequest = httptest.NewRequest("GET", "/categories/lawn-mower/notsort", nil)
	match = routeMatch{currentPath: rawRequest.URL.Path}
	suite.False(route.match(rawRequest, &match))
}

func (suite *RouteTestSuite) TestAccessors() {
	route := &Route{
		name:    "route-name",
		uri:     "/product/{id:[0-9+]}",
		parent:  newRouter(),
		methods: []string{"GET", "POST"},
	}

	suite.Equal("route-name", route.GetName())

	suite.Panics(func() {
		route.Name("new-name") // Cannot re-set name
	})

	route = &Route{
		name:    "",
		uri:     "/product/{id:[0-9+]}",
		parent:  newRouter(),
		methods: []string{"GET", "POST"},
	}
	route.Name("new-name")
	suite.Equal("new-name", route.GetName())

	suite.Equal("/product/{id:[0-9+]}", route.GetURI())
	suite.Equal([]string{"GET", "POST"}, route.GetMethods())
}

func (suite *RouteTestSuite) TestGetFullURI() {
	router := newRouter().Subrouter("/product").Subrouter("/{id:[0-9+]}")
	route := router.Route("GET|POST", "/{name}/accessories", func(resp *Response, r *Request) {}).Name("route-name")

	suite.Equal("/product/{id:[0-9+]}/{name}/accessories", route.GetFullURI())
}

func (suite *RouteTestSuite) TestBuildURI() {
	route := &Route{
		name:    "route-name",
		uri:     "/product/{id:[0-9+]}",
		methods: []string{"GET", "POST"},
	}
	route.compileParameters(route.uri, true)
	suite.Equal("/product/42", route.BuildURI("42"))

	suite.Panics(func() {
		route.BuildURI()
	})
	suite.Panics(func() {
		route.BuildURI("42", "more")
	})

	route = &Route{
		name:    "route-name",
		uri:     "/product/{id:[0-9+]}/{name}/accessories",
		methods: []string{"GET", "POST"},
	}
	route.compileParameters(route.uri, true)
	suite.Equal("/product/42/screwdriver/accessories", route.BuildURI("42", "screwdriver"))

	router := newRouter().Subrouter("/product").Subrouter("/{id:[0-9+]}")
	route = router.Route("GET|POST", "/{name}/accessories", func(resp *Response, r *Request) {}).Name("route-name")

	suite.Equal("/product/42/screwdriver/accessories", route.BuildURI("42", "screwdriver"))
}

func (suite *RouteTestSuite) TestBuildURL() {
	route := &Route{
		name:    "route-name",
		uri:     "/product/{id:[0-9+]}",
		methods: []string{"GET", "POST"},
	}
	route.compileParameters(route.uri, true)
	suite.Equal("http://127.0.0.1:1235/product/42", route.BuildURL("42"))

	suite.Panics(func() {
		route.BuildURL()
	})
	suite.Panics(func() {
		route.BuildURL("42", "more")
	})

	route = &Route{
		name:    "route-name",
		uri:     "/product/{id:[0-9+]}/{name}/accessories",
		methods: []string{"GET", "POST"},
	}
	route.compileParameters(route.uri, true)
	suite.Equal("http://127.0.0.1:1235/product/42/screwdriver/accessories", route.BuildURL("42", "screwdriver"))

	router := newRouter().Subrouter("/product").Subrouter("/{id:[0-9+]}")
	route = router.Route("GET|POST", "/{name}/accessories", func(resp *Response, r *Request) {}).Name("route-name")

	suite.Equal("http://127.0.0.1:1235/product/42/screwdriver/accessories", route.BuildURL("42", "screwdriver"))
}

func (suite *RouteTestSuite) TestValidate() {
	route := &Route{
		name:    "route-name",
		uri:     "/product/{id:[0-9+]}",
		methods: []string{"GET", "POST"},
	}
	rules := &validation.Rules{
		Fields: map[string]*validation.Field{
			"field": {
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "string"},
				},
			},
		},
	}
	route.Validate(rules)
	suite.Equal(rules, route.validationRules)
}

func (suite *RouteTestSuite) TestMiddleware() {
	route := &Route{
		name:    "route-name",
		uri:     "/product/{id:[0-9+]}",
		methods: []string{"GET", "POST"},
	}
	middelware := func(next Handler) Handler {
		return func(response *Response, r *Request) {}
	}
	middelware2 := func(next Handler) Handler {
		return func(response *Response, r *Request) {}
	}
	route.Middleware(middelware, middelware2)
	suite.Len(route.middleware, 2)
}

func (suite *RouteTestSuite) TestGetHandler() {
	executed := false
	route := &Route{
		name:    "route-name",
		uri:     "/product/{id:[0-9+]}",
		methods: []string{"GET", "POST"},
		handler: func(resp *Response, r *Request) {
			executed = true
		},
	}
	handler := route.GetHandler()
	suite.NotNil(handler)
	handler(nil, nil)
	suite.True(executed)
}

func (suite *RouteTestSuite) TestGetValidationRules() {
	route := &Route{
		name:    "route-name",
		uri:     "/product/{id:[0-9+]}",
		methods: []string{"GET", "POST"},
	}
	rules := &validation.Rules{
		Fields: validation.FieldMap{
			"email": {
				Rules: []*validation.Rule{
					{Name: "required"},
					{Name: "string"},
					{Name: "between", Params: []string{"3", "125"}},
					{Name: "email"},
				},
			},
		},
	}
	route.Validate(rules)

	suite.Same(rules, route.GetValidationRules())
}

func TestRouteTestSuite(t *testing.T) {
	RunTest(t, new(RouteTestSuite))
}
