package semconv

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func BenchmarkMetricAttrs(b *testing.B) {
	request := &http.Request{
		Method:     http.MethodGet,
		Host:       "example.org:443",
		RemoteAddr: "192.0.2.60:44444",
		URL: &url.URL{
			Scheme: "https",
			Host:   "example.org",
			Path:   "/test",
		},
		Proto: "HTTP/2",
	}
	route := "/test/{param}"
	testErr := fmt.Errorf("test error")
	b.ReportAllocs()
	for b.Loop() {
		MetricAttrs(request, route, http.StatusOK, testErr)
	}
}
