package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeysInValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := KeysIn("a", "b", "c")
		assert.NotNil(t, v)
		assert.Equal(t, "keys_in", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":values", "a, b, c"}, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		input       any
		allowedKeys []string
		want        bool
	}{
		{allowedKeys: []string{"a", "b", "c"}, input: map[string]any{"a": 1, "b": 2, "c": 3}, want: true},
		{allowedKeys: []string{"a", "b", "c"}, input: map[string]any{"a": 1, "b": 2, "c": 3, "d": 4}, want: false},
		{allowedKeys: []string{"a", "b", "c"}, input: map[string]any{"a": 1, "b": 2}, want: true},
		{allowedKeys: []string{"a", "b", "c"}, input: map[string]any{}, want: true},
		{allowedKeys: []string{"a", "b", "c"}, input: "", want: false},
		{allowedKeys: []string{"a", "b", "c"}, input: 'a', want: false},
		{allowedKeys: []string{"a", "b", "c"}, input: 2, want: false},
		{allowedKeys: []string{"a", "b", "c"}, input: 2.5, want: false},
		{allowedKeys: []string{"a", "b", "c"}, input: true, want: false},
		{allowedKeys: []string{"a", "b", "c"}, input: nil, want: false},
		{allowedKeys: []string{"a", "b", "c"}, input: (map[string]any)(nil), want: true},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", tc.input, tc.want), func(t *testing.T) {
			v := KeysIn(tc.allowedKeys...)
			assert.Equal(t, tc.want, v.Validate(&Context{
				Value: tc.input,
			}))
		})
	}
}
