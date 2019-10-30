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

func TestValidateMin(t *testing.T) {
	assert.True(t, validateMin("field", "not numeric", []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateMin("field", "not numeric", []string{"20"}, map[string]interface{}{}))

	assert.True(t, validateMin("field", 2, []string{"1"}, map[string]interface{}{}))
	assert.False(t, validateMin("field", 10, []string{"20"}, map[string]interface{}{}))

	assert.True(t, validateMin("field", 2.0, []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateMin("field", 10.0, []string{"20"}, map[string]interface{}{}))
	assert.True(t, validateMin("field", 3.7, []string{"2.5"}, map[string]interface{}{}))
	assert.False(t, validateMin("field", 10.0, []string{"20.4"}, map[string]interface{}{}))

	assert.True(t, validateMin("field", []int{5, 4}, []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateMin("field", []int{5, 4, 3, 2}, []string{"20"}, map[string]interface{}{}))

	assert.True(t, validateMin("field", []string{"5", "4"}, []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateMin("field", []string{"5", "4", "3", "2"}, []string{"20"}, map[string]interface{}{}))

	assert.True(t, validateMin("field", true, []string{"2"}, map[string]interface{}{}))
	assert.Panics(t, func() { validateMin("field", true, []string{"test"}, map[string]interface{}{}) })
}

func TestValidateMax(t *testing.T) {
	assert.True(t, validateMax("field", "not numeric", []string{"12"}, map[string]interface{}{}))
	assert.False(t, validateMax("field", "not numeric", []string{"5"}, map[string]interface{}{}))

	assert.True(t, validateMax("field", 1, []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateMax("field", 20, []string{"10"}, map[string]interface{}{}))

	assert.True(t, validateMax("field", 2.0, []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateMax("field", 10.0, []string{"5"}, map[string]interface{}{}))
	assert.True(t, validateMax("field", 2.5, []string{"3.7"}, map[string]interface{}{}))
	assert.False(t, validateMax("field", 20.4, []string{"10.0"}, map[string]interface{}{}))

	assert.True(t, validateMax("field", []int{5, 4}, []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateMax("field", []int{5, 4, 3, 2}, []string{"3"}, map[string]interface{}{}))

	assert.True(t, validateMax("field", []string{"5", "4"}, []string{"3"}, map[string]interface{}{}))
	assert.False(t, validateMax("field", []string{"5", "4", "3", "2"}, []string{"2"}, map[string]interface{}{}))

	assert.True(t, validateMax("field", true, []string{"2"}, map[string]interface{}{}))
	assert.Panics(t, func() { validateMax("field", true, []string{"test"}, map[string]interface{}{}) })
}

func TestValidateBetween(t *testing.T) {
	assert.True(t, validateBetween("field", "not numeric", []string{"5", "12"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", "not numeric", []string{"12", "20"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", "not numeric", []string{"5", "6"}, map[string]interface{}{}))

	assert.True(t, validateBetween("field", 1, []string{"0", "3"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", 20, []string{"5", "10"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", 20, []string{"21", "23"}, map[string]interface{}{}))

	assert.True(t, validateBetween("field", 2.0, []string{"2", "5"}, map[string]interface{}{}))
	assert.True(t, validateBetween("field", 2.0, []string{"1.0", "5.0"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", 10.0, []string{"5", "7"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", 10.0, []string{"15", "17"}, map[string]interface{}{}))
	assert.True(t, validateBetween("field", 2.5, []string{"1.7", "3.7"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", 20.4, []string{"10.0", "14.7"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", 20.4, []string{"25.0", "54.7"}, map[string]interface{}{}))

	assert.True(t, validateBetween("field", []int{5, 4}, []string{"1", "5"}, map[string]interface{}{}))
	assert.True(t, validateBetween("field", []int{5, 4}, []string{"2.2", "5.7"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", []int{5, 4, 3, 2}, []string{"1", "3"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", []int{5, 4, 3, 2}, []string{"5", "7"}, map[string]interface{}{}))

	assert.True(t, validateBetween("field", []string{"5", "4"}, []string{"1", "5"}, map[string]interface{}{}))
	assert.True(t, validateBetween("field", []string{"5", "4"}, []string{"2.2", "5.7"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", []string{"5", "4", "3", "2"}, []string{"1", "3"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", []string{"5", "4", "3", "2"}, []string{"5", "7"}, map[string]interface{}{}))

	assert.True(t, validateBetween("field", true, []string{"2", "3"}, map[string]interface{}{}))
	assert.Panics(t, func() { validateBetween("field", true, []string{"test"}, map[string]interface{}{}) })
	assert.Panics(t, func() { validateBetween("field", true, []string{"1"}, map[string]interface{}{}) })
	assert.Panics(t, func() { validateBetween("field", true, []string{"test", "2"}, map[string]interface{}{}) })
	assert.Panics(t, func() { validateBetween("field", true, []string{"2", "test"}, map[string]interface{}{}) })
}
