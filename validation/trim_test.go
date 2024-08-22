package validation

import (
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestTrimValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Trim()
		assert.NotNil(t, v)
		assert.Equal(t, "trim", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value any
		want  *string
	}{
		{value: "string", want: lo.ToPtr("string")},
		{value: "", want: lo.ToPtr("")},
		{value: "\t\n\v\f\r ", want: lo.ToPtr("")},
		{value: "\t\n\v\f\r hello\t\n\v\f\r ", want: lo.ToPtr("hello")},
		{value: "hello\t\n\v\f\r ", want: lo.ToPtr("hello")},
		{value: 'a'},
		{value: 2},
		{value: 2.5},
		{value: []string{"string"}},
		{value: map[string]any{"a": 1}},
		{value: true},
		{value: nil},
	}

	for _, c := range cases {
		format := "Validate_%q"
		if _, ok := c.value.(string); !ok {
			format = "Validate_%v"
		}
		name := ""
		if c.want != nil {
			format += "_%v"
			name = fmt.Sprintf(format, c.value, *c.want)
		} else {
			name = fmt.Sprintf(format, c.value)
		}
		t.Run(name, func(t *testing.T) {
			v := Trim()

			ctx := &Context{
				Value: c.value,
			}
			assert.True(t, v.Validate(ctx))

			if c.want != nil {
				assert.Equal(t, *c.want, ctx.Value)
			}
		})
	}
}
