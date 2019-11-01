package helpers

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

func TestToFloat64(t *testing.T) {
	v, err := ToFloat64(1)
	assert.Nil(t, err)
	assert.Equal(t, 1.0, v)

	v, err = ToFloat64("1")
	assert.Nil(t, err)
	assert.Equal(t, 1.0, v)

	v, err = ToFloat64(uint(5))
	assert.Nil(t, err)
	assert.Equal(t, 5.0, v)

	v, err = ToFloat64("1.58")
	assert.Nil(t, err)
	assert.Equal(t, 1.58, v)

	v, err = ToFloat64(1.5)
	assert.Nil(t, err)
	assert.Equal(t, 1.5, v)
}

func TestToString(t *testing.T) {
	assert.Equal(t, "12", ToString(12))
	assert.Equal(t, "-12", ToString(-12))
	assert.Equal(t, "12.5", ToString(12.5))
	assert.Equal(t, "-12.5", ToString(-12.5))
	assert.Equal(t, "true", ToString(true))
	assert.Equal(t, "[test]", ToString([]string{"test"}))
}

func TestSliceEqual(t *testing.T) {
	assert.True(t, SliceEqual([]string{"one", "two", "three"}, []string{"one", "two", "three"}))
	assert.True(t, SliceEqual([]int{1, 2, 3}, []int{1, 2, 3}))
	assert.True(t, SliceEqual([]float64{1.1, 2.2, 3.3}, []float64{1.1, 2.2, 3.3}))

	assert.False(t, SliceEqual([]string{"one", "two", "three"}, []string{"one", "three", "two"}))
	assert.False(t, SliceEqual([]string{"one", "two", "three"}, []string{"one"}))
	assert.False(t, SliceEqual([]string{"one", "two", "three"}, []int{1, 2, 3}))
}
