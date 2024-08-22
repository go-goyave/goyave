package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlphaValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Alpha()
		assert.NotNil(t, v)
		assert.Equal(t, "alpha", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":regexp", alphaRegex.String()}, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: "abc", want: true},
		{value: "abcéèçàùµêë", want: true},
		{value: "", want: false},
		{value: "abc123", want: false},
		{value: "abc_", want: false},
		{value: "abc-", want: false},
		{value: "abc.", want: false},
		{value: "abc&", want: false},
		{value: "abc~", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Alpha()
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}

func TestAlphaNumValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := AlphaNum()
		assert.NotNil(t, v)
		assert.Equal(t, "alpha_num", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":regexp", alphaNumRegex.String()}, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: "abc", want: true},
		{value: "abc123", want: true},
		{value: "abcéèçàùµêë123456789", want: true},
		{value: "", want: false},
		{value: "abc123_", want: false},
		{value: "abc123-", want: false},
		{value: "abc123.", want: false},
		{value: "abc123&", want: false},
		{value: "abc123~", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := AlphaNum()
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}

func TestAlphaDashValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := AlphaDash()
		assert.NotNil(t, v)
		assert.Equal(t, "alpha_dash", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":regexp", alphaDashRegex.String()}, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: "abc", want: true},
		{value: "abc123-_", want: true},
		{value: "abcéèçàùµêë123456789-_", want: true},
		{value: "", want: false},
		{value: "abc123-_.", want: false},
		{value: "abc123-_&", want: false},
		{value: "abc123-_~", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := AlphaDash()
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}
