package semconv

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

var methods = map[string]attribute.KeyValue{
	http.MethodConnect: semconv.HTTPRequestMethodConnect,
	http.MethodDelete:  semconv.HTTPRequestMethodDelete,
	http.MethodGet:     semconv.HTTPRequestMethodGet,
	http.MethodHead:    semconv.HTTPRequestMethodHead,
	http.MethodOptions: semconv.HTTPRequestMethodOptions,
	http.MethodPatch:   semconv.HTTPRequestMethodPatch,
	http.MethodPost:    semconv.HTTPRequestMethodPost,
	http.MethodPut:     semconv.HTTPRequestMethodPut,
	http.MethodTrace:   semconv.HTTPRequestMethodTrace,
}

func SpanAttrs(request *http.Request) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 12) // Max possible length
	Method(request.Method, &attrs)
	ServerAddress(request, &attrs)
	ClientAddress(request, &attrs)
	URL(request.URL, &attrs)
	Network(request, &attrs)
	return attrs
}

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

func defaultPort(scheme string) int {
	switch scheme {
	case "https":
		return 443
	case "http":
		return 80
	}
	return -1
}

func ClientAddress(request *http.Request, attrs *[]attribute.KeyValue) {
	forwardedHost := getFor(request)
	if forwardedHost == "" {
		return
	}

	host, portStr, err := SplitHostPort(forwardedHost)
	if err != nil {
		return
	}

	*attrs = append(*attrs, semconv.ClientAddress(host))

	port, _ := strconv.Atoi(portStr)
	if port <= 0 {
		return
	}

	*attrs = append(*attrs, semconv.ClientPort(port))
}

func getFor(request *http.Request) string {
	// Order of priority:
	// Header Forwarded: for=
	// Header X-Forwarded-For
	// request.RemoteAddr

	forwarded := parseForwardedHeader(request.Header, "for")
	if forwarded != "" {
		return forwarded
	}

	xForwarded := request.Header.Get("X-Forwarded-For")
	if xForwarded != "" {
		return xForwarded
	}

	return request.RemoteAddr
}

func URL(u *url.URL, attrs *[]attribute.KeyValue) {
	if u.User != nil {
		u = u.Clone()
		u.User = url.UserPassword("REDACTED", "REDACTED")
	}

	*attrs = append(*attrs,
		semconv.URLFull(u.String()),
		semconv.URLPath(u.Path),
	)

	if u.RawQuery == "" {
		return
	}

	*attrs = append(*attrs, semconv.URLQuery(u.RawQuery))
}

func Network(request *http.Request, attrs *[]attribute.KeyValue) {
	_, protoVersion, _ := strings.Cut(request.Proto, "/")
	if protoVersion != "" {
		*attrs = append(*attrs, semconv.NetworkProtocolVersion(protoVersion))
	}

	peer, peerPortStr, err := SplitHostPort(request.RemoteAddr)
	if err == nil && peer != "" {
		*attrs = append(*attrs, semconv.NetworkPeerAddress(peer))

		peerPort, _ := strconv.Atoi(peerPortStr)
		if peerPort > 0 {
			*attrs = append(*attrs, semconv.NetworkPeerPort(peerPort))
		}
	}
}

func SplitHostPort(hostport string) (string, string, error) {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		if strings.Contains(err.Error(), "missing port") {
			// We just want a valid host in that case, reparse with a fake port that won't parse in Atoi.
			return net.SplitHostPort(hostport + ":a")
		}
		return "", "", err
	}
	return host, port, nil
}

func SpanName(method, route string) string {
	_, ok := methods[method]
	if !ok {
		method = "HTTP"
	}

	if route == "" {
		return method
	}

	return method + " " + route
}
