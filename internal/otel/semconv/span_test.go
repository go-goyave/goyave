package semconv

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

func TestClientAddress(t *testing.T) {
	cases := []struct {
		desc    string
		request *http.Request
		want    []attribute.KeyValue
	}{
		{
			desc: "Forwarded",
			request: &http.Request{
				RemoteAddr: "192.0.2.60:44444",
				Header: http.Header{
					"Forwarded":       []string{`for=192.0.2.60:44444;proto=https;host="example.org:8443"`},
					"X-Forwarded-For": []string{"192.0.2.61:44444"},
				},
			},
			want: []attribute.KeyValue{
				semconv.ClientAddress("192.0.2.60"),
				semconv.ClientPort(44444),
			},
		},
		{
			desc: "XForwarded",
			request: &http.Request{
				RemoteAddr: "192.0.2.61:44444",
				Header: http.Header{
					"X-Forwarded-For": []string{"192.0.2.60:44444"},
				},
			},
			want: []attribute.KeyValue{
				semconv.ClientAddress("192.0.2.60"),
				semconv.ClientPort(44444),
			},
		},
		{
			desc: "RemoteAddr",
			request: &http.Request{
				RemoteAddr: "192.0.2.60:44444",
				Header:     http.Header{},
			},
			want: []attribute.KeyValue{
				semconv.ClientAddress("192.0.2.60"),
				semconv.ClientPort(44444),
			},
		},
		{
			desc: "missing_port",
			request: &http.Request{
				RemoteAddr: "192.0.2.60",
				Header:     http.Header{},
			},
			want: []attribute.KeyValue{
				semconv.ClientAddress("192.0.2.60"),
			},
		},
		{
			desc: "invalid_port",
			request: &http.Request{
				RemoteAddr: "192.0.2.60:abc",
				Header:     http.Header{},
			},
			want: []attribute.KeyValue{
				semconv.ClientAddress("192.0.2.60"),
			},
		},
		{
			desc: "invalid_addr",
			request: &http.Request{
				RemoteAddr: "[::1",
				Header:     http.Header{},
			},
			want: []attribute.KeyValue{},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			attrs := []attribute.KeyValue{}
			ClientAddress(c.request, &attrs)
			assert.Equal(t, c.want, attrs)
		})
	}
}

func TestURL(t *testing.T) {
	cases := []struct {
		desc string
		url  *url.URL
		want []attribute.KeyValue
	}{
		{
			desc: "OK",
			url: &url.URL{
				Scheme:   "https",
				Host:     "example.org",
				Path:     "/test",
				RawQuery: "q=otel",
			},
			want: []attribute.KeyValue{
				semconv.URLFull("https://example.org/test?q=otel"),
				semconv.URLPath("/test"),
				semconv.URLQuery("q=otel"),
			},
		},
		{
			desc: "no_query",
			url: &url.URL{
				Scheme: "https",
				Host:   "example.org",
				Path:   "/test",
			},
			want: []attribute.KeyValue{
				semconv.URLFull("https://example.org/test"),
				semconv.URLPath("/test"),
			},
		},
		{
			desc: "redact_user",
			url: &url.URL{
				Scheme: "https",
				Host:   "example.org",
				Path:   "/test",
				User:   url.UserPassword("johndoe", "secret"),
			},
			want: []attribute.KeyValue{
				semconv.URLFull("https://REDACTED:REDACTED@example.org/test"),
				semconv.URLPath("/test"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			attrs := []attribute.KeyValue{}
			URL(c.url, &attrs)
			assert.Equal(t, c.want, attrs)
		})
	}
}

func TestNetworkPeer(t *testing.T) {
	cases := []struct {
		desc    string
		request *http.Request
		want    []attribute.KeyValue
	}{
		{
			desc: "OK",
			request: &http.Request{
				Proto:      "HTTP/1.1",
				RemoteAddr: "192.0.2.60:44444",
			},
			want: []attribute.KeyValue{
				semconv.NetworkPeerAddress("192.0.2.60"),
				semconv.NetworkPeerPort(44444),
			},
		},
		{
			desc: "no_port",
			request: &http.Request{
				Proto:      "HTTP/1.1",
				RemoteAddr: "192.0.2.60",
			},
			want: []attribute.KeyValue{
				semconv.NetworkPeerAddress("192.0.2.60"),
			},
		},
		{
			desc: "invalid_host",
			request: &http.Request{
				Proto:      "HTTP/1.1",
				RemoteAddr: "[::1",
			},
			want: []attribute.KeyValue{},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			attrs := []attribute.KeyValue{}
			NetworkPeer(c.request, &attrs)
			assert.Equal(t, c.want, attrs)
		})
	}
}

func TestSpanAttrs(t *testing.T) {
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
		semconv.URLScheme("https"),
		semconv.ServerAddress("example.org"),
		semconv.ServerPort(443),
		semconv.ClientAddress("192.0.2.60"),
		semconv.ClientPort(44444),
		semconv.URLFull("https://example.org/test"),
		semconv.URLPath("/test"),
		semconv.NetworkProtocolVersion("2"),
		semconv.NetworkPeerAddress("192.0.2.60"),
		semconv.NetworkPeerPort(44444),
	}

	assert.Equal(t, want, SpanAttrs(request, route))
}

func TestSpanName(t *testing.T) {
	cases := []struct {
		desc   string
		method string
		uri    string
		want   string
	}{
		{
			desc:   "OK",
			method: http.MethodPost,
			uri:    "/test/{param}",
			want:   "POST /test/{param}",
		},
		{
			desc:   "wrong_method_case",
			method: "PoSt",
			uri:    "/test/{param}",
			want:   "HTTP /test/{param}",
		},
		{
			desc:   "other_method",
			method: "NOT_A_METHOD",
			uri:    "/test/{param}",
			want:   "HTTP /test/{param}",
		},
		{
			desc:   "empty_route",
			method: http.MethodPost,
			uri:    "",
			want:   "POST",
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			assert.Equal(t, c.want, SpanName(c.method, c.uri))
		})
	}
}
