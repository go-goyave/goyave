package validation

import (
	"fmt"
	"math"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/util/fsutil"
)

func TestSizeValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Size(123)
		assert.NotNil(t, v)
		assert.Equal(t, "size", v.Name())
		assert.False(t, v.IsType())
		assert.True(t, v.IsTypeDependent())
		assert.Equal(t, []string{":value", "123"}, v.MessagePlaceholders(&Context{}))
	})

	file := fsutil.File{Header: &multipart.FileHeader{Size: 2048}}
	largeFile := fsutil.File{Header: &multipart.FileHeader{Size: math.MaxInt64}}

	cases := []struct {
		value any
		size  int
		want  bool
	}{
		{value: 2, want: true, size: 2},
		{value: 2.5, want: true, size: 2},
		{value: float32(2.5), want: true, size: 2},
		{value: 'a', want: true, size: 2},
		{value: "string", want: false, size: 3},
		{value: "string", want: true, size: 6},
		{value: []string{"a", "b"}, want: false, size: 3},
		{value: []string{"a", "b", "c"}, want: true, size: 3},
		{value: map[string]any{"a": 1, "b": 2}, want: false, size: 3},
		{value: map[string]any{"a": 1, "b": 2, "c": 3}, want: true, size: 3},
		{value: true, want: true, size: 1},
		{value: nil, want: true, size: 1},
		{value: []fsutil.File{file}, want: false, size: 1},
		{value: []fsutil.File{file}, want: true, size: 2},
		{value: []fsutil.File{largeFile}, want: false, size: math.MaxInt64}, // Don't pass because above max int value that can accurately fit in float64
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%d_%t", c.value, c.size, c.want), func(t *testing.T) {
			v := Size(c.size)
			ctx := &Context{
				Value: c.value,
			}
			ok := v.Validate(ctx)
			assert.Equal(t, c.want, ok)
		})
	}
}
