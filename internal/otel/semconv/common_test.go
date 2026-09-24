package semconv

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
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
			want:   []attribute.KeyValue{semconv.HTTPRequestMethodOther, semconv.HTTPRequestMethodOriginal("PoSt")},
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

func TestRoute(t *testing.T) {
	cases := []struct {
		desc  string
		route string
		want  []attribute.KeyValue
	}{
		{
			desc:  "empty",
			route: "",
			want:  []attribute.KeyValue{},
		},
		{
			desc:  "OK",
			route: "/test/{param}",
			want:  []attribute.KeyValue{semconv.HTTPRoute("/test/{param}")},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			attrs := []attribute.KeyValue{}
			Route(c.route, &attrs)
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

func TestNetworkProtocolVersion(t *testing.T) {
	cases := []struct {
		desc    string
		request *http.Request
		want    []attribute.KeyValue
	}{
		{
			desc: "HTTP_1.1",
			request: &http.Request{
				Proto: "HTTP/1.1",
			},
			want: []attribute.KeyValue{
				semconv.NetworkProtocolVersion("1.1"),
			},
		},
		{
			desc: "HTTP_2",
			request: &http.Request{
				Proto: "HTTP/2",
			},
			want: []attribute.KeyValue{
				semconv.NetworkProtocolVersion("2"),
			},
		},
		{
			desc: "HTTP_2.0",
			request: &http.Request{
				Proto: "HTTP/2.0",
			},
			want: []attribute.KeyValue{
				semconv.NetworkProtocolVersion("2.0"),
			},
		},
		{
			desc: "no_version",
			request: &http.Request{
				Proto: "HTTP",
			},
			want: []attribute.KeyValue{},
		},
		{
			desc:    "no_proto",
			request: &http.Request{},
			want:    []attribute.KeyValue{},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			attrs := []attribute.KeyValue{}
			NetworkProtocolVersion(c.request, &attrs)
			assert.Equal(t, c.want, attrs)
		})
	}
}
