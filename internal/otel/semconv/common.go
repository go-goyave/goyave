package semconv

import (
	"net/http"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func Method(method string, attrs *[]attribute.KeyValue) {
	attr, ok := methods[method]
	if ok {
		*attrs = append(*attrs, attr)
		return
	}

	*attrs = append(*attrs, semconv.HTTPRequestMethodOther)

	// Note: Goyave method matching is case-sensitive
	if _, ok := methods[strings.ToUpper(method)]; ok {
		*attrs = append(*attrs, semconv.HTTPRequestMethodOriginal(method))
	}
}

func Route(route string, attrs *[]attribute.KeyValue) {
	if route == "" {
		return
	}
	*attrs = append(*attrs, semconv.HTTPRoute(route))
}

func ServerAddress(request *http.Request, attrs *[]attribute.KeyValue) {
	// Handling scheme here instead of in [URL] so we can use it for default server.port
	scheme := getScheme(request)
	if scheme != "" {
		*attrs = append(*attrs, semconv.URLScheme(scheme))
	}

	forwardedHost := getHost(request)
	if forwardedHost == "" {
		return
	}

	host, portStr, err := SplitHostPort(forwardedHost)
	if err != nil {
		return
	}

	*attrs = append(*attrs, semconv.ServerAddress(host))

	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = defaultPort(scheme)
	}

	if port <= 0 {
		return
	}

	*attrs = append(*attrs, semconv.ServerPort(port))
}

func getHost(request *http.Request) string {
	// Order of priority according to OpenTelemetry spec:
	// Header Forwarded: host=
	// Header X-Forwarded-Host
	// request.Host (covers the Host header and :authority pseudo-header)

	forwarded := parseForwardedHeader(request.Header, "host")
	if forwarded != "" {
		return forwarded
	}

	xForwarded := request.Header.Get("X-Forwarded-Host")
	if xForwarded != "" {
		return xForwarded
	}

	return request.Host
}

func defaultPort(scheme string) int {
	switch scheme {
	case "https":
		return 443
	case "http":
		return 80
	}
	return -1
}

func getScheme(request *http.Request) string {
	// Order of priority:
	// Header Forwarded: proto=
	// Header X-Forwarded-Proto
	// request.URL.Scheme

	forwarded := parseForwardedHeader(request.Header, "proto")
	if forwarded != "" {
		return forwarded
	}

	xForwarded := request.Header.Get("X-Forwarded-Proto")
	if xForwarded != "" {
		return xForwarded
	}

	return request.URL.Scheme
}

func NetworkProtocolVersion(request *http.Request, attrs *[]attribute.KeyValue) {
	_, protoVersion, _ := strings.Cut(request.Proto, "/")
	if protoVersion == "" {
		return
	}
	*attrs = append(*attrs, semconv.NetworkProtocolVersion(protoVersion))
}
