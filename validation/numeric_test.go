package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateNumeric(t *testing.T) {
	assert.True(t, validateNumeric(newTestContext("field", 1, []string{}, map[string]interface{}{"field": 1})))
	assert.True(t, validateNumeric(newTestContext("field", 1.2, []string{}, map[string]interface{}{"field": 1.2})))
	assert.True(t, validateNumeric(newTestContext("field", uint(1), []string{}, map[string]interface{}{"field": uint(1)})))
	assert.True(t, validateNumeric(newTestContext("field", uint8(1), []string{}, map[string]interface{}{"field": uint8(1)})))
	assert.True(t, validateNumeric(newTestContext("field", uint16(1), []string{}, map[string]interface{}{"field": uint16(1)})))
	assert.True(t, validateNumeric(newTestContext("field", float32(1.3), []string{}, map[string]interface{}{"field": float32(1.3)})))
	assert.True(t, validateNumeric(newTestContext("field", "2", []string{}, map[string]interface{}{"field": "2"})))
	assert.True(t, validateNumeric(newTestContext("field", "1.2", []string{}, map[string]interface{}{"field": "1.2"})))
	assert.True(t, validateNumeric(newTestContext("field", "-1", []string{}, map[string]interface{}{"field": "-1"})))
	assert.True(t, validateNumeric(newTestContext("field", "-1.3", []string{}, map[string]interface{}{"field": "1.3"})))
	assert.False(t, validateNumeric(newTestContext("field", uintptr(1), []string{}, map[string]interface{}{"field": uintptr(1)})))
	assert.False(t, validateNumeric(newTestContext("field", []string{}, []string{}, map[string]interface{}{"field": []string{}})))
	assert.False(t, validateNumeric(newTestContext("field", map[string]string{}, []string{}, map[string]interface{}{"field": map[string]string{}})))
	assert.False(t, validateNumeric(newTestContext("field", "test", []string{}, map[string]interface{}{"field": "test"})))
}

func TestValidateNumericConvertString(t *testing.T) {
	form1 := map[string]interface{}{"field": "1.2"}
	ctx1 := newTestContext("field", form1["field"], []string{}, form1)
	validateNumeric(ctx1)
	assert.Equal(t, 1.2, ctx1.Value)

	form2 := map[string]interface{}{"field": "-1.3"}
	ctx2 := newTestContext("field", form2["field"], []string{}, form2)
	validateNumeric(ctx2)
	assert.Equal(t, -1.3, ctx2.Value)

	form3 := map[string]interface{}{"field": "2"}
	ctx3 := newTestContext("field", form3["field"], []string{}, form3)
	validateNumeric(ctx3)
	assert.Equal(t, float64(2), ctx3.Value)
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
	assert.True(t, validateInteger(newTestContext("field", 1, []string{}, map[string]interface{}{"field": 1})))
	assert.True(t, validateInteger(newTestContext("field", float64(2), []string{}, map[string]interface{}{"field": float64(2)})))
	assert.True(t, validateInteger(newTestContext("field", float32(3), []string{}, map[string]interface{}{"field": float32(3)})))
	assert.True(t, validateInteger(newTestContext("field", uint(1), []string{}, map[string]interface{}{"field": uint(1)})))
	assert.True(t, validateInteger(newTestContext("field", uint8(1), []string{}, map[string]interface{}{"field": uint8(1)})))
	assert.True(t, validateInteger(newTestContext("field", uint16(1), []string{}, map[string]interface{}{"field": uint16(1)})))
	assert.True(t, validateInteger(newTestContext("field", "2", []string{}, map[string]interface{}{"field": "2"})))
	assert.True(t, validateInteger(newTestContext("field", "-1", []string{}, map[string]interface{}{"field": "-1"})))
	assert.False(t, validateInteger(newTestContext("field", 2.2, []string{}, map[string]interface{}{"field": 2.2})))
	assert.False(t, validateInteger(newTestContext("field", float32(3.2), []string{}, map[string]interface{}{"field": float32(3.2)})))
	assert.False(t, validateInteger(newTestContext("field", "1.2", []string{}, map[string]interface{}{"field": "1.2"})))
	assert.False(t, validateInteger(newTestContext("field", "-1.3", []string{}, map[string]interface{}{"field": "1.3"})))
	assert.False(t, validateInteger(newTestContext("field", uintptr(1), []string{}, map[string]interface{}{"field": uintptr(1)})))
	assert.False(t, validateInteger(newTestContext("field", []string{}, []string{}, map[string]interface{}{"field": []string{}})))
	assert.False(t, validateInteger(newTestContext("field", map[string]string{}, []string{}, map[string]interface{}{"field": map[string]string{}})))
	assert.False(t, validateInteger(newTestContext("field", "test", []string{}, map[string]interface{}{"field": "test"})))
}

func TestValidateIntegerConvert(t *testing.T) {
	form1 := map[string]interface{}{"field": "1"}
	ctx1 := newTestContext("field", form1["field"], []string{}, form1)
	validateInteger(ctx1)
	assert.Equal(t, 1, ctx1.Value)

	form2 := map[string]interface{}{"field": "-2"}
	ctx2 := newTestContext("field", form2["field"], []string{}, form2)
	validateInteger(ctx2)
	assert.Equal(t, -2, ctx2.Value)

	form3 := map[string]interface{}{"field": float64(3)}
	ctx3 := newTestContext("field", form3["field"], []string{}, form3)
	validateInteger(ctx3)
	assert.Equal(t, 3, ctx3.Value)
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
