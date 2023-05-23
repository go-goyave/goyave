package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4/util/typeutil"
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
		{value: "string", want: typeutil.Ptr("string")},
		{value: "", want: typeutil.Ptr("")},
		{value: "\t\n\v\f\r ", want: typeutil.Ptr("")},
		{value: "\t\n\v\f\r hello\t\n\v\f\r ", want: typeutil.Ptr("hello")},
		{value: "hello\t\n\v\f\r ", want: typeutil.Ptr("hello")},
		{value: 'a'},
		{value: 2},
		{value: 2.5},
		{value: []string{"string"}},
		{value: map[string]any{"a": 1}},
		{value: true},
		{value: nil},
	}

	for _, c := range cases {
		c := c
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
