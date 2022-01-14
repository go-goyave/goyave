package typeutil

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	v, err = ToFloat64("NaN")
	assert.Nil(t, err)
	assert.True(t, math.IsNaN(v))

	v, err = ToFloat64([]string{})
	assert.NotNil(t, err)
	assert.Equal(t, float64(0), v)
}

func TestToString(t *testing.T) {
	assert.Equal(t, "12", ToString(12))
	assert.Equal(t, "-12", ToString(-12))
	assert.Equal(t, "12.5", ToString(12.5))
	assert.Equal(t, "-12.5", ToString(-12.5))
	assert.Equal(t, "true", ToString(true))
	assert.Equal(t, "[test]", ToString([]string{"test"}))
}
