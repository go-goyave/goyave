package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := JSON()
		assert.NotNil(t, v)
		assert.Equal(t, "json", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: `"string"`, want: true, wantValue: "string"},
		{value: `0.3`, want: true, wantValue: 0.3},
		{value: `true`, want: true, wantValue: true},
		{value: `["a","b"]`, want: true, wantValue: []any{"a", "b"}},
		{value: `{"a": "b"}`, want: true, wantValue: map[string]any{"a": "b"}},
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
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := JSON()
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
