package goyave

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type RouteTestSuite struct {
	TestSuite
}

func (suite *RouteTestSuite) TestNewRoute() {
	route := newRoute(func(resp *Response, r *Request) {})
	suite.NotNil(route)
	suite.NotNil(route.handler)
}

func (suite *RouteTestSuite) TestMakeParameters() {
	route := newRoute(func(resp *Response, r *Request) {})
	route.compileParameters("/product/{id:[0-9]+}")
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
		uri:             "/product/{id:[0-9]+}", // TODO use partial route only for optimization
		methods:         []string{"GET", "POST"},
		parent:          nil,
		handler:         handler,
		validationRules: nil,
	}
	route.compileParameters(route.uri)

	rawRequest := httptest.NewRequest("GET", "/product/33", nil)
	match := routeMatch{}
	suite.True(route.match(rawRequest, &match))
	suite.Equal("33", match.parameters["id"])

	rawRequest = httptest.NewRequest("POST", "/product/33", nil)
	match = routeMatch{}
	suite.True(route.match(rawRequest, &match))
	suite.Equal("33", match.parameters["id"])

	rawRequest = httptest.NewRequest("PUT", "/product/33", nil)
	match = routeMatch{}
	suite.False(route.match(rawRequest, &match))
	suite.Equal(errMatchMethodNotAllowed, match.err)

	// Test error has not been overridden
	rawRequest = httptest.NewRequest("PUT", "/product/test", nil)
	suite.False(route.match(rawRequest, &match))
	suite.Equal(errMatchMethodNotAllowed, match.err)

	match = routeMatch{}
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
	route.compileParameters(route.uri)
	rawRequest = httptest.NewRequest("GET", "/product/666/test", nil)
	match = routeMatch{}
	suite.True(route.match(rawRequest, &match))
	suite.Equal("666", match.parameters["id"])
	suite.Equal("test", match.parameters["name"])
}

func (suite *RouteTestSuite) TestAccessors() {
	route := &Route{
		name:    "route-name",
		uri:     "/product/{id:[0-9+]}",
		parent:  newRouter(),
		methods: []string{"GET", "POST"},
	}

	suite.Equal("route-name", route.GetName())

	route.Name("new-name")
	suite.Equal("new-name", route.GetName())

	suite.Equal("/product/{id:[0-9+]}", route.GetURI())
	suite.Equal([]string{"GET", "POST"}, route.GetMethods())
}

func (suite *RouteTestSuite) TestBuildURL() {
	route := &Route{
		name:    "route-name",
		uri:     "/product/{id:[0-9+]}",
		methods: []string{"GET", "POST"},
	}
	route.compileParameters(route.uri)
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
	route.compileParameters(route.uri)
	suite.Equal("http://127.0.0.1:1235/product/42/screwdriver/accessories", route.BuildURL("42", "screwdriver"))
}

func TestRouteTestSuite(t *testing.T) {
	RunTest(t, new(RouteTestSuite))
}
