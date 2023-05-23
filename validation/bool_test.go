package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Bool()
		assert.NotNil(t, v)
		assert.Equal(t, "bool", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		want      bool
		wantValue bool
	}{
		{value: true, want: true, wantValue: true},
		{value: false, want: true, wantValue: false},

		// Strings
		{value: "string", want: false},
		{value: "", want: false},
		{value: "1", want: true, wantValue: true},
		{value: "on", want: true, wantValue: true},
		{value: "true", want: true, wantValue: true},
		{value: "yes", want: true, wantValue: true},
		{value: "0", want: true, wantValue: false},
		{value: "off", want: true, wantValue: false},
		{value: "false", want: true, wantValue: false},
		{value: "no", want: true, wantValue: false},

		// Numbers (!= 0)
		{value: 'a', want: true, wantValue: true},
		{value: rune(0), want: true, wantValue: false},
		{value: int(1), want: true, wantValue: true},
		{value: int8(1), want: true, wantValue: true},
		{value: int16(1), want: true, wantValue: true},
		{value: int32(1), want: true, wantValue: true},
		{value: int64(1), want: true, wantValue: true},
		{value: uint(1), want: true, wantValue: true},
		{value: uint8(1), want: true, wantValue: true},
		{value: uint16(1), want: true, wantValue: true},
		{value: uint32(1), want: true, wantValue: true},
		{value: uint64(1), want: true, wantValue: true},
		{value: float32(1), want: true, wantValue: true},
		{value: float64(1), want: true, wantValue: true},
		{value: int(0), want: true, wantValue: false},
		{value: int8(0), want: true, wantValue: false},
		{value: int16(0), want: true, wantValue: false},
		{value: int32(0), want: true, wantValue: false},
		{value: int64(0), want: true, wantValue: false},
		{value: uint(0), want: true, wantValue: false},
		{value: uint8(0), want: true, wantValue: false},
		{value: uint16(0), want: true, wantValue: false},
		{value: uint32(0), want: true, wantValue: false},
		{value: uint64(0), want: true, wantValue: false},
		{value: float32(0), want: true, wantValue: false},
		{value: float64(0), want: true, wantValue: false},

		// Invalid types
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Bool()
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
