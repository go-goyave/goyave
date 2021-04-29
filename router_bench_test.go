package goyave

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"goyave.dev/goyave/v3/validation"
)

type routerDefinition struct {
	prefix     string // Empty for main router
	middleware []Middleware
	routes     []*routeDefinition
	subrouters []*routerDefinition
}

type routeDefinition struct {
	handler Handler
	rules   validation.RuleSet
	uri     string
	methods string
	name    string
}

var handler Handler = func(response *Response, request *Request) {
	response.Status(200)
}

var sampleRouteDefinition *routerDefinition = &routerDefinition{
	prefix:     "",
	middleware: []Middleware{},
	routes: []*routeDefinition{
		{
			uri:     "/hello",
			methods: "GET",
			name:    "hello",
			handler: handler,
			rules:   nil,
		},
		{
			uri:     "/world",
			methods: "POST",
			name:    "post",
			handler: handler,
			rules:   nil,
		},
		{
			uri:     "/{param}",
			methods: "POST",
			name:    "param",
			handler: handler,
			rules:   nil,
		},
	},
	subrouters: []*routerDefinition{
		{
			prefix:     "/product",
			middleware: []Middleware{},
			routes: []*routeDefinition{
				{
					uri:     "/",
					methods: "GET",
					name:    "product.index",
					handler: handler,
					rules:   nil,
				},
				{
					uri:     "/",
					methods: "POST",
					name:    "product.store",
					handler: handler,
					rules:   nil,
				},
				{
					uri:     "/{id:[0-9]+}",
					methods: "GET",
					name:    "product.show",
					handler: handler,
					rules:   nil,
				},
				{
					uri:     "/{id:[0-9]+}",
					methods: "PUT|PATCH",
					name:    "product.update",
					handler: handler,
					rules:   nil,
				},
				{
					uri:     "/{id:[0-9]+}",
					methods: "DELETE",
					name:    "product.destroy",
					handler: handler,
					rules:   nil,
				},
			},
			subrouters: []*routerDefinition{},
		},
	},
}

var sampleRequests []*http.Request = []*http.Request{
	httptest.NewRequest("GET", "/", nil), // 404
	httptest.NewRequest("GET", "/hello", nil),
	httptest.NewRequest("POST", "/world", nil),
	httptest.NewRequest("POST", "/param", nil),
	httptest.NewRequest("GET", "/product", nil),
	httptest.NewRequest("POST", "/product", nil),
	httptest.NewRequest("GET", "/product/test", nil), // 404
	httptest.NewRequest("GET", "/product/1", nil),
	httptest.NewRequest("PUT", "/product/1", nil),
	httptest.NewRequest("DELETE", "/product/1", nil),
}

func registerAll(def *routerDefinition) *Router {
	main := NewRouter()
	registerRouter(main, def)
	return main
}

func registerRouter(router *Router, def *routerDefinition) {
	for _, subdef := range def.subrouters {
		subrouter := router.Subrouter(subdef.prefix)
		registerRouter(subrouter, subdef)
	}
	for _, routeDef := range def.routes {
		router.registerRoute(routeDef.methods, routeDef.uri, routeDef.handler).Validate(routeDef.rules).Name(routeDef.name)
	}
}

func BenchmarkRouteRegistration(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		registerAll(sampleRouteDefinition)
	}
}

func BenchmarkRootLevelNotFound(b *testing.B) {
	router := setupRouteBench(b)

	for n := 0; n < b.N; n++ {
		router.match(sampleRequests[0], &routeMatch{currentPath: sampleRequests[0].URL.Path})
	}
}

func BenchmarkRootLevelMatch(b *testing.B) {
	router := setupRouteBench(b)

	for n := 0; n < b.N; n++ {
		router.match(sampleRequests[1], &routeMatch{currentPath: sampleRequests[1].URL.Path})
	}
}

func BenchmarkRootLevelPostMatch(b *testing.B) {
	router := setupRouteBench(b)

	for n := 0; n < b.N; n++ {
		router.match(sampleRequests[2], &routeMatch{currentPath: sampleRequests[2].URL.Path})
	}
}

func BenchmarkRootLevelPostParamMatch(b *testing.B) {
	router := setupRouteBench(b)

	for n := 0; n < b.N; n++ {
		router.match(sampleRequests[3], &routeMatch{currentPath: sampleRequests[3].URL.Path})
	}
}

func BenchmarkSubrouterMatch(b *testing.B) {
	router := setupRouteBench(b)

	for n := 0; n < b.N; n++ {
		router.match(sampleRequests[4], &routeMatch{currentPath: sampleRequests[4].URL.Path})
	}
}

func BenchmarkSubrouterPostMatch(b *testing.B) {
	router := setupRouteBench(b)

	for n := 0; n < b.N; n++ {
		router.match(sampleRequests[5], &routeMatch{currentPath: sampleRequests[5].URL.Path})
	}
}

func BenchmarkSubrouterNotFound(b *testing.B) {
	router := setupRouteBench(b)

	for n := 0; n < b.N; n++ {
		router.match(sampleRequests[6], &routeMatch{currentPath: sampleRequests[6].URL.Path})
	}
}

func BenchmarkParamMatch(b *testing.B) {
	router := setupRouteBench(b)

	for n := 0; n < b.N; n++ {
		router.match(sampleRequests[7], &routeMatch{currentPath: sampleRequests[7].URL.Path})
	}
}

func BenchmarkParamPutMatch(b *testing.B) {
	router := setupRouteBench(b)

	for n := 0; n < b.N; n++ {
		router.match(sampleRequests[8], &routeMatch{currentPath: sampleRequests[8].URL.Path})
	}
}

func BenchmarkParamDeleteMatch(b *testing.B) {
	router := setupRouteBench(b)

	for n := 0; n < b.N; n++ {
		router.match(sampleRequests[9], &routeMatch{currentPath: sampleRequests[9].URL.Path})
	}
}

func BenchmarkMatchAll(b *testing.B) {
	router := setupRouteBench(b)

	for n := 0; n < b.N; n++ {
		for _, r := range sampleRequests {
			router.match(r, &routeMatch{currentPath: r.URL.Path})
		}
	}
}

func setupRouteBench(b *testing.B) *Router {
	router := registerAll(sampleRouteDefinition)
	b.ReportAllocs()
	runtime.GC()
	defer b.ResetTimer()
	return router
}
