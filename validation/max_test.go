package validation

import (
	"fmt"
	"math"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/util/fsutil"
)

func TestMaxValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Max(123.456)
		assert.NotNil(t, v)
		assert.Equal(t, "max", v.Name())
		assert.False(t, v.IsType())
		assert.True(t, v.IsTypeDependent())
		assert.Equal(t, []string{":max", "123.456"}, v.MessagePlaceholders(&Context{}))
	})

	file := fsutil.File{Header: &multipart.FileHeader{Size: 2048}}
	largeFile := fsutil.File{Header: &multipart.FileHeader{Size: math.MaxInt64}}

	cases := []struct {
		value any
		max   float64
		want  bool
	}{
		{value: 2, want: false, max: 1},
		{value: 2.5, want: false, max: 1},
		{value: float32(2.5), want: false, max: 1},
		{value: float64(math.MaxInt64), want: true, max: math.MaxInt64},
		{value: float64(math.MinInt64), want: true, max: math.MaxInt64},
		{value: int64(math.MaxInt64), want: false, max: math.MaxInt64}, // Don't pass because above max int value that can accurately fit in float64
		{value: int64(math.MinInt64), want: false, max: math.MaxInt64}, // Don't pass because below min int value that can accurately fit in float64
		{value: 'a', want: false, max: 1},
		{value: "string", want: false, max: 3},
		{value: []string{"a", "b"}, want: false, max: 1},
		{value: map[string]any{"a": 1, "b": 2}, want: false, max: 1},
		{value: true, want: true, max: 1},
		{value: nil, want: true, max: 1},
		{value: []fsutil.File{file}, want: false, max: 1},
		{value: []fsutil.File{largeFile}, want: false, max: math.MaxInt64}, // Don't pass because above max int value that can accurately fit in float64

		{value: 2, want: true, max: 3},
		{value: 2.5, want: true, max: 3},
		{value: float32(2.5), want: true, max: 3},
		{value: 'a', want: true, max: 100},
		{value: "abc", want: true, max: 3},
		{value: []string{"a", "b", "c"}, want: true, max: 3},
		{value: map[string]any{"a": 1, "b": 2, "c": 3}, want: true, max: 3},
		{value: []fsutil.File{file}, want: true, max: 3},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%f_%t", c.value, c.max, c.want), func(t *testing.T) {
			v := Max(c.max)
			ctx := &Context{
				Value: c.value,
			}
			ok := v.Validate(ctx)
			assert.Equal(t, c.want, ok)
		})
	}
}
