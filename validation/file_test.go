package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/util/fsutil"
)

func TestFileValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := File()
		assert.NotNil(t, v)
		assert.Equal(t, "file", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: []fsutil.File{}, want: true},
		{value: fsutil.File{}, want: false},
		{value: "string", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: struct{}{}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := File()
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}

func TestFileCountValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := FileCount(5)
		assert.NotNil(t, v)
		assert.Equal(t, "file_count", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":value", "5"}, v.MessagePlaceholders(&Context{}))
		assert.Equal(t, uint(5), v.Count)
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: []fsutil.File{{}, {}, {}, {}}, want: false},
		{value: []fsutil.File{{}, {}, {}}, want: true},
		{value: []fsutil.File{{}, {}}, want: false},
		{value: []fsutil.File{{}}, want: false},
		{value: []fsutil.File{}, want: false},
		{value: fsutil.File{}, want: false},
		{value: "string", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: struct{}{}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := FileCount(3)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}

func TestMinFileCountValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := MinFileCount(5)
		assert.NotNil(t, v)
		assert.Equal(t, "min_file_count", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":min", "5"}, v.MessagePlaceholders(&Context{}))
		assert.Equal(t, uint(5), v.Min)
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: []fsutil.File{{}, {}, {}, {}}, want: true},
		{value: []fsutil.File{{}, {}, {}}, want: true},
		{value: []fsutil.File{{}, {}}, want: false},
		{value: []fsutil.File{{}}, want: false},
		{value: []fsutil.File{}, want: false},
		{value: fsutil.File{}, want: false},
		{value: "string", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: struct{}{}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := MinFileCount(3)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}

func TestMaxFileCountValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := MaxFileCount(5)
		assert.NotNil(t, v)
		assert.Equal(t, "max_file_count", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":max", "5"}, v.MessagePlaceholders(&Context{}))
		assert.Equal(t, uint(5), v.Max)
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: []fsutil.File{{}, {}, {}, {}}, want: false},
		{value: []fsutil.File{{}, {}, {}}, want: true},
		{value: []fsutil.File{{}, {}}, want: true},
		{value: []fsutil.File{{}}, want: true},
		{value: []fsutil.File{}, want: true},
		{value: fsutil.File{}, want: false},
		{value: "string", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: struct{}{}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := MaxFileCount(3)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}

func TestFileCountBetweenValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := FileCountBetween(2, 3)
		assert.NotNil(t, v)
		assert.Equal(t, "file_count_between", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":min", "2", ":max", "3"}, v.MessagePlaceholders(&Context{}))
		assert.Equal(t, uint(2), v.Min)
		assert.Equal(t, uint(3), v.Max)
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: []fsutil.File{{}, {}, {}, {}}, want: false},
		{value: []fsutil.File{{}, {}, {}}, want: true},
		{value: []fsutil.File{{}, {}}, want: true},
		{value: []fsutil.File{{}}, want: false},
		{value: []fsutil.File{}, want: false},
		{value: fsutil.File{}, want: false},
		{value: "string", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: struct{}{}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := FileCountBetween(2, 3)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}
