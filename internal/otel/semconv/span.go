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

func SpanAttrs(request *http.Request, route string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 13) // Max possible length
	Method(request.Method, &attrs)
	Route(route, &attrs)
	ServerAddress(request, &attrs)
	ClientAddress(request, &attrs)
	URL(request.URL, &attrs)
	NetworkProtocolVersion(request, &attrs)
	NetworkPeer(request, &attrs)
	return attrs
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

func NetworkPeer(request *http.Request, attrs *[]attribute.KeyValue) {
	peer, peerPortStr, err := SplitHostPort(request.RemoteAddr)
	if err != nil || peer == "" {
		return
	}
	*attrs = append(*attrs, semconv.NetworkPeerAddress(peer))

	peerPort, _ := strconv.Atoi(peerPortStr)
	if peerPort > 0 {
		*attrs = append(*attrs, semconv.NetworkPeerPort(peerPort))
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
