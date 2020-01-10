package goyave

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

type route struct {
	method string
	path   string
}

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
	println("   "+name+":", after-before, "Bytes")
}

func goyaveHandler(res *Response, req *Request) {
	res.String(http.StatusOK, "OK")
}

func loadGoyaveRouter(routes []route) *Router {
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
