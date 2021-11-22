package sliceutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	assert.True(t, Contains([]string{"val1", "val2", "val3"}, "val2"))
	assert.False(t, Contains([]string{"val1", "val2", "val3"}, "val4"))
	assert.True(t, Contains([]int{1, 2, 3}, 2))
	assert.False(t, Contains([]int{1, 2, 3}, 4))
	assert.True(t, Contains([]float64{1, 2, 3}, float64(2)))
	assert.False(t, Contains([]float64{1, 2, 3}, 2))
	assert.False(t, Contains([]float64{1, 2, 3}, 4))
}

func TestIndexOf(t *testing.T) {
	assert.Equal(t, 1, IndexOf([]string{"val1", "val2", "val3"}, "val2"))
	assert.Equal(t, -1, IndexOf([]string{"val1", "val2", "val3"}, "val4"))
	assert.Equal(t, 1, IndexOf([]int{1, 2, 3}, 2))
	assert.Equal(t, -1, IndexOf([]int{1, 2, 3}, 4))
	assert.Equal(t, 1, IndexOf([]float64{1, 2, 3}, float64(2)))
	assert.Equal(t, -1, IndexOf([]float64{1, 2, 3}, 2))
	assert.Equal(t, -1, IndexOf([]float64{1, 2, 3}, 4))
}

func TestContainsStr(t *testing.T) {
	assert.True(t, ContainsStr([]string{"val1", "val2", "val3"}, "val2"))
	assert.False(t, ContainsStr([]string{"val1", "val2", "val3"}, "val4"))
}

func TestIndexOfStr(t *testing.T) {
	assert.Equal(t, 1, IndexOfStr([]string{"val1", "val2", "val3"}, "val2"))
	assert.Equal(t, -1, IndexOfStr([]string{"val1", "val2", "val3"}, "val4"))
}

func TestEqual(t *testing.T) {
	assert.True(t, Equal([]string{"one", "two", "three"}, []string{"one", "two", "three"}))
	assert.True(t, Equal([]int{1, 2, 3}, []int{1, 2, 3}))
	assert.True(t, Equal([]float64{1.1, 2.2, 3.3}, []float64{1.1, 2.2, 3.3}))

	assert.False(t, Equal([]string{"one", "two", "three"}, []string{"one", "three", "two"}))
	assert.False(t, Equal([]string{"one", "two", "three"}, []string{"one"}))
	assert.False(t, Equal([]string{"one", "two", "three"}, []int{1, 2, 3}))
}
