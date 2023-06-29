package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/lang"
)

func TestDifferentValidator(t *testing.T) {
	path := "object.field[]"
	t.Run("Constructor", func(t *testing.T) {
		v := Different(path)
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "different", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":other", "field"}, v.MessagePlaceholders(&Context{}))

		assert.Panics(t, func() {
			Different("invalid[path.")
		})
	})

	cases := []struct {
		value any
		data  map[string]any
		desc  string
		want  bool
	}{
		{desc: "equal strings", data: makeComparisonData("abc"), value: "abc", want: false},
		{desc: "different strings", data: makeComparisonData("ab"), value: "abc", want: true},
		{desc: "many different strings", data: makeComparisonData("ab", "cd"), value: "abc", want: true},
		{desc: "many equal strings", data: makeComparisonData("abc", "cde"), value: "abc", want: false},
		{desc: "equal bool", data: makeComparisonData(true), value: true, want: false},
		{desc: "different bool", data: makeComparisonData(false), value: true, want: true},
		{desc: "many different bool", data: makeComparisonData(false, false), value: true, want: true},
		{desc: "many equal bool", data: makeComparisonData(false, true), value: true, want: false},
		{desc: "equal numbers", data: makeComparisonData(1), value: 1, want: false},
		{desc: "equal numbers different type", data: makeComparisonData(1), value: 1.0, want: true},
		{desc: "different numbers", data: makeComparisonData(1), value: 2, want: true},
		{desc: "different numbers different types", data: makeComparisonData(1), value: 1.1, want: true},
		{desc: "many different numbers", data: makeComparisonData(0.3, 2.1), value: 0.4, want: true},
		{desc: "many equal numbers", data: makeComparisonData(1, 2.1, 3, 0.3), value: 0.3, want: false},
		{desc: "equal arrays", data: makeComparisonData([]string{"a", "b"}), value: []string{"a", "b"}, want: false},
		{desc: "different arrays", data: makeComparisonData([]string{"a", "b"}), value: []string{"b", "a"}, want: true},
		{desc: "different arrays", data: makeComparisonData([]string{"a", "b"}), value: []string{"b", "a", "z"}, want: true},
		{desc: "many different arrays", data: makeComparisonData([]string{"b", "a"}, []string{}, []any{"a", "b"}), value: []string{"a", "b"}, want: true},
		{desc: "many equal arrays", data: makeComparisonData([]string{"a", "b"}, []string{"b", "a"}), value: []string{"a", "b"}, want: false},
		{desc: "equal objects", data: makeComparisonData(map[string]any{"a": 0}), value: map[string]any{"a": 0}, want: false},
		{desc: "different objects", data: makeComparisonData(map[string]any{"a": 1}), value: map[string]any{"a": 0}, want: true},
		{desc: "many different objects", data: makeComparisonData(map[string]any{"a": 1}, map[string]any{"b": 0}), value: map[string]any{"a": 0}, want: true},
		{desc: "many equal objects", data: makeComparisonData(map[string]any{"a": 0}, map[string]any{"a": 1}, map[string]any{"b": 0}), value: map[string]any{"a": 0}, want: false},

		{desc: "nil with nil", data: makeComparisonData(nil), value: nil, want: true},
		{desc: "string with nil", data: makeComparisonData(nil), value: "abc", want: true},
		{desc: "nil with string", data: makeComparisonData("ab"), value: nil, want: true},
		{desc: "not found", data: map[string]any{"object": map[string]any{}}, value: "abc", want: false},
		{desc: "empty array", data: makeComparisonData(), value: "abc", want: true},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			v := Different(path)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
				Data:  c.data,
			}))
		})
	}
}
