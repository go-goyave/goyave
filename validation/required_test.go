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
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value    any
		nullable bool
		found    bool
		want     bool
	}{
		{value: "string", found: true, want: true},
		{value: "", found: true, want: true},
		{value: 'a', found: true, want: true},
		{value: 2, found: true, want: true},
		{value: 2.5, found: true, want: true},
		{value: []string{"string"}, found: true, want: true},
		{value: map[string]any{"a": 1}, found: true, want: true},
		{value: true, found: true, want: true},
		{value: nil, found: true, want: false},
		{value: nil, found: true, want: true, nullable: true},
		{value: nil, found: false, want: false, nullable: true},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Required()
			ctx := &Context{
				Value: c.value,
				Field: &Field{
					isNullable: c.nullable,
				},
				Found: c.found,
			}

			assert.Equal(t, c.want, v.Validate(ctx))
		})
	}
}

func TestRequiredIfValidator(t *testing.T) {
	alwaysRequired := func(_ *Context) bool { return true }
	neverRequired := func(_ *Context) bool { return false }
	t.Run("Constructor", func(t *testing.T) {
		v := RequiredIf(alwaysRequired)
		assert.NotNil(t, v)
		assert.Equal(t, "required", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		condition func(*Context) bool
		found     bool
		nullable  bool
		want      bool
	}{
		{value: "string", found: true, want: true, condition: alwaysRequired},
		{value: "", found: true, want: true, condition: alwaysRequired},
		{value: 'a', found: true, want: true, condition: alwaysRequired},
		{value: 2, found: true, want: true, condition: alwaysRequired},
		{value: 2.5, found: true, want: true, condition: alwaysRequired},
		{value: []string{"string"}, found: true, want: true, condition: alwaysRequired},
		{value: map[string]any{"a": 1}, found: true, want: true, condition: alwaysRequired},
		{value: true, found: true, want: true, condition: alwaysRequired},
		{value: nil, found: true, want: false, condition: alwaysRequired},
		{value: nil, found: false, want: false, condition: alwaysRequired},
		{value: nil, found: true, want: true, nullable: true, condition: alwaysRequired},
		{value: nil, found: false, want: false, nullable: true, condition: alwaysRequired},
		{value: nil, found: true, want: true, condition: neverRequired},
		{value: nil, found: false, want: true, condition: neverRequired},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := RequiredIf(c.condition)
			ctx := &Context{
				Value: c.value,
				Field: &Field{
					isNullable: c.nullable,
				},
				Found: c.found,
			}

			assert.Equal(t, c.want, v.Validate(ctx))
		})
	}
}
