package goyave

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/cors"
	"goyave.dev/goyave/v5/util/fsutil"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
)

type testStatusHandler struct {
	Component
	override Handler
}

func (h *testStatusHandler) Handle(response *Response, request *Request) {
	if h.override != nil {
		h.override(response, request)
		return
	}
	message := map[string]string{
		"status": http.StatusText(response.GetStatus()),
	}
	response.JSON(response.GetStatus(), message)
}

type extraMiddlewareOrder struct{}

type testMiddleware struct {
	Component
	key string
}

func (m *testMiddleware) Handle(next Handler) Handler {
	return func(r *Response, req *Request) {
		var slice []string
		if s, ok := req.Extra[extraMiddlewareOrder{}]; !ok {
			slice = []string{}
		} else {
			slice = s.([]string)
		}
		slice = append(slice, m.key)
		req.Extra[extraMiddlewareOrder{}] = slice
		next(r, req)
	}
}

func prepareRouterTest() *Router {
	server, err := New(Options{Config: config.LoadDefault()})
	if err != nil {
		panic(err)
	}
	return NewRouter(server)
}

type testController struct {
	Component
	registered bool
}

func (c *testController) RegisterRoutes(_ *Router) {
	c.registered = true
}

func TestRouter(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		router := prepareRouterTest()
		if !assert.NotNil(t, router) {
			return
		}
		assert.NotNil(t, router.server)
		assert.Nil(t, router.parent)
		assert.Empty(t, router.prefix)
		assert.Len(t, router.statusHandlers, 41)
		assert.NotNil(t, router.namedRoutes)
		assert.NotNil(t, router.Meta)

		recoveryMiddleware := findMiddleware[*recoveryMiddleware](router.globalMiddleware.middleware)
		langMiddleware := findMiddleware[*languageMiddleware](router.globalMiddleware.middleware)
		if assert.NotNil(t, recoveryMiddleware) {
			assert.Equal(t, router.server, recoveryMiddleware.server)
		}
		if assert.NotNil(t, langMiddleware) {
			assert.Equal(t, router.server, langMiddleware.server)
		}
	})

	t.Run("ClearRegexCache", func(t *testing.T) {
		router := prepareRouterTest()
		subrouter := router.Subrouter("/subrouter")

		assert.NotNil(t, router.regexCache)

		router.ClearRegexCache()
		assert.Nil(t, router.regexCache)
		assert.Nil(t, subrouter.regexCache)
	})

	t.Run("Accessors", func(t *testing.T) {
		router := prepareRouterTest()
		subrouter := router.Subrouter("/subrouter")
		route := subrouter.Get("/route", func(_ *Response, _ *Request) {}).Name("route-name")

		assert.Equal(t, router, subrouter.GetParent())
		assert.Equal(t, []*Route{route}, subrouter.GetRoutes())
		assert.Equal(t, []*Router{subrouter}, router.GetSubrouters())
		assert.Equal(t, route, router.GetRoute("route-name"))
		assert.Equal(t, route, subrouter.GetRoute("route-name"))
	})

	t.Run("Meta", func(t *testing.T) {
		router := prepareRouterTest()
		router.Meta["parent-meta"] = "parent-value"
		subrouter := router.Subrouter("/subrouter")
		subrouter.SetMeta("meta-key", "meta-value")
		assert.Equal(t, map[string]any{"meta-key": "meta-value"}, subrouter.Meta)

		val, ok := subrouter.LookupMeta("meta-key")
		assert.Equal(t, "meta-value", val)
		assert.True(t, ok)

		val, ok = subrouter.LookupMeta("parent-meta")
		assert.Equal(t, "parent-value", val)
		assert.True(t, ok)

		val, ok = subrouter.LookupMeta("nonexistent")
		assert.Nil(t, val)
		assert.False(t, ok)

		subrouter.RemoveMeta("meta-key")
		assert.Empty(t, subrouter.Meta)

		subrouter.SetMeta("parent-meta", "override")
		val, ok = subrouter.LookupMeta("parent-meta")
		assert.Equal(t, "override", val)
		assert.True(t, ok)
	})

	t.Run("GlobalMiddleware", func(t *testing.T) {
		router := prepareRouterTest()
		router.GlobalMiddleware(&corsMiddleware{}, &validateRequestMiddleware{})
		assert.Len(t, router.globalMiddleware.middleware, 4)
		for _, m := range router.globalMiddleware.middleware {
			assert.NotNil(t, m.Server())
		}
	})

	t.Run("Middleware", func(t *testing.T) {
		router := prepareRouterTest()
		router.Middleware(&corsMiddleware{}, &validateRequestMiddleware{})
		assert.Len(t, router.middleware, 2)
		for _, m := range router.middleware {
			assert.NotNil(t, m.Server())
		}
	})

	t.Run("CORS", func(t *testing.T) {
		router := prepareRouterTest()
		opts := cors.Default()

		router.CORS(opts)

		assert.Equal(t, opts, router.Meta[MetaCORS])
		assert.True(t, hasMiddleware[*corsMiddleware](router.globalMiddleware.middleware))

		// OPTIONS method is added to routes if the router has CORS
		route := router.Get("/route", func(_ *Response, _ *Request) {})
		assert.Equal(t, []string{http.MethodGet, http.MethodOptions, http.MethodHead}, route.methods)

		// OPTIONS method is added to routes if one of the parent routes has CORS
		route = router.Subrouter("/subrouter").Get("/route", func(_ *Response, _ *Request) {})
		assert.Equal(t, []string{http.MethodGet, http.MethodOptions, http.MethodHead}, route.methods)

		// Disable in subrouter
		subrouter := router.Subrouter("/subrouter2")
		subrouter.CORS(nil)
		route = subrouter.Get("/route-2", func(_ *Response, _ *Request) {})
		assert.Equal(t, []string{http.MethodGet, http.MethodHead}, route.methods)

		// Disable
		router.CORS(nil)
		assert.Contains(t, router.Meta, MetaCORS)
		assert.Nil(t, router.Meta[MetaCORS])
		route = router.Get("/route-2", func(_ *Response, _ *Request) {})
		assert.Equal(t, []string{http.MethodGet, http.MethodHead}, route.methods)
	})

	t.Run("StatusHandler", func(t *testing.T) {
		router := prepareRouterTest()

		statusHandler := &testStatusHandler{}
		router.StatusHandler(statusHandler, 1, 2, 3)

		assert.Equal(t, router.server, statusHandler.server)
		assert.Equal(t, statusHandler, router.statusHandlers[1])
		assert.Equal(t, statusHandler, router.statusHandlers[2])
		assert.Equal(t, statusHandler, router.statusHandlers[3])
	})

	t.Run("Subrouter", func(t *testing.T) {
		router := prepareRouterTest()
		router.Get("/named", nil).Name("route-name")
		subrouter := router.Subrouter("/subrouter")

		assert.Equal(t, router.server, subrouter.server)
		assert.Equal(t, router, subrouter.parent)
		assert.Equal(t, "/subrouter", subrouter.prefix)
		assert.Equal(t, router.statusHandlers, subrouter.statusHandlers)
		assert.NotSame(t, router.statusHandlers, subrouter.statusHandlers)
		assert.Equal(t, router.namedRoutes, subrouter.namedRoutes)
		assert.Equal(t, router.globalMiddleware, subrouter.globalMiddleware)
		assert.Equal(t, router.regexCache, subrouter.regexCache)
		assert.NotNil(t, subrouter.Meta)
		assert.Empty(t, subrouter.Meta)
		assert.Equal(t, []*Router{subrouter}, router.subrouters)
		assert.NotNil(t, subrouter.regex)

		slash := router.Subrouter("/")
		group := router.Group()
		assert.Empty(t, slash.prefix)
		assert.Equal(t, slash, group)
	})

	t.Run("Route", func(t *testing.T) {
		router := prepareRouterTest()

		route := router.Route([]string{http.MethodPost, http.MethodPut}, "/uri/{param}", func(_ *Response, _ *Request) {})
		assert.Empty(t, route.name)
		assert.Equal(t, "/uri/{param}", route.uri)
		assert.Equal(t, []string{http.MethodPost, http.MethodPut}, route.methods)
		assert.Equal(t, router, route.parent)
		assert.NotNil(t, route.handler)
		assert.NotNil(t, route.Meta)
		assert.NotNil(t, route.regex)

		t.Run("HEAD_added_on_GET_routes", func(t *testing.T) {
			route := router.Route([]string{http.MethodGet}, "/uri", func(_ *Response, _ *Request) {})
			assert.Equal(t, []string{http.MethodGet, http.MethodHead}, route.methods)
		})

		t.Run("trim_slash", func(t *testing.T) {
			// Not trimmed because no parent
			route := router.Route([]string{http.MethodGet}, "/", func(_ *Response, _ *Request) {})
			assert.Equal(t, "/", route.uri)

			route = router.Subrouter("/subrouter").Route([]string{http.MethodGet}, "/", func(_ *Response, _ *Request) {})
			assert.Equal(t, "", route.uri)
		})
	})

	t.Run("Get", func(t *testing.T) {
		router := prepareRouterTest()
		route := router.Get("/uri", func(_ *Response, _ *Request) {})
		assert.Equal(t, []string{http.MethodGet, http.MethodHead}, route.methods)
	})

	t.Run("Post", func(t *testing.T) {
		router := prepareRouterTest()
		route := router.Post("/uri", func(_ *Response, _ *Request) {})
		assert.Equal(t, []string{http.MethodPost}, route.methods)
	})

	t.Run("Put", func(t *testing.T) {
		router := prepareRouterTest()
		route := router.Put("/uri", func(_ *Response, _ *Request) {})
		assert.Equal(t, []string{http.MethodPut}, route.methods)
	})

	t.Run("Patch", func(t *testing.T) {
		router := prepareRouterTest()
		route := router.Patch("/uri", func(_ *Response, _ *Request) {})
		assert.Equal(t, []string{http.MethodPatch}, route.methods)
	})

	t.Run("Delete", func(t *testing.T) {
		router := prepareRouterTest()
		route := router.Delete("/uri", func(_ *Response, _ *Request) {})
		assert.Equal(t, []string{http.MethodDelete}, route.methods)
	})

	t.Run("Options", func(t *testing.T) {
		router := prepareRouterTest()
		route := router.Options("/uri", func(_ *Response, _ *Request) {})
		assert.Equal(t, []string{http.MethodOptions}, route.methods)
	})

	t.Run("Static", func(t *testing.T) {
		router := prepareRouterTest()
		f, err := fs.Sub(&osfs.FS{}, "resources")
		require.NoError(t, err)
		route := router.Static(fsutil.NewEmbed(f.(fs.ReadDirFS)), "/uri", false)
		assert.Equal(t, []string{http.MethodGet, http.MethodHead}, route.methods)
		assert.Equal(t, []string{"resource"}, route.parameters)
		assert.Equal(t, "/uri{resource:.*}", route.uri)
	})

	t.Run("Controller", func(t *testing.T) {
		router := prepareRouterTest()
		ctrl := &testController{}
		router.Controller(ctrl)
		assert.Equal(t, router.server, ctrl.server)
		assert.True(t, ctrl.registered)
	})

	t.Run("ServeHTTP", func(t *testing.T) {
		router := prepareRouterTest()
		router.server.config.Set("server.proxy.host", "proxy.io")
		router.server.config.Set("server.proxy.protocol", "http")
		router.server.config.Set("server.proxy.port", 80)
		router.server.config.Set("server.proxy.base", "/base")

		var route *Route
		route = router.Get("/route/{param}", func(r *Response, req *Request) {
			assert.Equal(t, map[string]string{"param": "value"}, req.RouteParams)
			assert.Equal(t, route, req.Route)
			assert.False(t, req.Now.IsZero())
			r.String(http.StatusOK, "hello world")
		})
		router.Put("/empty", func(_ *Response, _ *Request) {})
		router.Get("/forbidden", func(r *Response, _ *Request) {
			r.Status(http.StatusForbidden)
		})

		router.Subrouter("/noparam").Get("", func(r *Response, req *Request) {
			assert.Equal(t, map[string]string{}, req.RouteParams)
			r.Status(http.StatusOK)
		})

		subrouter := router.Subrouter("/subrouter/{param}")
		subrouter.Get("/subroute", func(r *Response, req *Request) {
			assert.Equal(t, map[string]string{"param": "value"}, req.RouteParams)
			r.Status(http.StatusOK)
		})
		subrouter.Get("/subroute/{name}", func(r *Response, req *Request) {
			assert.Equal(t, map[string]string{"param": "value", "name": "johndoe"}, req.RouteParams)
			r.Status(http.StatusOK)
		})

		router.Middleware(&testMiddleware{key: "router"})
		router.GlobalMiddleware(&testMiddleware{key: "global"})
		router.Get("/middleware", func(r *Response, req *Request) {
			assert.Equal(t, []string{"global", "router", "route"}, req.Extra[extraMiddlewareOrder{}])
			r.Status(http.StatusOK)
		}).Middleware(&testMiddleware{key: "route"})

		statusHandlerSubrouter := router.Subrouter("/statushandler")
		statusHandlerSubrouter.StatusHandler(&testStatusHandler{
			override: func(response *Response, _ *Request) {
				response.String(response.GetStatus(), "Override Bad Request")
			},
		}, http.StatusBadRequest)
		statusHandlerSubrouter.Get("/", func(r *Response, _ *Request) {
			r.Status(http.StatusBadRequest)
		})

		cases := []struct {
			desc           string
			requestMethod  string
			requestURL     string
			expectedBody   string
			expectedStatus int
		}{
			{
				desc:           "simple_param",
				requestMethod:  http.MethodGet,
				requestURL:     "/route/value",
				expectedStatus: http.StatusOK,
				expectedBody:   "hello world",
			},
			{
				desc:           "multiple_param",
				requestMethod:  http.MethodGet,
				requestURL:     "/subrouter/value/subroute/johndoe",
				expectedStatus: http.StatusOK,
				expectedBody:   "",
			},
			{
				desc:           "no_param_in_leaf",
				requestMethod:  http.MethodGet,
				requestURL:     "/subrouter/value/subroute",
				expectedStatus: http.StatusOK,
				expectedBody:   "",
			},
			{
				desc:           "no_param",
				requestMethod:  http.MethodGet,
				requestURL:     "/noparam",
				expectedStatus: http.StatusOK,
				expectedBody:   "",
			},
			{
				desc:           "protocol_rediect",
				requestMethod:  http.MethodGet,
				requestURL:     "https://127.0.0.1:8080/route/value?query=abc",
				expectedStatus: http.StatusPermanentRedirect,
				expectedBody:   "<a href=\"http://proxy.io/base/route/value?query=abc\">Permanent Redirect</a>.\n\n",
			},
			{
				desc:           "empty_response",
				requestMethod:  http.MethodPut,
				requestURL:     "/empty",
				expectedStatus: http.StatusNoContent,
				expectedBody:   "",
			},
			{
				desc:           "status_handler",
				requestMethod:  http.MethodGet,
				requestURL:     "/forbidden",
				expectedStatus: http.StatusForbidden,
				expectedBody:   "{\"error\":\"Forbidden\"}\n",
			},
			{
				desc:           "not_found",
				requestMethod:  http.MethodGet,
				requestURL:     "/not_found",
				expectedStatus: http.StatusNotFound,
				expectedBody:   "{\"error\":\"Not Found\"}\n",
			},
			{
				desc:           "method_not_allowed",
				requestMethod:  http.MethodPatch,
				requestURL:     "/empty",
				expectedStatus: http.StatusMethodNotAllowed,
				expectedBody:   "{\"error\":\"Method Not Allowed\"}\n",
			},
			{
				desc:           "middleware_order",
				requestMethod:  http.MethodGet,
				requestURL:     "/middleware",
				expectedStatus: http.StatusOK,
				expectedBody:   "",
			},
			{
				desc:           "subrouter_status_handler",
				requestMethod:  http.MethodGet,
				requestURL:     "/statushandler",
				expectedStatus: http.StatusBadRequest,
				expectedBody:   "Override Bad Request",
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				recorder := httptest.NewRecorder()
				rawRequest := httptest.NewRequest(c.requestMethod, c.requestURL, nil)
				router.ServeHTTP(recorder, rawRequest)

				res := recorder.Result()

				assert.Equal(t, c.expectedStatus, res.StatusCode)

				body, err := io.ReadAll(res.Body)
				assert.NoError(t, res.Body.Close())
				require.NoError(t, err)
				assert.Equal(t, c.expectedBody, string(body))
			})
		}
	})

	t.Run("match", func(t *testing.T) {
		router := prepareRouterTest()

		router.Get("/", nil).Name("root")
		router.Get("/first-level", nil).Name("first-level")

		categories := router.Subrouter("/categories")
		categories.Get("/", nil).Name("categories.index")
		category := categories.Subrouter("/{categoryId:[0-9]+}")
		category.Get("/", nil).Name("categories.show")
		category.Get("/inventory", nil).Name("categories.inventory")

		// Subrouter has priority over route, this one will never match
		router.Get("/categories/{categoryId:[0-9]+}", nil).Name("never-match")

		// The first segment in the URI matches the subrouter, so this one will never match neither
		router.Get("/categories/test", nil).Name("never-match-first-segment")

		products := category.Subrouter("/products")
		products.Get("/", nil).Name("products.index")
		products.Post("/", nil).Name("products.create")
		products.Get("/{id:[0-9]+}", nil).Name("products.show")

		// Route groups, we should be able to match /profile even with the admins
		// subrouter (because it has an empty prefix)
		users := router.Subrouter("/users")
		admins := users.Group()
		admins.Get("/manage", nil).Name("users.admins.manage")
		admins.Post("/", nil).Name("users.admins.create")
		viewers := users.Group()
		viewers.Get("/profile", nil).Name("users.viewers.profile")
		viewers.Get("/", nil).Name("users.viewers.show")
		users.Put("/", nil).Name("users.update")

		// Conflicting subrouters
		conflict := router.Subrouter("/conflict")
		conflict.Get("/", nil).Name("conflict.root")
		conflict.Get("/child", nil).Name("conflict.child")
		conflict2 := router.Subrouter("/conflict-2")
		conflict2.Get("/", nil).Name("conflict-2.root")
		conflict2.Get("/child", nil).Name("conflict-2.child")

		// Multiple segments in subrouter path
		subrouter := router.Subrouter("/subrouter/{param}")
		subrouter.Get("/", nil).Name("multiple-segments.subroute.index")
		subrouter.Get("/subroute", nil).Name("multiple-segments.subroute.show")
		subrouter.Get("/subroute/{name}", nil).Name("multiple-segments.subroute.name")

		cases := []struct {
			path          string
			method        string
			expectedRoute string
		}{
			{path: "/", method: http.MethodGet, expectedRoute: "root"},
			{path: "/", method: http.MethodPost, expectedRoute: RouteMethodNotAllowed},
			{path: "/first-level", method: http.MethodGet, expectedRoute: "first-level"},
			{path: "/first-level/", method: http.MethodGet, expectedRoute: RouteNotFound}, // Trailing slash
			{path: "/first-level", method: http.MethodPost, expectedRoute: RouteMethodNotAllowed},
			{path: "/categories", method: http.MethodGet, expectedRoute: "categories.index"},
			{path: "/categories/", method: http.MethodGet, expectedRoute: RouteNotFound}, // Trailing slash
			{path: "/categories/123", method: http.MethodGet, expectedRoute: "categories.show"},
			{path: "/categories/123/inventory", method: http.MethodGet, expectedRoute: "categories.inventory"},
			{path: "/categories/test", method: http.MethodGet, expectedRoute: RouteNotFound},
			{path: "/categories/123/products", method: http.MethodGet, expectedRoute: "products.index"},
			{path: "/categories/123/products", method: http.MethodPost, expectedRoute: "products.create"},
			{path: "/categories/123/products/1234567890", method: http.MethodGet, expectedRoute: "products.show"},
			{path: "/users/manage", method: http.MethodGet, expectedRoute: "users.admins.manage"},
			{path: "/users/profile", method: http.MethodGet, expectedRoute: "users.viewers.profile"},
			{path: "/users", method: http.MethodGet, expectedRoute: "users.viewers.show"}, // Method not allowed on users.admins.create
			{path: "/users", method: http.MethodPut, expectedRoute: "users.update"},
			{path: "/conflict", method: http.MethodGet, expectedRoute: "conflict.root"},
			{path: "/conflict/", method: http.MethodGet, expectedRoute: RouteNotFound},
			{path: "/conflict/child", method: http.MethodGet, expectedRoute: "conflict.child"},
			{path: "/conflict-2", method: http.MethodGet, expectedRoute: "conflict-2.root"},
			{path: "/conflict-2/", method: http.MethodGet, expectedRoute: RouteNotFound},
			{path: "/conflict-2/child", method: http.MethodGet, expectedRoute: "conflict-2.child"},
			{path: "/categories/123/not-a-route", method: http.MethodGet, expectedRoute: RouteNotFound},
			{path: "/categories/123/not-a-route/", method: http.MethodGet, expectedRoute: RouteNotFound},
			{path: "/subrouter/value", method: http.MethodGet, expectedRoute: "multiple-segments.subroute.index"},
			{path: "/subrouter/value/", method: http.MethodGet, expectedRoute: RouteNotFound},
			{path: "/subrouter/value/subroute", method: http.MethodGet, expectedRoute: "multiple-segments.subroute.show"},
			{path: "/subrouter/value/subroute/", method: http.MethodGet, expectedRoute: RouteNotFound},
			{path: "/subrouter/value/subroute/johndoe", method: http.MethodGet, expectedRoute: "multiple-segments.subroute.name"},
		}

		for _, c := range cases {
			t.Run(fmt.Sprintf("%s_%s", c.method, strings.ReplaceAll(c.path, "/", "_")), func(t *testing.T) {
				match := routeMatch{currentPath: c.path}
				router.match(c.method, &match)
				assert.Equal(t, c.expectedRoute, match.route.name)
			})
		}
	})
}
