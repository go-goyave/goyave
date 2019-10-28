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
