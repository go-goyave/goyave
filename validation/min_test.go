package validation

import (
	"fmt"
	"math"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/util/fsutil"
)

func TestMinalidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Min(123.456)
		assert.NotNil(t, v)
		assert.Equal(t, "min", v.Name())
		assert.False(t, v.IsType())
		assert.True(t, v.IsTypeDependent())
		assert.Equal(t, []string{":min", "123.456"}, v.MessagePlaceholders(&Context{}))
	})

	file := fsutil.File{Header: &multipart.FileHeader{Size: 2048}}
	largeFile := fsutil.File{Header: &multipart.FileHeader{Size: math.MaxInt64}}

	cases := []struct {
		value any
		min   float64
		want  bool
	}{
		{value: 2, want: false, min: 3},
		{value: 2.5, want: false, min: 3},
		{value: float32(2.5), want: false, min: 3},
		{value: float64(math.MaxInt64), want: true, min: math.MinInt64},
		{value: float64(math.MinInt64), want: true, min: math.MinInt64},
		{value: int64(math.MaxInt64), want: false, min: math.MinInt64}, // Don't pass because above max int value that can accurately fit in float64
		{value: int64(math.MinInt64), want: false, min: math.MinInt64}, // Don't pass because below min int value that can accurately fit in float64
		{value: 'a', want: false, min: 100},
		{value: "abc", want: false, min: 4},
		{value: []string{"a", "b"}, want: false, min: 3},
		{value: map[string]any{"a": 1, "b": 2}, want: false, min: 3},
		{value: true, want: true, min: 1},
		{value: nil, want: true, min: 1},
		{value: []fsutil.File{file}, want: false, min: 3},
		{value: []fsutil.File{largeFile}, want: false, min: math.MinInt64}, // Don't pass because above max int value that can accurately fit in float64

		{value: 3, want: true, min: 3},
		{value: 3.5, want: true, min: 3},
		{value: float32(3.5), want: true, min: 3},
		{value: 'a', want: true, min: 97},
		{value: "string", want: true, min: 3},
		{value: []string{"a", "b", "c"}, want: true, min: 3},
		{value: map[string]any{"a": 1, "b": 2, "c": 3}, want: true, min: 3},
		{value: []fsutil.File{file}, want: true, min: 2},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%f_%t", c.value, c.min, c.want), func(t *testing.T) {
			v := Min(c.min)
			ctx := &Context{
				Value: c.value,
			}
			ok := v.Validate(ctx)
			assert.Equal(t, c.want, ok)
		})
	}
}
