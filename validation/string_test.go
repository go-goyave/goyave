package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := String()
		assert.NotNil(t, v)
		assert.Equal(t, "string", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&ContextV5{}))
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: "string", want: true},
		{value: "", want: true},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := String()
			assert.Equal(t, c.want, v.Validate(&ContextV5{
				Value: c.value,
			}))
		})
	}
}

func TestStartsWithValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := StartsWith("pre", "fix")
		assert.NotNil(t, v)
		assert.Equal(t, "starts_with", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":values", "pre, fix"}, v.MessagePlaceholders(&ContextV5{}))
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: "string", want: false},
		{value: "pre-string", want: true},
		{value: "fix-string", want: true},
		{value: "string-pre", want: false},
		{value: "string-fix", want: false},
		{value: "", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: []string{"pre-string"}, want: false},
		{value: []string{"fix-string"}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := StartsWith("pre", "fix")
			assert.Equal(t, c.want, v.Validate(&ContextV5{
				Value: c.value,
			}))
		})
	}
}

func TestEndsWithValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := EndsWith("suf", "fix")
		assert.NotNil(t, v)
		assert.Equal(t, "ends_with", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":values", "suf, fix"}, v.MessagePlaceholders(&ContextV5{}))
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: "string", want: false},
		{value: "suf-string", want: false},
		{value: "fix-string", want: false},
		{value: "string-suf", want: true},
		{value: "string-fix", want: true},
		{value: "", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: []string{"string-suf"}, want: false},
		{value: []string{"string-fix"}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := EndsWith("suf", "fix")
			assert.Equal(t, c.want, v.Validate(&ContextV5{
				Value: c.value,
			}))
		})
	}
}
