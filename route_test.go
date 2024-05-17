package goyave

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/cors"
	"goyave.dev/goyave/v5/validation"
)

func prepareRouteTest() *Router {
	server, err := New(Options{Config: config.LoadDefault()})
	if err != nil {
		panic(err)
	}
	return NewRouter(server)
}

func routeTestValidationRules(_ *Request) validation.RuleSet {
	return validation.RuleSet{
		{Path: "field", Rules: validation.List{
			validation.Required(),
		}},
	}
}

func TestRoute(t *testing.T) {

	t.Run("Name", func(t *testing.T) {
		router := prepareRouteTest()
		route := &Route{parent: router}
		route.Name("route-name")
		assert.Equal(t, "route-name", route.name)
		assert.Equal(t, router.namedRoutes["route-name"], route)

		t.Run("already_set", func(t *testing.T) {
			router := prepareRouteTest()
			route := &Route{parent: router, name: "route-name"}
			assert.Panics(t, func() {
				route.Name("route-rename")
			})
		})

		t.Run("already_exists", func(t *testing.T) {
			router := prepareRouteTest()
			route := &Route{parent: router}
			route.Name("route-name")
			assert.Panics(t, func() {
				anotherRoute := &Route{parent: router}
				anotherRoute.Name("route-name")
			})
		})
	})

	t.Run("Meta", func(t *testing.T) {
		router := prepareRouteTest()
		router.Meta["parent-meta"] = "parent-value"
		route := &Route{parent: router, Meta: make(map[string]any)}
		route.SetMeta("meta-key", "meta-value")
		assert.Equal(t, map[string]any{"meta-key": "meta-value"}, route.Meta)

		val, ok := route.LookupMeta("meta-key")
		assert.Equal(t, "meta-value", val)
		assert.True(t, ok)

		val, ok = route.LookupMeta("parent-meta")
		assert.Equal(t, "parent-value", val)
		assert.True(t, ok)

		val, ok = route.LookupMeta("nonexistent")
		assert.Nil(t, val)
		assert.False(t, ok)

		route.RemoveMeta("meta-key")
		assert.Empty(t, route.Meta)

		// Shouldn't panic if there is no meta (usually the case in special
		// routes such as notFound and methodNotAllowed)
		route.parent = nil
		val, ok = route.LookupMeta("nonexistent")
		assert.Nil(t, val)
		assert.False(t, ok)
	})

	t.Run("ValidateBody", func(t *testing.T) {
		router := prepareRouteTest()
		route := &Route{
			parent: router,
			middlewareHolder: middlewareHolder{
				middleware: []Middleware{},
			},
		}

		route.ValidateBody(routeTestValidationRules)

		validationMiddleware := findMiddleware[*validateRequestMiddleware](route.middleware)
		if !assert.NotNil(t, validationMiddleware) {
			return
		}
		assert.NotNil(t, validationMiddleware.BodyRules)
		assert.Nil(t, validationMiddleware.QueryRules)

		// Replace body query
		route.ValidateBody(nil)
		assert.Nil(t, validationMiddleware.BodyRules)
		assert.Nil(t, validationMiddleware.QueryRules)
	})

	t.Run("ValidateQuery", func(t *testing.T) {
		router := prepareRouteTest()
		route := &Route{
			parent: router,
			middlewareHolder: middlewareHolder{
				middleware: []Middleware{},
			},
		}

		route.ValidateQuery(routeTestValidationRules)

		validationMiddleware := findMiddleware[*validateRequestMiddleware](route.middleware)
		if !assert.NotNil(t, validationMiddleware) {
			return
		}
		assert.NotNil(t, validationMiddleware.QueryRules)
		assert.Nil(t, validationMiddleware.BodyRules)

		// Replace query validation
		route.ValidateQuery(nil)
		assert.Nil(t, validationMiddleware.BodyRules)
		assert.Nil(t, validationMiddleware.QueryRules)
	})

	t.Run("CORS", func(t *testing.T) {
		router := prepareRouteTest()
		route := &Route{
			parent:  router,
			methods: []string{http.MethodGet},
			Meta:    make(map[string]any),
			middlewareHolder: middlewareHolder{
				middleware: []Middleware{},
			},
		}

		opts := cors.Default()
		route.CORS(opts)

		assert.Equal(t, opts, route.Meta[MetaCORS])
		middleware := findMiddleware[*corsMiddleware](router.globalMiddleware.middleware)
		if !assert.NotNil(t, middleware) {
			return
		}
		assert.Equal(t, []string{http.MethodGet, http.MethodOptions}, route.methods)

		// Disable
		route.CORS(nil)
		assert.Nil(t, route.Meta[MetaCORS])
		assert.Contains(t, route.Meta, MetaCORS)
		assert.Equal(t, []string{http.MethodGet}, route.methods)

		t.Run("don't_add_middleware_if_parent_has_it_already", func(t *testing.T) {
			router := prepareRouteTest()
			router.CORS(cors.Default())
			route := &Route{
				parent:  router,
				methods: []string{http.MethodGet},
				Meta:    make(map[string]any),
				middlewareHolder: middlewareHolder{
					middleware: []Middleware{},
				},
			}
			route.CORS(opts)
			middleware := findMiddleware[*corsMiddleware](route.middleware)
			assert.Nil(t, middleware)
		})
	})

	t.Run("Middleware", func(t *testing.T) {
		router := prepareRouteTest()
		route := &Route{
			parent: router,
			middlewareHolder: middlewareHolder{
				middleware: []Middleware{},
			},
		}

		route.Middleware(&recoveryMiddleware{}, &languageMiddleware{})
		assert.Len(t, route.middleware, 2)
		for _, m := range route.middleware {
			assert.NotNil(t, m.Server())
		}
	})

	t.Run("GetFullURIAndParameters", func(t *testing.T) {
		router := prepareRouteTest()
		subrouter := router.Subrouter("/product").Subrouter("/{id:[0-9+]}")
		route := subrouter.Route([]string{http.MethodGet}, "/{name}/accessories", nil)

		uri, params := route.GetFullURIAndParameters()
		assert.Equal(t, "/product/{id:[0-9+]}/{name}/accessories", uri)
		assert.Equal(t, []string{"id", "name"}, params)
	})

	t.Run("BuildURI", func(t *testing.T) {
		router := prepareRouteTest()
		subrouter := router.Subrouter("/product").Subrouter("/{id:[0-9+]}")
		route := subrouter.Route([]string{http.MethodGet}, "/{name}/accessories", nil)

		uri := route.BuildURI("123", "keyboard")
		assert.Equal(t, "/product/123/keyboard/accessories", uri)

		t.Run("invalid_param_count", func(t *testing.T) {
			assert.Panics(t, func() {
				route.BuildURI()
			})
		})
	})

	t.Run("BuildURL", func(t *testing.T) {
		router := prepareRouteTest()
		subrouter := router.Subrouter("/product").Subrouter("/{id:[0-9+]}")
		route := subrouter.Route([]string{http.MethodGet}, "/{name}/accessories", nil)

		uri := route.BuildURL("123", "keyboard")
		assert.Equal(t, "http://127.0.0.1:8080/product/123/keyboard/accessories", uri)
	})

	t.Run("BuildProxyURL", func(t *testing.T) {
		router := prepareRouteTest()
		subrouter := router.Subrouter("/product").Subrouter("/{id:[0-9+]}")
		route := subrouter.Route([]string{http.MethodGet}, "/{name}/accessories", nil)

		uri := route.BuildProxyURL("123", "keyboard")
		assert.Equal(t, "http://127.0.0.1:8080/product/123/keyboard/accessories", uri)
	})

	t.Run("GetFullURI", func(t *testing.T) {
		router := prepareRouteTest()
		subrouter := router.Subrouter("/product").Subrouter("/{id:[0-9+]}")
		route := subrouter.Route([]string{http.MethodGet}, "/{name}/accessories", nil)

		uri := route.GetFullURI()
		assert.Equal(t, "/product/{id:[0-9+]}/{name}/accessories", uri)
	})

	t.Run("Accessors", func(t *testing.T) {
		router := prepareRouteTest()
		route := router.Route([]string{http.MethodGet}, "/{name}/accessories", func(_ *Response, _ *Request) {}).Name("route-name")
		assert.Equal(t, "route-name", route.GetName())
		assert.Equal(t, []string{http.MethodGet, http.MethodHead}, route.GetMethods())
		assert.NotNil(t, route.GetHandler())
		assert.Equal(t, router, route.GetParent())
		assert.Equal(t, "/{name}/accessories", route.GetURI())
	})

	t.Run("Match", func(t *testing.T) {

		router := prepareRouteTest()
		route1 := router.Route([]string{http.MethodGet, http.MethodPost}, "/product/{id:[0-9]+}", func(_ *Response, _ *Request) {})
		route2 := router.Route([]string{http.MethodGet}, "/product/{id:[0-9]+}/{name}", func(_ *Response, _ *Request) {})
		route3 := router.Route([]string{http.MethodGet}, "/categories/{category}/{sort:(?:asc|desc|new)}", func(_ *Response, _ *Request) {})
		route4 := router.Route([]string{http.MethodGet}, "/product", func(_ *Response, _ *Request) {})

		cases := []struct {
			route              *Route
			expectedParameters map[string]string
			expectedError      error
			method             string
			uri                string
			expectedResult     bool
		}{
			{route: route1, method: http.MethodGet, uri: "/product/33", expectedResult: true, expectedParameters: map[string]string{"id": "33"}, expectedError: nil},
			{route: route1, method: http.MethodPost, uri: "/product/33", expectedResult: true, expectedParameters: map[string]string{"id": "33"}, expectedError: nil},
			{route: route1, method: http.MethodPut, uri: "/product/33", expectedResult: false, expectedParameters: nil, expectedError: errMatchMethodNotAllowed},
			{route: route1, method: http.MethodGet, uri: "/product/test", expectedResult: false, expectedParameters: nil, expectedError: errMatchNotFound},
			{route: route2, method: http.MethodGet, uri: "/product/666/test", expectedResult: true, expectedParameters: map[string]string{"id": "666", "name": "test"}, expectedError: nil},
			{route: route3, method: http.MethodGet, uri: "/categories/lawn-mower/asc", expectedResult: true, expectedParameters: map[string]string{"category": "lawn-mower", "sort": "asc"}, expectedError: nil},
			{route: route3, method: http.MethodGet, uri: "/categories/lawn-mower/notasc", expectedResult: false, expectedParameters: nil, expectedError: errMatchNotFound},
			{route: route4, method: http.MethodGet, uri: "/product", expectedResult: true, expectedParameters: nil, expectedError: nil},
		}

		for _, c := range cases {
			c := c
			t.Run(fmt.Sprintf("%s_%s", c.method, c.uri), func(t *testing.T) {
				match := routeMatch{currentPath: c.uri}
				assert.Equal(t, c.expectedResult, c.route.match(c.method, &match))
				assert.Equal(t, c.expectedParameters, match.parameters)
				assert.Equal(t, c.expectedError, match.err)
			})
		}

		t.Run("err_not_overridden", func(t *testing.T) {
			match := routeMatch{currentPath: "/product/33"}
			route1.match(http.MethodPut, &match)
			assert.Equal(t, errMatchMethodNotAllowed, match.err)

			match.currentPath = "/product/test"
			route1.match(http.MethodGet, &match)
			assert.Equal(t, errMatchMethodNotAllowed, match.err)
		})
	})
}
