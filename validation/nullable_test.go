package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullableValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Nullable()
		assert.NotNil(t, v)
		assert.Equal(t, "nullable", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value any
		want  bool
	}{
		// Should always return true
		{value: "string", want: true},
		{value: "", want: true},
		{value: 'a', want: true},
		{value: 2, want: true},
		{value: 2.5, want: true},
		{value: []string{"string"}, want: true},
		{value: map[string]any{"a": 1}, want: true},
		{value: true, want: true},
		{value: nil, want: true},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Nullable()
			ctx := &Context{
				Value: c.value,
			}

			assert.Equal(t, c.want, v.Validate(ctx))
		})
	}
}
