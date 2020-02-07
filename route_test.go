package goyave

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RouteTestSuite struct {
	suite.Suite
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
		methods: []string{"GET", "POST"},
	}

	suite.Equal("route-name", route.GetName())
	suite.Equal("/product/{id:[0-9+]}", route.GetURI())
	suite.Equal([]string{"GET", "POST"}, route.GetMethods())
}

func TestRouteTestSuite(t *testing.T) {
	suite.Run(t, new(RouteTestSuite))
}
