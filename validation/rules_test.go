package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateString(t *testing.T) {
	assert.True(t, validateString("field", "string", []string{}, map[string]interface{}{}))
	assert.False(t, validateString("field", 2, []string{}, map[string]interface{}{}))
	assert.False(t, validateString("field", 2.5, []string{}, map[string]interface{}{}))
	assert.False(t, validateString("field", []byte{}, []string{}, map[string]interface{}{}))
	assert.False(t, validateString("field", []string{}, []string{}, map[string]interface{}{}))
}

func TestValidateRequired(t *testing.T) {
	assert.True(t, validateRequired("field", "not empty", []string{}, map[string]interface{}{"field": "not empty"}))
	assert.True(t, validateRequired("field", 1, []string{}, map[string]interface{}{"field": 1}))
	assert.True(t, validateRequired("field", 2.5, []string{}, map[string]interface{}{"field": 2.5}))
	assert.True(t, validateRequired("field", []string{}, []string{}, map[string]interface{}{"field": []string{}}))
	assert.True(t, validateRequired("field", []float64{}, []string{}, map[string]interface{}{"field": []float64{}}))
	assert.True(t, validateRequired("field", 0, []string{}, map[string]interface{}{"field": 0}))
	assert.False(t, validateRequired("field", nil, []string{}, map[string]interface{}{"field": nil}))
	assert.False(t, validateRequired("field", "", []string{}, map[string]interface{}{"field": ""}))
}
