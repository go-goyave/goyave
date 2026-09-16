package validation

import (
	"context"
	"fmt"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestUniqueValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Unique(func(_ context.Context, _ string) (bool, error) { return true, nil })
		assert.NotNil(t, v)
		assert.Equal(t, "unique", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
		assert.NotNil(t, v.fn)
	})

	cases := []struct {
		desc           string
		fn             func(context.Context, string) (bool, error)
		value          any
		valid          bool
		expected       bool
		expectedErrors []string
	}{
		{
			desc: "OK",
			fn: func(_ context.Context, _ string) (bool, error) {
				return true, nil
			},
			value:          "johndoe",
			valid:          true,
			expected:       true,
			expectedErrors: []string{},
		},
		{
			desc: "NOK",
			fn: func(_ context.Context, _ string) (bool, error) {
				return false, nil
			},
			value:          "johndoe",
			valid:          true,
			expected:       false,
			expectedErrors: []string{},
		},
		{
			desc: "error",
			fn: func(_ context.Context, _ string) (bool, error) {
				return false, fmt.Errorf("test error")
			},
			value:          "johndoe",
			valid:          true,
			expected:       false,
			expectedErrors: []string{"test error"},
		},
		{
			desc:           "ctx_invalid",
			value:          "johndoe",
			valid:          false,
			expected:       true,
			expectedErrors: []string{},
		},
		{
			desc:           "incorrect_type",
			value:          123,
			valid:          true,
			expected:       true,
			expectedErrors: []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			v := Unique(c.fn)
			ctx := &Context{
				Invalid: !c.valid,
				Value:   c.value,
			}
			assert.Equal(t, c.expected, v.Validate(ctx))
			assert.Equal(t, c.expectedErrors, lo.Map(ctx.errors, func(e error, _ int) string { return e.Error() }))
		})
	}
}

func TestExistsValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Exists(func(_ context.Context, _ string) (bool, error) { return true, nil })
		assert.NotNil(t, v)
		assert.Equal(t, "exists", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
		assert.NotNil(t, v.fn)
	})

	cases := []struct {
		desc           string
		fn             func(context.Context, string) (bool, error)
		value          any
		valid          bool
		expected       bool
		expectedErrors []string
	}{
		{
			desc: "OK",
			fn: func(_ context.Context, _ string) (bool, error) {
				return true, nil
			},
			value:          "johndoe",
			valid:          true,
			expected:       true,
			expectedErrors: []string{},
		},
		{
			desc: "NOK",
			fn: func(_ context.Context, _ string) (bool, error) {
				return false, nil
			},
			value:          "johndoe",
			valid:          true,
			expected:       false,
			expectedErrors: []string{},
		},
		{
			desc: "error",
			fn: func(_ context.Context, _ string) (bool, error) {
				return false, fmt.Errorf("test error")
			},
			value:          "johndoe",
			valid:          true,
			expected:       false,
			expectedErrors: []string{"test error"},
		},
		{
			desc:           "ctx_invalid",
			value:          "johndoe",
			valid:          false,
			expected:       true,
			expectedErrors: []string{},
		},
		{
			desc:           "incorrect_type",
			value:          123,
			valid:          true,
			expected:       true,
			expectedErrors: []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			v := Exists(c.fn)
			ctx := &Context{
				Invalid: !c.valid,
				Value:   c.value,
			}
			assert.Equal(t, c.expected, v.Validate(ctx))
			assert.Equal(t, c.expectedErrors, lo.Map(ctx.errors, func(e error, _ int) string { return e.Error() }))
		})
	}
}

func TestUniqueArrayValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := UniqueArray(func(_ context.Context, _ []int) ([]int, error) { return []int{}, nil })
		assert.NotNil(t, v)
		assert.Equal(t, "unique", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
		assert.NotEmpty(t, v.fn)
	})

	cases := []struct {
		desc                       string
		fn                         func(context.Context, []int) ([]int, error)
		value                      any
		expectedErrors             []string
		expectedArrayElementErrors []int
		valid                      bool
	}{
		{
			desc:                       "OK",
			fn:                         func(_ context.Context, _ []int) ([]int, error) { return []int{}, nil },
			value:                      []int{7, 5},
			valid:                      true,
			expectedErrors:             []string{},
			expectedArrayElementErrors: nil,
		},
		{
			desc:                       "NOK",
			fn:                         func(_ context.Context, _ []int) ([]int, error) { return []int{1, 2}, nil },
			value:                      []int{7, 5, 4},
			valid:                      true,
			expectedErrors:             []string{},
			expectedArrayElementErrors: []int{1, 2},
		},
		{
			desc:                       "error",
			fn:                         func(_ context.Context, _ []int) ([]int, error) { return nil, fmt.Errorf("test error") },
			value:                      []int{7, 5, 4},
			valid:                      true,
			expectedErrors:             []string{"test error"},
			expectedArrayElementErrors: nil,
		},
		{
			desc:                       "ctx_invalid",
			fn:                         func(_ context.Context, _ []int) ([]int, error) { return []int{1, 2}, nil },
			value:                      []int{7, 5, 4},
			valid:                      false,
			expectedErrors:             []string{},
			expectedArrayElementErrors: nil,
		},
		{
			desc:                       "incorrect_type",
			fn:                         func(_ context.Context, _ []int) ([]int, error) { return []int{1, 2}, nil },
			value:                      []float64{7, 5, 4},
			valid:                      true,
			expectedErrors:             []string{},
			expectedArrayElementErrors: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			v := UniqueArray(c.fn)

			ctx := &Context{
				Invalid: !c.valid,
				Value:   c.value,
			}
			assert.True(t, v.Validate(ctx)) // Always true, if there's an error, it will be on the array element.
			assert.Equal(t, c.expectedErrors, lo.Map(ctx.errors, func(e error, _ int) string { return e.Error() }))
			assert.Equal(t, c.expectedArrayElementErrors, ctx.arrayElementErrors)
		})
	}
}

func TestExistsArrayValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := ExistsArray(func(_ context.Context, _ []int) ([]int, error) { return []int{}, nil })
		assert.NotNil(t, v)
		assert.Equal(t, "exists", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
		assert.NotEmpty(t, v.fn)
	})

	cases := []struct {
		desc                       string
		fn                         func(context.Context, []int) ([]int, error)
		value                      any
		expectedErrors             []string
		expectedArrayElementErrors []int
		valid                      bool
	}{
		{
			desc:                       "OK",
			fn:                         func(_ context.Context, _ []int) ([]int, error) { return []int{}, nil },
			value:                      []int{7, 5},
			valid:                      true,
			expectedErrors:             []string{},
			expectedArrayElementErrors: nil,
		},
		{
			desc:                       "NOK",
			fn:                         func(_ context.Context, _ []int) ([]int, error) { return []int{1, 2}, nil },
			value:                      []int{7, 5, 4},
			valid:                      true,
			expectedErrors:             []string{},
			expectedArrayElementErrors: []int{1, 2},
		},
		{
			desc:                       "error",
			fn:                         func(_ context.Context, _ []int) ([]int, error) { return nil, fmt.Errorf("test error") },
			value:                      []int{7, 5, 4},
			valid:                      true,
			expectedErrors:             []string{"test error"},
			expectedArrayElementErrors: nil,
		},
		{
			desc:                       "ctx_invalid",
			fn:                         func(_ context.Context, _ []int) ([]int, error) { return []int{1, 2}, nil },
			value:                      []int{7, 5, 4},
			valid:                      false,
			expectedErrors:             []string{},
			expectedArrayElementErrors: nil,
		},
		{
			desc:                       "incorrect_type",
			fn:                         func(_ context.Context, _ []int) ([]int, error) { return []int{1, 2}, nil },
			value:                      []float64{7, 5, 4},
			valid:                      true,
			expectedErrors:             []string{},
			expectedArrayElementErrors: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			v := ExistsArray(c.fn)

			ctx := &Context{
				Invalid: !c.valid,
				Value:   c.value,
			}
			assert.True(t, v.Validate(ctx)) // Always true, if there's an error, it will be on the array element.
			assert.Equal(t, c.expectedErrors, lo.Map(ctx.errors, func(e error, _ int) string { return e.Error() }))
			assert.Equal(t, c.expectedArrayElementErrors, ctx.arrayElementErrors)
		})
	}
}
