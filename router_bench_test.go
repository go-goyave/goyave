package goyave

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"goyave.dev/goyave/v5/config"
)

func BenchmarkServeHTTP(b *testing.B) {
	s, _ := New(Options{Config: config.LoadDefault()})

	s.RegisterRoutes(func(_ *Server, r *Router) {
		r.Get("/user/{id}", func(r *Response, req *Request) {
			r.String(http.StatusOK, req.RouteParams["id"])
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/user/1", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s.router.ServeHTTP(httptest.NewRecorder(), req)
	}
}
