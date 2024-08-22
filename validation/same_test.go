package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/lang"
)

func TestSameValidator(t *testing.T) {
	path := "object.field[]"
	t.Run("Constructor", func(t *testing.T) {
		v := Same(path)
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "same", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":other", "field"}, v.MessagePlaceholders(&Context{}))

		assert.Panics(t, func() {
			Same("invalid[path.")
		})
	})

	cases := []struct {
		value any
		data  map[string]any
		desc  string
		want  bool
	}{
		{desc: "equal strings", data: makeComparisonData("abc"), value: "abc", want: true},
		{desc: "different strings", data: makeComparisonData("ab"), value: "abc", want: false},
		{desc: "many different strings", data: makeComparisonData("abc", "cd"), value: "abc", want: false},
		{desc: "many equal strings", data: makeComparisonData("abc", "abc"), value: "abc", want: true},
		{desc: "equal bool", data: makeComparisonData(true), value: true, want: true},
		{desc: "different bool", data: makeComparisonData(false), value: true, want: false},
		{desc: "many different bool", data: makeComparisonData(false, true), value: true, want: false},
		{desc: "many equal bool", data: makeComparisonData(true, true), value: true, want: true},
		{desc: "equal numbers", data: makeComparisonData(1), value: 1, want: true},
		{desc: "equal numbers different type", data: makeComparisonData(1), value: 1.0, want: false},
		{desc: "different numbers", data: makeComparisonData(1), value: 2, want: false},
		{desc: "different numbers different types", data: makeComparisonData(1), value: 1.1, want: false},
		{desc: "many different numbers", data: makeComparisonData(0.3, 2.1), value: 0.3, want: false},
		{desc: "many equal numbers", data: makeComparisonData(0.3, 0.3), value: 0.3, want: true},
		{desc: "equal arrays", data: makeComparisonData([]string{"a", "b"}), value: []string{"a", "b"}, want: true},
		{desc: "different arrays", data: makeComparisonData([]string{"a", "b"}), value: []string{"b", "a"}, want: false},
		{desc: "different arrays", data: makeComparisonData([]string{"a", "b"}), value: []string{"b", "a", "z"}, want: false},
		{desc: "many different arrays", data: makeComparisonData([]string{"a", "b"}, []string{}, []any{"a", "b"}), value: []string{"a", "b"}, want: false},
		{desc: "many equal arrays", data: makeComparisonData([]string{"a", "b"}, []string{"a", "b"}), value: []string{"a", "b"}, want: true},
		{desc: "equal objects", data: makeComparisonData(map[string]any{"a": 0}), value: map[string]any{"a": 0}, want: true},
		{desc: "different objects", data: makeComparisonData(map[string]any{"a": 1}), value: map[string]any{"a": 0}, want: false},
		{desc: "many different objects", data: makeComparisonData(map[string]any{"a": 1}, map[string]any{"a": 0}), value: map[string]any{"a": 0}, want: false},
		{desc: "many equal objects", data: makeComparisonData(map[string]any{"a": 0}, map[string]any{"a": 0}), value: map[string]any{"a": 0}, want: true},

		{desc: "nil with nil", data: makeComparisonData(nil), value: nil, want: false},
		{desc: "string with nil", data: makeComparisonData(nil), value: "abc", want: false},
		{desc: "nil with string", data: makeComparisonData("ab"), value: nil, want: false},
		{desc: "not found", data: map[string]any{"object": map[string]any{}}, value: "abc", want: false},
		{desc: "empty array", data: makeComparisonData(), value: "abc", want: true},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			v := Same(path)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
				Data:  c.data,
			}))
		})
	}
}
