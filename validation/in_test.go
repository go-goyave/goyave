package validation

import (
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/lang"
)

type inTestCase[T comparable] struct {
	value  any
	values []T
	want   bool
}

func TestInValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := In([]string{"a", "b", "c"})
		assert.NotNil(t, v)
		assert.Equal(t, "in", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":values", "a, b, c"}, v.MessagePlaceholders(&Context{}))
	})

	cases := []inTestCase[string]{
		{value: "a", want: true, values: []string{"a", "b", "c"}},
		{value: "b", want: true, values: []string{"a", "b", "c"}},
		{value: "c", want: true, values: []string{"a", "b", "c"}},
		{value: "d", want: false, values: []string{"a", "b", "c"}},
		{value: 1, want: false, values: []string{"a", "b", "c"}},
		{value: 1.2, want: false, values: []string{"a", "b", "c"}},
		{value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:06:42Z")), want: false, values: []string{"a", "b", "c"}},
		{value: "string", want: false, values: []string{"a", "b", "c"}},
		{value: []string{"string"}, want: false, values: []string{"a", "b", "c"}},
		{value: map[string]any{"a": 1}, want: false, values: []string{"a", "b", "c"}},
		{value: true, want: false, values: []string{"a", "b", "c"}},
		{value: nil, want: false, values: []string{"a", "b", "c"}},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := In(c.values)
			ctx := &Context{
				Value: c.value,
			}
			assert.Equal(t, c.want, v.Validate(ctx))
		})
	}
}

func TestNotInValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := NotIn([]string{"a", "b", "c"})
		assert.NotNil(t, v)
		assert.Equal(t, "not_in", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":values", "a, b, c"}, v.MessagePlaceholders(&Context{}))
	})

	cases := []inTestCase[string]{
		{value: "a", want: false, values: []string{"a", "b", "c"}},
		{value: "b", want: false, values: []string{"a", "b", "c"}},
		{value: "c", want: false, values: []string{"a", "b", "c"}},
		{value: "d", want: true, values: []string{"a", "b", "c"}},
		{value: 1, want: false, values: []string{"a", "b", "c"}},
		{value: 1.2, want: false, values: []string{"a", "b", "c"}},
		{value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:06:42Z")), want: false, values: []string{"a", "b", "c"}},
		{value: []string{"string"}, want: false, values: []string{"a", "b", "c"}},
		{value: map[string]any{"a": 1}, want: false, values: []string{"a", "b", "c"}},
		{value: true, want: false, values: []string{"a", "b", "c"}},
		{value: nil, want: false, values: []string{"a", "b", "c"}},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := NotIn(c.values)
			ctx := &Context{
				Value: c.value,
			}
			assert.Equal(t, c.want, v.Validate(ctx))
		})
	}
}

func TestInFieldValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := InField[string]("field")
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "in_field", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":other", "field"}, v.MessagePlaceholders(&Context{}))

		assert.Panics(t, func() {
			InField[string](".path[")
		})
	})

	cases := []struct {
		value any
		field any
		want  bool
	}{
		{value: "a", want: true, field: []string{"a", "b", "c"}},
		{value: "a", want: false, field: []rune{'a', 'b', 'c'}},
		{value: "a", want: false, field: "a"},
		{value: "b", want: true, field: []string{"a", "b", "c"}},
		{value: "c", want: true, field: []string{"a", "b", "c"}},
		{value: "d", want: false, field: []string{"a", "b", "c"}},
		{value: 1, want: false, field: []string{"a", "b", "c"}},
		{value: 1.2, want: false, field: []string{"a", "b", "c"}},
		{value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:06:42Z")), want: false, field: []string{"a", "b", "c"}},
		{value: "string", want: false, field: []string{"a", "b", "c"}},
		{value: []string{"string"}, want: false, field: []string{"a", "b", "c"}},
		{value: map[string]any{"a": 1}, want: false, field: []string{"a", "b", "c"}},
		{value: true, want: false, field: []string{"a", "b", "c"}},
		{value: nil, want: false, field: []string{"a", "b", "c"}},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := InField[string]("field")
			v.lang = &lang.Language{}
			ctx := &Context{
				Data: map[string]any{
					"field": c.field,
				},
				Value: c.value,
			}
			assert.Equal(t, c.want, v.Validate(ctx))
		})
	}

	t.Run("Validate_n-dimensional_array_missing", func(t *testing.T) {
		v := InField[string]("field[]")
		ctx := &Context{
			Data: map[string]any{
				"field": [][]string{},
			},
			Value: "a",
		}
		assert.False(t, v.Validate(ctx))
	})
}

func TestNotInFieldValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := NotInField[string]("field")
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "not_in_field", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":other", "field"}, v.MessagePlaceholders(&Context{}))

		assert.Panics(t, func() {
			NotInField[string](".path[")
		})
	})

	cases := []struct {
		value any
		field any
		want  bool
	}{
		{value: "a", want: false, field: []string{"a", "b", "c"}},
		{value: "a", want: true, field: []rune{'a', 'b', 'c'}},
		{value: "a", want: true, field: "a"},
		{value: "b", want: false, field: []string{"a", "b", "c"}},
		{value: "c", want: false, field: []string{"a", "b", "c"}},
		{value: "d", want: true, field: []string{"a", "b", "c"}},
		{value: 1, want: false, field: []string{"a", "b", "c"}},
		{value: 1.2, want: false, field: []string{"a", "b", "c"}},
		{value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:06:42Z")), want: false, field: []string{"a", "b", "c"}},
		{value: []string{"string"}, want: false, field: []string{"a", "b", "c"}},
		{value: map[string]any{"a": 1}, want: false, field: []string{"a", "b", "c"}},
		{value: true, want: false, field: []string{"a", "b", "c"}},
		{value: nil, want: false, field: []string{"a", "b", "c"}},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := NotInField[string]("field")
			v.lang = &lang.Language{}
			ctx := &Context{
				Data: map[string]any{
					"field": c.field,
				},
				Value: c.value,
			}
			assert.Equal(t, c.want, v.Validate(ctx))
		})
	}

	t.Run("Validate_n-dimensional_array_missing", func(t *testing.T) {
		v := NotInField[string]("field[]")
		ctx := &Context{
			Data: map[string]any{
				"field": [][]string{},
			},
			Value: "a",
		}
		assert.True(t, v.Validate(ctx))
	})
}
