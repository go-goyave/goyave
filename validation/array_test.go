package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/util/fsutil"
)

func TestArrayValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Array()
		assert.NotNil(t, v)
		assert.Equal(t, "array", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		parent    any
		want      bool
	}{
		{value: []string{"a", "b"}, want: true, wantValue: []string{"a", "b"}},
		{value: []any{"a", "b"}, want: true, wantValue: []string{"a", "b"}},
		{value: []any{"a", 2}, want: true, wantValue: []any{"a", 2}},
		{value: []any{1, 2}, want: true, wantValue: []int{1, 2}},
		{value: []any{1, nil, 3}, want: true, wantValue: []any{1, nil, 3}}, // No conversion because one element is nil
		{value: []any{nil, 2, 3}, want: true, wantValue: []any{nil, 2, 3}}, // First element is nil (invalid)
		{value: []any{0, 1, 2}, want: true, wantValue: []int{0, 1, 2}},     // First element is zero value but not nil
		{value: []any{[]string{"a", "b"}}, want: true, wantValue: [][]string{{"a", "b"}}},
		{value: []any{[]any{"a", "b"}}, want: true, wantValue: [][]any{{"a", "b"}}},
		{value: []any{"a", "b"}, want: true, wantValue: []string{"a", "b"}, parent: []any{[]any{"a", "b"}}},
		{value: []string{"a", "b"}, want: true, wantValue: []string{"a", "b"}, parent: []any{[]any{"a", "b"}}},
		{value: []string{}, want: true, wantValue: []string{}, parent: []any{[]string{}}},
		{value: []any{1, 2, 3}, want: true, wantValue: []any{1, 2, 3}, parent: []string{}}, // Child element not assignable to parent
		{value: "a", want: false, wantValue: "a"},
		{value: 'a', want: false, wantValue: 'a'},
		{value: 2, want: false, wantValue: 2},
		{value: 2.5, want: false, wantValue: 2.5},
		{value: map[string]any{"a": 1}, want: false, wantValue: map[string]any{"a": 1}},
		{value: true, want: false, wantValue: true},
		{value: []fsutil.File{{}}, want: false, wantValue: []fsutil.File{{}}},
		{value: nil, want: false, wantValue: nil},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Array()
			var parent any = map[string]any{CurrentElement: c.value}
			if c.parent != nil {
				parent = c.parent
			}
			ctx := &Context{
				Value:  c.value,
				Parent: parent,
			}
			ok := v.Validate(ctx)
			if assert.Equal(t, c.want, ok) && ok {
				assert.Equal(t, c.wantValue, ctx.Value)
			}
		})
	}
}
