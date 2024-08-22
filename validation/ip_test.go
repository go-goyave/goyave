package validation

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := IP()
		assert.NotNil(t, v)
		assert.Equal(t, "ip", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue net.IP
		want      bool
	}{
		{value: "127.0.0.1", want: true, wantValue: net.ParseIP("127.0.0.1")},
		{value: net.ParseIP("127.0.0.1"), want: true, wantValue: net.ParseIP("127.0.0.1")},
		{value: "192.168.0.1", want: true, wantValue: net.ParseIP("192.168.0.1")},
		{value: "88.88.88.88", want: true, wantValue: net.ParseIP("88.88.88.88")},
		{value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", want: true, wantValue: net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")},
		{value: "2001:db8:85a3::8a2e:370:7334", want: true, wantValue: net.ParseIP("2001:db8:85a3::8a2e:370:7334")},
		{value: "2001:db8:85a3:0:0:8a2e:370:7334", want: true, wantValue: net.ParseIP("2001:db8:85a3:0:0:8a2e:370:7334")},
		{value: "2001:db8:85a3:8d3:1319:8a2e:370:7348", want: true, wantValue: net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348")},
		{value: "::1", want: true, wantValue: net.ParseIP("::1")},
		{value: net.ParseIP("::1"), want: true, wantValue: net.ParseIP("::1")},
		{value: "256.0.0.1", want: false},
		{value: "string", want: false},
		{value: "", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: []byte{}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := IP()
			ctx := &Context{
				Value: c.value,
			}
			ok := v.Validate(ctx)
			if assert.Equal(t, c.want, ok) && ok {
				assert.Equal(t, c.wantValue, ctx.Value)
			}
		})
	}
}

func TestIPv4Validator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := IPv4()
		assert.NotNil(t, v)
		assert.Equal(t, "ipv4", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue net.IP
		want      bool
	}{
		{value: "127.0.0.1", want: true, wantValue: net.ParseIP("127.0.0.1")},
		{value: net.ParseIP("127.0.0.1"), want: true, wantValue: net.ParseIP("127.0.0.1")},
		{value: "192.168.0.1", want: true, wantValue: net.ParseIP("192.168.0.1")},
		{value: "88.88.88.88", want: true, wantValue: net.ParseIP("88.88.88.88")},
		{value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", want: false, wantValue: net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")},
		{value: "2001:db8:85a3::8a2e:370:7334", want: false, wantValue: net.ParseIP("2001:db8:85a3::8a2e:370:7334")},
		{value: "2001:db8:85a3:0:0:8a2e:370:7334", want: false, wantValue: net.ParseIP("2001:db8:85a3:0:0:8a2e:370:7334")},
		{value: "2001:db8:85a3:8d3:1319:8a2e:370:7348", want: false, wantValue: net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348")},
		{value: "::1", want: false, wantValue: net.ParseIP("::1")},
		{value: net.ParseIP("::1"), want: false, wantValue: net.ParseIP("::1")},
		{value: "256.0.0.1", want: false},
		{value: "string", want: false},
		{value: "", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: []byte{}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := IPv4()
			ctx := &Context{
				Value: c.value,
			}
			ok := v.Validate(ctx)
			if assert.Equal(t, c.want, ok) && ok {
				assert.Equal(t, c.wantValue, ctx.Value)
			}
		})
	}
}

func TestIPv6Validator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := IPv6()
		assert.NotNil(t, v)
		assert.Equal(t, "ipv6", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue net.IP
		want      bool
	}{
		{value: "127.0.0.1", want: false, wantValue: net.ParseIP("127.0.0.1")},
		{value: net.ParseIP("127.0.0.1"), want: false, wantValue: net.ParseIP("127.0.0.1")},
		{value: "192.168.0.1", want: false, wantValue: net.ParseIP("192.168.0.1")},
		{value: "88.88.88.88", want: false, wantValue: net.ParseIP("88.88.88.88")},
		{value: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", want: true, wantValue: net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")},
		{value: "2001:db8:85a3::8a2e:370:7334", want: true, wantValue: net.ParseIP("2001:db8:85a3::8a2e:370:7334")},
		{value: "2001:db8:85a3:0:0:8a2e:370:7334", want: true, wantValue: net.ParseIP("2001:db8:85a3:0:0:8a2e:370:7334")},
		{value: "2001:db8:85a3:8d3:1319:8a2e:370:7348", want: true, wantValue: net.ParseIP("2001:db8:85a3:8d3:1319:8a2e:370:7348")},
		{value: "::1", want: true, wantValue: net.ParseIP("::1")},
		{value: net.ParseIP("::1"), want: true, wantValue: net.ParseIP("::1")},
		{value: "256.0.0.1", want: false},
		{value: "string", want: false},
		{value: "", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: []byte{}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := IPv6()
			ctx := &Context{
				Value: c.value,
			}
			ok := v.Validate(ctx)
			if assert.Equal(t, c.want, ok) && ok {
				assert.Equal(t, c.wantValue, ctx.Value)
			}
		})
	}
}
