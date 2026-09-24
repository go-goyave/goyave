package semconv

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

func TestMetricAttrs(t *testing.T) {
	// Short test just to see all attributes are generated
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

	want := []attribute.KeyValue{
		semconv.HTTPRequestMethodGet,
		semconv.HTTPRoute("/test/{param}"),
		semconv.NetworkProtocolVersion("2"),
		semconv.URLScheme("https"),
		semconv.ServerAddress("example.org"),
		semconv.ServerPort(443),
		semconv.ErrorType(fmt.Errorf("test error")),
		semconv.HTTPResponseStatusCode(http.StatusOK),
	}

	assert.Equal(t, want, MetricAttrs(request, route, http.StatusOK, fmt.Errorf("test error")))
}
