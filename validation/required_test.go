package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiredValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Required()
		assert.NotNil(t, v)
		assert.Equal(t, "required", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&ContextV5{}))
	})

	cases := []struct {
		value    any
		nullable bool
		want     bool
	}{
		{value: "string", want: true},
		{value: "", want: true},
		{value: 'a', want: true},
		{value: 2, want: true},
		{value: 2.5, want: true},
		{value: []string{"string"}, want: true},
		{value: map[string]any{"a": 1}, want: true},
		{value: true, want: true},
		{value: nil, want: false},
		{value: nil, want: true, nullable: true},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Required()
			ctx := &ContextV5{
				Value: c.value,
				Field: &FieldV5{
					isNullable: c.nullable,
				},
			}

			assert.Equal(t, c.want, v.Validate(ctx))
		})
	}
}

func TestRequiredIfValidator(t *testing.T) {
	alwaysRequired := func(c *ContextV5) bool { return true }
	t.Run("Constructor", func(t *testing.T) {
		v := RequiredIf(alwaysRequired)
		assert.NotNil(t, v)
		assert.Equal(t, "required", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&ContextV5{}))
	})

	cases := []struct {
		value     any
		condition func(*ContextV5) bool
		nullable  bool
		want      bool
	}{
		{value: "string", want: true, condition: alwaysRequired},
		{value: "", want: true, condition: alwaysRequired},
		{value: 'a', want: true, condition: alwaysRequired},
		{value: 2, want: true, condition: alwaysRequired},
		{value: 2.5, want: true, condition: alwaysRequired},
		{value: []string{"string"}, want: true, condition: alwaysRequired},
		{value: map[string]any{"a": 1}, want: true, condition: alwaysRequired},
		{value: true, want: true, condition: alwaysRequired},
		{value: nil, want: false, condition: alwaysRequired},
		{value: nil, want: true, nullable: true, condition: alwaysRequired},
		{value: nil, want: true, condition: func(c *ContextV5) bool {
			return false
		}},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := RequiredIf(c.condition)
			ctx := &ContextV5{
				Value: c.value,
				Field: &FieldV5{
					isNullable: c.nullable,
				},
			}

			assert.Equal(t, c.want, v.Validate(ctx))
		})
	}
}
