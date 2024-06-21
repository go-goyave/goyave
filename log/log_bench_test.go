package log

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/slog"
)

func BenchmarkServeHTTPWithLogs(b *testing.B) {
	cfg := config.LoadDefault()
	cfg.Set("app.debug", false)
	logger := slog.New(slog.NewHandler(false, io.Discard))
	s, _ := goyave.New(goyave.Options{Config: cfg, Logger: logger})

	s.RegisterRoutes(func(_ *goyave.Server, r *goyave.Router) {
		r.GlobalMiddleware(CombinedLogMiddleware())
		r.Get("/user/{id}", func(r *goyave.Response, req *goyave.Request) {
			r.String(http.StatusOK, req.RouteParams["id"])
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/user/1", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s.Router().ServeHTTP(httptest.NewRecorder(), req)
	}
}
