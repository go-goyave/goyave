package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Object()
		assert.NotNil(t, v)
		assert.Equal(t, "object", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: map[string]any{"a": 1}, want: true, wantValue: map[string]any{"a": 1}},
		{value: `{"a": 1}`, want: true, wantValue: map[string]any{"a": 1.0}},
		{value: `"a"`, want: false},
		{value: "string", want: false},
		{value: "", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Object()
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
