package semconv

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

func MetricAttrs(request *http.Request, route string, status int, err error) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 8) // Max possible length
	Method(request.Method, &attrs)
	Route(route, &attrs)
	NetworkProtocolVersion(request, &attrs)
	ServerAddress(request, &attrs)
	if err != nil {
		attrs = append(attrs, semconv.ErrorType(err))
	}
	attrs = append(attrs, semconv.HTTPResponseStatusCode(status))
	return attrs
}
