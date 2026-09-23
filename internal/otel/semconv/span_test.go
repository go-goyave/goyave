package semconv

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

func TestMethod(t *testing.T) {
	cases := []struct {
		desc   string
		method string
		want   []attribute.KeyValue
	}{
		{
			desc:   "OK",
			method: http.MethodPost,
			want:   []attribute.KeyValue{semconv.HTTPRequestMethodPost},
		},
		{
			desc:   "wrong_case",
			method: "PoSt",
			want:   []attribute.KeyValue{semconv.HTTPRequestMethodPost, semconv.HTTPRequestMethodOriginal("PoSt")},
		},
		{
			desc:   "other",
			method: "NOT_A_METHOD",
			want:   []attribute.KeyValue{semconv.HTTPRequestMethodOther},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			attrs := []attribute.KeyValue{}
			Method(c.method, &attrs)
			assert.Equal(t, c.want, attrs)
		})
	}
}

func TestServerAddress(t *testing.T) {
	cases := []struct {
		desc    string
		request *http.Request
		want    []attribute.KeyValue
	}{
		{
			desc: "Forwarded",
			request: &http.Request{
				Host: "example.io:8443",
				URL:  &url.URL{Scheme: "not_a_proto"},
				Header: http.Header{
					"Forwarded":        []string{`for=192.0.2.60;proto=https;host="example.org:8443"`},
					"X-Forwarded-Host": []string{"example.com:8443"},
				},
			},
			want: []attribute.KeyValue{
				semconv.URLScheme("https"),
				semconv.ServerAddress("example.org"),
				semconv.ServerPort(8443),
			},
		},
		{
			desc: "Forwarded_default_port_https",
			request: &http.Request{
				Host: "example.io",
				URL:  &url.URL{Scheme: "not_a_proto"},
				Header: http.Header{
					"Forwarded":        []string{`for=192.0.2.60;proto=https;host="example.org"`},
					"X-Forwarded-Host": []string{"example.com"},
				},
			},
			want: []attribute.KeyValue{
				semconv.URLScheme("https"),
				semconv.ServerAddress("example.org"),
				semconv.ServerPort(443),
			},
		},
		{
			desc: "Forwarded_default_port_http",
			request: &http.Request{
				Host: "example.io",
				URL:  &url.URL{Scheme: "not_a_proto"},
				Header: http.Header{
					"Forwarded":        []string{`for=192.0.2.60;proto=http;host="example.org"`},
					"X-Forwarded-Host": []string{"example.com"},
				},
			},
			want: []attribute.KeyValue{
				semconv.URLScheme("http"),
				semconv.ServerAddress("example.org"),
				semconv.ServerPort(80),
			},
		},
		{
			desc: "Forwarded_default_port_ipv6",
			request: &http.Request{
				Host:   "[::1]",
				URL:    &url.URL{Scheme: "https"},
				Header: http.Header{},
			},
			want: []attribute.KeyValue{
				semconv.URLScheme("https"),
				semconv.ServerAddress("::1"),
				semconv.ServerPort(443),
			},
		},
		{
			desc: "X-Forwarded",
			request: &http.Request{
				Host: "example.com:8443",
				URL:  &url.URL{Scheme: "not_a_proto"},
				Header: http.Header{
					"X-Forwarded-Host":  []string{"example.org:8443"},
					"X-Forwarded-Proto": []string{"https"},
				},
			},
			want: []attribute.KeyValue{
				semconv.URLScheme("https"),
				semconv.ServerAddress("example.org"),
				semconv.ServerPort(8443),
			},
		},
		{
			desc: "Host",
			request: &http.Request{
				Host:   "example.org:8443",
				URL:    &url.URL{Scheme: "https"},
				Header: http.Header{},
			},
			want: []attribute.KeyValue{
				semconv.URLScheme("https"),
				semconv.ServerAddress("example.org"),
				semconv.ServerPort(8443),
			},
		},
		{
			desc: "invalid_host",
			request: &http.Request{
				Host:   "[::1",
				URL:    &url.URL{Scheme: "https"},
				Header: http.Header{},
			},
			want: []attribute.KeyValue{
				semconv.URLScheme("https"),
			},
		},
		{
			desc: "no_proto",
			request: &http.Request{
				Host:   "[::1",
				URL:    &url.URL{},
				Header: http.Header{},
			},
			want: []attribute.KeyValue{},
		},
		{
			desc: "invalid_port",
			request: &http.Request{
				Host:   "example.org",
				URL:    &url.URL{Scheme: "not_a_proto"},
				Header: http.Header{},
			},
			want: []attribute.KeyValue{
				semconv.URLScheme("not_a_proto"),
				semconv.ServerAddress("example.org"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			attrs := []attribute.KeyValue{}
			ServerAddress(c.request, &attrs)
			assert.Equal(t, c.want, attrs)
		})
	}
}

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

func TestNetwork(t *testing.T) {
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
				semconv.NetworkProtocolVersion("1.1"),
				semconv.NetworkPeerAddress("192.0.2.60"),
				semconv.NetworkPeerPort(44444),
			},
		},
		{
			desc: "no_proto_version",
			request: &http.Request{
				Proto:      "HTTP",
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
				semconv.NetworkProtocolVersion("1.1"),
				semconv.NetworkPeerAddress("192.0.2.60"),
			},
		},
		{
			desc: "invalid_host",
			request: &http.Request{
				Proto:      "HTTP/1.1",
				RemoteAddr: "[::1",
			},
			want: []attribute.KeyValue{
				semconv.NetworkProtocolVersion("1.1"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			attrs := []attribute.KeyValue{}
			Network(c.request, &attrs)
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

	want := []attribute.KeyValue{
		semconv.HTTPRequestMethodGet,
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

	assert.Equal(t, want, SpanAttrs(request))
}
