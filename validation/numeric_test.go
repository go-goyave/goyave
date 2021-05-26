package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateNumeric(t *testing.T) {
	assert.True(t, validateNumeric("field", 1, []string{}, map[string]interface{}{"field": 1}))
	assert.True(t, validateNumeric("field", 1.2, []string{}, map[string]interface{}{"field": 1.2}))
	assert.True(t, validateNumeric("field", uint(1), []string{}, map[string]interface{}{"field": uint(1)}))
	assert.True(t, validateNumeric("field", uint8(1), []string{}, map[string]interface{}{"field": uint8(1)}))
	assert.True(t, validateNumeric("field", uint16(1), []string{}, map[string]interface{}{"field": uint16(1)}))
	assert.True(t, validateNumeric("field", float32(1.3), []string{}, map[string]interface{}{"field": float32(1.3)}))
	assert.True(t, validateNumeric("field", "2", []string{}, map[string]interface{}{"field": "2"}))
	assert.True(t, validateNumeric("field", "1.2", []string{}, map[string]interface{}{"field": "1.2"}))
	assert.True(t, validateNumeric("field", "-1", []string{}, map[string]interface{}{"field": "-1"}))
	assert.True(t, validateNumeric("field", "-1.3", []string{}, map[string]interface{}{"field": "1.3"}))
	assert.False(t, validateNumeric("field", uintptr(1), []string{}, map[string]interface{}{"field": uintptr(1)}))
	assert.False(t, validateNumeric("field", []string{}, []string{}, map[string]interface{}{"field": []string{}}))
	assert.False(t, validateNumeric("field", map[string]string{}, []string{}, map[string]interface{}{"field": map[string]string{}}))
	assert.False(t, validateNumeric("field", "test", []string{}, map[string]interface{}{"field": "test"}))
}

func TestValidateNumericConvertString(t *testing.T) {
	form1 := map[string]interface{}{"field": "1.2"}
	validateNumeric("field", form1["field"], []string{}, form1)
	assert.Equal(t, 1.2, form1["field"])

	form2 := map[string]interface{}{"field": "-1.3"}
	validateNumeric("field", form2["field"], []string{}, form2)
	assert.Equal(t, -1.3, form2["field"])

	form3 := map[string]interface{}{"field": "2"}
	validateNumeric("field", form3["field"], []string{}, form3)
	assert.Equal(t, float64(2), form3["field"])
}

func TestValidateNumericConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"numeric": "1",
		},
	}

	set := RuleSet{
		"object":         {"required", "object"},
		"object.numeric": {"required", "numeric"},
	}

	errors := Validate(data, set, true, "en-US")
	assert.Empty(t, errors)
	_, ok := data["object"].(map[string]interface{})["numeric"].(float64)
	assert.True(t, ok)
}

func TestValidateInteger(t *testing.T) {
	assert.True(t, validateInteger("field", 1, []string{}, map[string]interface{}{"field": 1}))
	assert.True(t, validateInteger("field", float64(2), []string{}, map[string]interface{}{"field": float64(2)}))
	assert.True(t, validateInteger("field", float32(3), []string{}, map[string]interface{}{"field": float32(3)}))
	assert.True(t, validateInteger("field", uint(1), []string{}, map[string]interface{}{"field": uint(1)}))
	assert.True(t, validateInteger("field", uint8(1), []string{}, map[string]interface{}{"field": uint8(1)}))
	assert.True(t, validateInteger("field", uint16(1), []string{}, map[string]interface{}{"field": uint16(1)}))
	assert.True(t, validateInteger("field", "2", []string{}, map[string]interface{}{"field": "2"}))
	assert.True(t, validateInteger("field", "-1", []string{}, map[string]interface{}{"field": "-1"}))
	assert.False(t, validateInteger("field", 2.2, []string{}, map[string]interface{}{"field": 2.2}))
	assert.False(t, validateInteger("field", float32(3.2), []string{}, map[string]interface{}{"field": float32(3.2)}))
	assert.False(t, validateInteger("field", "1.2", []string{}, map[string]interface{}{"field": "1.2"}))
	assert.False(t, validateInteger("field", "-1.3", []string{}, map[string]interface{}{"field": "1.3"}))
	assert.False(t, validateInteger("field", uintptr(1), []string{}, map[string]interface{}{"field": uintptr(1)}))
	assert.False(t, validateInteger("field", []string{}, []string{}, map[string]interface{}{"field": []string{}}))
	assert.False(t, validateInteger("field", map[string]string{}, []string{}, map[string]interface{}{"field": map[string]string{}}))
	assert.False(t, validateInteger("field", "test", []string{}, map[string]interface{}{"field": "test"}))
}

func TestValidateIntegerConvert(t *testing.T) {
	form1 := map[string]interface{}{"field": "1"}
	validateInteger("field", form1["field"], []string{}, form1)
	assert.Equal(t, 1, form1["field"])

	form2 := map[string]interface{}{"field": "-2"}
	validateInteger("field", form2["field"], []string{}, form2)
	assert.Equal(t, -2, form2["field"])

	form3 := map[string]interface{}{"field": float64(3)}
	validateInteger("field", form3["field"], []string{}, form3)
	assert.Equal(t, 3, form3["field"])
}

func TestValidateIntegerConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"integer": "1",
		},
	}

	set := RuleSet{
		"object":         {"required", "object"},
		"object.integer": {"required", "integer"},
	}

	errors := Validate(data, set, true, "en-US")
	assert.Empty(t, errors)
	_, ok := data["object"].(map[string]interface{})["integer"].(int)
	assert.True(t, ok)
}
