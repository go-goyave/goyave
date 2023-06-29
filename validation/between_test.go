package validation

import (
	"fmt"
	"math"
	"mime/multipart"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/util/fsutil"
)

func TestBetweenValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Between(1.5, 3.5)
		assert.NotNil(t, v)
		assert.Equal(t, "between", v.Name())
		assert.False(t, v.IsType())
		assert.True(t, v.IsTypeDependent())
		assert.Equal(t, []string{":min", "1.5", ":max", "3.5"}, v.MessagePlaceholders(&Context{}))
	})

	smallFile := fsutil.File{
		Header: &multipart.FileHeader{
			Size: 0.5 * 1024,
		},
	}
	mediumFile := fsutil.File{
		Header: &multipart.FileHeader{
			Size: 2 * 1024,
		},
	}
	largeFile := fsutil.File{
		Header: &multipart.FileHeader{
			Size: 4 * 1024,
		},
	}

	cases := []struct {
		value any
		want  bool
		min   float64
		max   float64
	}{
		{value: "a", want: false},
		{value: "abc", want: true},
		{value: "abcd", want: false},
		{value: 0, want: false},
		{value: 0.0, want: false},
		{value: 2, want: true},
		{value: 2.5, want: true},
		{value: 1.5, want: true},
		{value: 3.5, want: true},
		{value: 4, want: false},
		{value: 4.5, want: false},
		{value: uint64(math.MaxInt64), max: math.MaxInt64, want: false}, // overflow
		{value: uint(math.MaxInt64), max: math.MaxInt64, want: false},   // overflow
		{value: []string{"string"}, want: false},
		{value: []string{"a", "b", "c"}, want: true},
		{value: []string{"a", "b", "c", "d"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: map[string]any{"a": 1, "b": 2, "c": 3}, want: true},
		{value: map[string]any{"a": 1, "b": 2, "c": 3, "d": 4}, want: false},
		{value: []fsutil.File{smallFile}, want: false},
		{value: []fsutil.File{smallFile, mediumFile}, want: false},
		{value: []fsutil.File{largeFile}, want: false},
		{value: []fsutil.File{mediumFile, largeFile}, want: false},
		{value: []fsutil.File{smallFile, mediumFile, largeFile}, want: false},
		{value: []fsutil.File{mediumFile}, want: true},
		{value: 'A', want: false}, // A rune is an int32, A is 65 then

		// Cannot validate the following types
		{value: true, want: true},
		{value: time.Now(), want: true},
		{value: nil, want: true},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			min := 1.5
			max := 3.5
			if c.min != 0 {
				min = c.min
			}
			if c.max != 0 {
				max = c.max
			}
			v := Between(min, max)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}
