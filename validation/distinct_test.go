package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDistinctValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Distinct[string]()
		assert.NotNil(t, v)
		assert.Equal(t, "distinct", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		validator Validator
		want      bool
	}{
		{value: []string{"a", "b"}, validator: Distinct[string](), want: true},
		{value: []string{"a", "b", "a"}, validator: Distinct[string](), want: false},
		{value: []int{1, 2}, validator: Distinct[string](), want: false},
		{value: []int{1, 2}, validator: Distinct[int](), want: true},
		{value: []int{1, 2, 1}, validator: Distinct[int](), want: false},
		{value: []float64{0.3, 2}, validator: Distinct[float64](), want: true},
		{value: []float64{0.3, 2, 0.3}, validator: Distinct[float64](), want: false},
		{value: "string", validator: Distinct[string](), want: false},
		{value: "", validator: Distinct[string](), want: false},
		{value: 'a', validator: Distinct[rune](), want: false},
		{value: 2, validator: Distinct[int](), want: false},
		{value: 2.5, validator: Distinct[float64](), want: false},
		{value: map[string]any{"a": 1}, validator: Distinct[string](), want: false},
		{value: true, validator: Distinct[bool](), want: false},
		{value: nil, validator: Distinct[string](), want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			assert.Equal(t, c.want, c.validator.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}
