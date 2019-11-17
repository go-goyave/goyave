package helpers

import (
	"fmt"
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

func TestParseMultiValuesHeader(t *testing.T) {
	expected := []HeaderValue{
		{Value: "text/html", Priority: 0.8},
		{Value: "text/*", Priority: 0.8},
		{Value: "*/*", Priority: 0.8},
	}
	result := ParseMultiValuesHeader("text/html;q=0.8,text/*;q=0.8,*/*;q=0.8")
	assert.True(t, SliceEqual(expected, result))

	result = ParseMultiValuesHeader("*/*;q=0.8,text/*;q=0.8,text/html;q=0.8")
	assert.True(t, SliceEqual(expected, result))

	expected = []HeaderValue{
		{Value: "text/html", Priority: 1},
		{Value: "*/*", Priority: 0.7},
		{Value: "text/*", Priority: 0.5},
	}
	result = ParseMultiValuesHeader("text/html,text/*;q=0.5,*/*;q=0.7")
	assert.True(t, SliceEqual(expected, result))

	expected = []HeaderValue{
		{Value: "fr", Priority: 1},
		{Value: "fr-FR", Priority: 0.8},
		{Value: "en-US", Priority: 0.5},
		{Value: "en-*", Priority: 0.3},
		{Value: "en", Priority: 0.3},
		{Value: "*", Priority: 0.3},
	}
	result = ParseMultiValuesHeader("fr , fr-FR;q=0.8, en-US ;q=0.5, *;q=0.3, en-*;q=0.3, en;q=0.3")
	fmt.Println(result)
	assert.True(t, SliceEqual(expected, result))
}
