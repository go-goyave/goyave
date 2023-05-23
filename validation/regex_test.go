package validation

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Regex(regexp.MustCompile(`^[0-9]+$`))
		assert.NotNil(t, v)
		assert.Equal(t, "regex", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":regexp", `^[0-9]+$`}, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value any
		regex *regexp.Regexp
		want  bool
	}{
		{value: "string", want: false, regex: regexp.MustCompile(`^[0-9]+$`)},
		{value: "0123456789", want: true, regex: regexp.MustCompile(`^[0-9]+$`)},
		{value: "", want: true, regex: regexp.MustCompile(`^[0-9]*$`)},
		{value: "", want: false, regex: regexp.MustCompile(`^[0-9]+$`)},
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
			v := Regex(c.regex)
			ctx := &Context{
				Value: c.value,
			}
			assert.Equal(t, c.want, v.Validate(ctx))
		})
	}
}
