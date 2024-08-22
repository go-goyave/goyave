package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDigitValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Digits()
		assert.NotNil(t, v)
		assert.Equal(t, "digits", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":regexp", digitsRegex.String()}, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: "0123456789", want: true},
		{value: "abc123", want: false},
		{value: "string", want: false},
		{value: "", want: true},
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
			v := Digits()
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}
