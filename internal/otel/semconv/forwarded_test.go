package semconv

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseForwardedHeader(t *testing.T) {
	cases := []struct {
		desc          string
		parameterName string
		values        []string
		want          string
	}{
		{desc: "simple token host", parameterName: "host", values: []string{"host=example.org"}, want: "example.org"},
		{desc: "simple token for", parameterName: "for", values: []string{"for=192.0.2.60"}, want: "192.0.2.60"},
		{desc: "quoted with port host", parameterName: "host", values: []string{`for=192.0.2.60;proto=https;host="example.org:8443"`}, want: "example.org:8443"},
		{desc: "quoted with port for", parameterName: "for", values: []string{`for=192.0.2.60;proto=https;host="example.org:8443"`}, want: "192.0.2.60"},
		{desc: "case-insensitive key", parameterName: "host", values: []string{"Host=example.org;Proto=https"}, want: "example.org"},
		{desc: "ipv6 host", parameterName: "host", values: []string{`host="[2001:db8::1]:443"`}, want: "[2001:db8::1]:443"},
		{desc: "ipv6 for", parameterName: "for", values: []string{`for="[2001:db8::1]:4443"`}, want: "[2001:db8::1]:4443"},
		{desc: "first hop wins host", parameterName: "host", values: []string{"host=example.org, host=internal.local"}, want: "example.org"},
		{desc: "first hop wins for", parameterName: "for", values: []string{"for=192.0.2.60:4444, for=192.0.2.61:4445"}, want: "192.0.2.60:4444"},
		{desc: "separate header lines", parameterName: "host", values: []string{"host=example.org", "host=internal.local"}, want: "example.org"},
		{desc: "quoted comma and semicolon", parameterName: "host", values: []string{`for="a,b;c";host=example.org`}, want: "example.org"},
		{desc: "first hop without host", parameterName: "host", values: []string{"for=192.0.2.60, host=internal.local"}, want: ""},
		{desc: "escaped chars in host", parameterName: "host", values: []string{`host="ex\"am\\ple\.org"`}, want: `ex"am\ple.org`},
		{desc: "escaped quote before separators", parameterName: "host", values: []string{`for="a\",b;c";host=example.org`}, want: "example.org"},
		{desc: "absent", parameterName: "host", values: nil, want: ""},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			h := http.Header{}
			for _, v := range c.values {
				h.Add("Forwarded", v)
			}
			assert.Equal(t, c.want, parseForwardedHeader(h, c.parameterName))
		})
	}
}
