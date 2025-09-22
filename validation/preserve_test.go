package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreserveValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Preserve(String())
		assert.NotNil(t, v)
		assert.Equal(t, "preserve", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
	})

	cases := []struct {
		validator Validator
		value     any
		wantValue any
		want      bool
	}{
		{validator: String(), value: "string", want: true, wantValue: "string"},
		{validator: URL(), value: "https://example.com", want: true, wantValue: "https://example.com"},
		{validator: URL(), value: "https//example.com", want: false},
		{validator: UUID(), value: "9b531323-c0b0-46d0-8388-3e7acc0e1913", want: true, wantValue: "9b531323-c0b0-46d0-8388-3e7acc0e1913"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Preserve(c.validator)
			ctx := &Context{
				Value: c.value,
			}
			ok := v.Validate(ctx)
			if ok {
				if assert.Equal(t, c.want, ok) && ok {
					assert.Equal(t, c.wantValue, ctx.Value)
				}
			}
		})
	}
}
