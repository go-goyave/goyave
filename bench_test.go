package goyave

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/System-Glitch/goyave/v2/config"
)

var staticGoyave *Router

func calcMem(name string, load func()) {
	m := new(runtime.MemStats)

	// before
	runtime.GC()
	runtime.GC()
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(m)
	before := m.HeapAlloc

	load()

	// after
	runtime.GC()
	runtime.GC()
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(m)
	after := m.HeapAlloc

	println(""+name+":", after-before, "Bytes")
}

func goyaveHandler(res *Response, req *Request) {
	res.String(http.StatusOK, "OK")
}

func loadGoyaveRouter(routes []route) *Router {
	config.Load()
	router := newRouter()

	for _, r := range routes {
		switch r.method {
		case "GET":
			router.Route("GET", r.path, goyaveHandler, nil)
		case "POST":
			router.Route("POST", r.path, goyaveHandler, nil)
		case "PATCH":
			router.Route("PATCH", r.path, goyaveHandler, nil)
		case "PUT":
			router.Route("PUT", r.path, goyaveHandler, nil)
		case "DELETE":
			router.Route("DELETE", r.path, goyaveHandler, nil)
		}
	}

	return router
}

func benchRoutes(b *testing.B, router *Router, routes []route) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	u := r.URL
	rq := u.RawQuery

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, route := range routes {
			r.Method = route.method
			r.RequestURI = route.path
			u.Path = route.path
			u.RawQuery = rq
			router.requestHandler(w, r, goyaveHandler, nil)
		}
	}
}

func Benchmark_Static(b *testing.B) {
	println("# Static routes len:", len(staticRoutes))
	calcMem("Goyave", func() {
		staticGoyave = loadGoyaveRouter(staticRoutes)
	})
	benchRoutes(b, staticGoyave, staticRoutes)
}

func Benchmark_Parse(b *testing.B) {
	println("# Parse routes len:", len(parseAPIRoutes))
	calcMem("Goyave", func() {
		staticGoyave = loadGoyaveRouter(parseAPIRoutes)
	})
	benchRoutes(b, staticGoyave, parseAPIRoutes)
}

func Benchmark_Github(b *testing.B) {
	println("# Github routes len:", len(githubAPIRoutes))
	calcMem("Goyave", func() {
		staticGoyave = loadGoyaveRouter(githubAPIRoutes)
	})
	benchRoutes(b, staticGoyave, githubAPIRoutes)
}

func Benchmark_GPlus(b *testing.B) {
	println("# Google+ routes len:", len(gplusAPIRoutes))
	calcMem("Goyave", func() {
		staticGoyave = loadGoyaveRouter(gplusAPIRoutes)
	})
	benchRoutes(b, staticGoyave, gplusAPIRoutes)
}
