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

	// TODO test files
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

	// TODO test files
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

	// TODO test file
}

func TestValidateGreaterThan(t *testing.T) {
	assert.True(t, validateGreaterThan("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 2}))
	assert.False(t, validateGreaterThan("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 20}))

	assert.True(t, validateGreaterThan("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 2.0}))
	assert.False(t, validateGreaterThan("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 5.1}))

	assert.True(t, validateGreaterThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "str"}))
	assert.False(t, validateGreaterThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "other string"}))

	assert.True(t, validateGreaterThan("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1}}))
	assert.False(t, validateGreaterThan("field", []int{6}, []string{"comparison"}, map[string]interface{}{"field": []int{6}, "comparison": []int{1, 2, 3}}))

	// Different type
	assert.False(t, validateGreaterThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": 2}))

	// Missing field
	assert.False(t, validateGreaterThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string"}))

	test := "string"
	assert.False(t, validateGreaterThan("field", &test, []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": &test}))
}

func TestValidateGreaterThanEqual(t *testing.T) {
	assert.True(t, validateGreaterThanEqual("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 2}))
	assert.True(t, validateGreaterThanEqual("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 5}))
	assert.False(t, validateGreaterThanEqual("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 20}))
	assert.False(t, validateGreaterThanEqual("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 5.1}))

	assert.True(t, validateGreaterThanEqual("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 2.0}))
	assert.True(t, validateGreaterThanEqual("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 5.0}))
	assert.False(t, validateGreaterThanEqual("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 5.1}))

	assert.True(t, validateGreaterThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "str"}))
	assert.True(t, validateGreaterThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "gnirts"}))
	assert.False(t, validateGreaterThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "other string"}))

	assert.True(t, validateGreaterThanEqual("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1}}))
	assert.True(t, validateGreaterThanEqual("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1, 2}}))
	assert.False(t, validateGreaterThanEqual("field", []int{6}, []string{"comparison"}, map[string]interface{}{"field": []int{6}, "comparison": []int{1, 2, 3}}))

	// Different type
	assert.False(t, validateGreaterThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": 2}))

	// Missing field
	assert.False(t, validateGreaterThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string"}))

	test := "string"
	assert.False(t, validateGreaterThanEqual("field", &test, []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": &test}))
}

func TestValidateLowerThan(t *testing.T) {
	assert.True(t, validateLowerThan("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 7}))
	assert.False(t, validateLowerThan("field", 20, []string{"comparison"}, map[string]interface{}{"field": 20, "comparison": 5}))

	assert.True(t, validateLowerThan("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 7.0}))
	assert.False(t, validateLowerThan("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 4.9}))

	assert.True(t, validateLowerThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "other string"}))
	assert.False(t, validateLowerThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "str"}))

	assert.True(t, validateLowerThan("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1, 2, 3}}))
	assert.False(t, validateLowerThan("field", []int{6, 7, 8}, []string{"comparison"}, map[string]interface{}{"field": []int{6, 7, 8}, "comparison": []int{1, 2}}))

	// Different type
	assert.False(t, validateLowerThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": 2}))

	// Missing field
	assert.False(t, validateLowerThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string"}))

	test := "string"
	assert.False(t, validateLowerThan("field", &test, []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": &test}))
}

func TestValidateLowerThanEqual(t *testing.T) {
	assert.True(t, validateLowerThanEqual("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 7}))
	assert.True(t, validateLowerThanEqual("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 5}))
	assert.False(t, validateLowerThanEqual("field", 20, []string{"comparison"}, map[string]interface{}{"field": 20, "comparison": 5}))
	assert.False(t, validateLowerThanEqual("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 4.9}))

	assert.True(t, validateLowerThanEqual("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 7.0}))
	assert.True(t, validateLowerThanEqual("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 5.0}))
	assert.False(t, validateLowerThanEqual("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 4.9}))

	assert.True(t, validateLowerThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "other string"}))
	assert.True(t, validateLowerThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "gnirts"}))
	assert.False(t, validateLowerThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "str"}))

	assert.True(t, validateLowerThanEqual("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1, 2, 3}}))
	assert.True(t, validateLowerThanEqual("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1, 2}}))
	assert.False(t, validateLowerThanEqual("field", []int{6, 7, 8}, []string{"comparison"}, map[string]interface{}{"field": []int{6, 7, 8}, "comparison": []int{1, 2}}))

	// Different type
	assert.False(t, validateLowerThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": 2}))

	// Missing field
	assert.False(t, validateLowerThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string"}))

	test := "string"
	assert.False(t, validateLowerThanEqual("field", &test, []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": &test}))
}

func TestValidateArray(t *testing.T) {
	assert.True(t, validateArray("field", []string{"test"}, []string{}, map[string]interface{}{}))
	assert.True(t, validateArray("field", []int{5}, []string{}, map[string]interface{}{}))
	assert.True(t, validateArray("field", []float64{5.5}, []string{}, map[string]interface{}{}))
	assert.True(t, validateArray("field", []bool{true}, []string{}, map[string]interface{}{}))
	assert.False(t, validateArray("field", map[string]string{}, []string{}, map[string]interface{}{}))
	assert.False(t, validateArray("field", "test", []string{}, map[string]interface{}{}))
	assert.False(t, validateArray("field", 5, []string{}, map[string]interface{}{}))
	assert.False(t, validateArray("field", 5.0, []string{}, map[string]interface{}{}))
	assert.False(t, validateArray("field", true, []string{}, map[string]interface{}{}))
}

func TestValidateDistinct(t *testing.T) {
	assert.True(t, validateDistinct("field", []string{"test", "test2", "test3"}, []string{}, map[string]interface{}{}))
	assert.True(t, validateDistinct("field", []int{1, 2, 3}, []string{}, map[string]interface{}{}))
	assert.True(t, validateDistinct("field", []float64{1.2, 4.3, 2.4, 3.5, 4.5, 4.30001}, []string{}, map[string]interface{}{}))
	assert.True(t, validateDistinct("field", []bool{true, false}, []string{}, map[string]interface{}{}))

	assert.False(t, validateDistinct("field", []string{"test", "test2", "test3", "test2"}, []string{}, map[string]interface{}{}))
	assert.False(t, validateDistinct("field", []int{1, 4, 2, 3, 4}, []string{}, map[string]interface{}{}))
	assert.False(t, validateDistinct("field", []float64{1.2, 4.3, 2.4, 3.5, 4.5, 4.30001, 4.3}, []string{}, map[string]interface{}{}))

	// Not array
	assert.False(t, validateDistinct("field", 8, []string{}, map[string]interface{}{}))
	assert.False(t, validateDistinct("field", 8.0, []string{}, map[string]interface{}{}))
	assert.False(t, validateDistinct("field", "string", []string{}, map[string]interface{}{}))
}

func TestValidateDigits(t *testing.T) {
	assert.True(t, validateDigits("field", "123", []string{}, map[string]interface{}{}))
	assert.True(t, validateDigits("field", "0123456789", []string{}, map[string]interface{}{}))
	assert.False(t, validateDigits("field", "2.3", []string{}, map[string]interface{}{}))
	assert.False(t, validateDigits("field", "-123", []string{}, map[string]interface{}{}))
	assert.False(t, validateDigits("field", "abcd", []string{}, map[string]interface{}{}))
	assert.False(t, validateDigits("field", "/*-&é\"'(-è_ç", []string{}, map[string]interface{}{}))

	// Not string
	assert.False(t, validateDigits("field", 1, []string{}, map[string]interface{}{}))
	assert.False(t, validateDigits("field", 1.2, []string{}, map[string]interface{}{}))
	assert.False(t, validateDigits("field", true, []string{}, map[string]interface{}{}))
}

func TestValidateLength(t *testing.T) {
	assert.True(t, validateLength("field", "123", []string{"3"}, map[string]interface{}{}))
	assert.True(t, validateLength("field", "", []string{"0"}, map[string]interface{}{}))
	assert.False(t, validateLength("field", "4567", []string{"5"}, map[string]interface{}{}))
	assert.False(t, validateLength("field", "4567", []string{"2"}, map[string]interface{}{}))

	assert.False(t, validateLength("field", 4567, []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateLength("field", 4567.8, []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateLength("field", true, []string{"2"}, map[string]interface{}{}))

	assert.Panics(t, func() { validateLength("field", "123", []string{"test"}, map[string]interface{}{}) })
}

func TestValidateRegex(t *testing.T) {
	assert.True(t, validateRegex("field", "sghtyhg", []string{"t"}, map[string]interface{}{}))
	assert.True(t, validateRegex("field", "sghtyhg", []string{"[^\\s]"}, map[string]interface{}{}))
	assert.False(t, validateRegex("field", "sgh tyhg", []string{"^[^\\s]+$"}, map[string]interface{}{}))
	assert.False(t, validateRegex("field", "48s9", []string{"^[^0-9]+$"}, map[string]interface{}{}))
	assert.True(t, validateRegex("field", "489", []string{"^[0-9]+$"}, map[string]interface{}{}))
	assert.False(t, validateRegex("field", 489, []string{"^[^0-9]+$"}, map[string]interface{}{}))

	assert.Panics(t, func() { validateRegex("field", "", []string{"doesn't compile \\"}, map[string]interface{}{}) })
}

func TestValidateEmail(t *testing.T) {
	assert.True(t, validateEmail("field", "simple@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "very.common@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "disposable.style.email.with+symbol@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "other.email-with-hyphen@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "fully-qualified-domain@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "user.name+tag+sorting@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "x@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "example-indeed@strange-example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "admin@mailserver1", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "example@s.example", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "\" \"@example.org", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "\"john..doe\"@example.org", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "mailhost!username@example.org", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "user%example.com@example.org", []string{}, map[string]interface{}{}))
	assert.False(t, validateEmail("field", "Abc.example.com", []string{}, map[string]interface{}{}))
	assert.False(t, validateEmail("field", "1234567890123456789012345678901234567890123456789012345678901234+x@example.com", []string{}, map[string]interface{}{}))
}

func TestValidateAlpha(t *testing.T) {
	assert.True(t, validateAlpha("field", "helloworld", []string{}, map[string]interface{}{}))
	assert.True(t, validateAlpha("field", "éèçàû", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlpha("field", "hello world", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlpha("field", "/+*(@)={}\"'", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlpha("field", "helloworld2", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlpha("field", 2, []string{}, map[string]interface{}{}))
}

func TestValidateAlphaDash(t *testing.T) {
	assert.True(t, validateAlphaDash("field", "helloworld", []string{}, map[string]interface{}{}))
	assert.True(t, validateAlphaDash("field", "éèçàû_-", []string{}, map[string]interface{}{}))
	assert.True(t, validateAlphaDash("field", "hello-world", []string{}, map[string]interface{}{}))
	assert.True(t, validateAlphaDash("field", "hello-world_2", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlphaDash("field", "hello world", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlphaDash("field", "/+*(@)={}\"'", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlphaDash("field", 2, []string{}, map[string]interface{}{}))
}

func TestValidateAlphaNumeric(t *testing.T) {
	assert.True(t, validateAlphaNumeric("field", "helloworld2", []string{}, map[string]interface{}{}))
	assert.True(t, validateAlphaNumeric("field", "éèçàû2", []string{}, map[string]interface{}{}))
	assert.True(t, validateAlphaNumeric("field", "helloworld2", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlphaNumeric("field", "hello world", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlphaNumeric("field", "/+*(@)={}\"'", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlphaNumeric("field", 2, []string{}, map[string]interface{}{}))
}

func TestValidateStartsWith(t *testing.T) {
	assert.True(t, validateStartsWith("field", "hello world", []string{"hello"}, map[string]interface{}{}))
	assert.True(t, validateStartsWith("field", "hi", []string{"hello", "hi", "hey"}, map[string]interface{}{}))
	assert.False(t, validateStartsWith("field", "sup'!", []string{"hello", "hi", "hey"}, map[string]interface{}{}))
	assert.Panics(t, func() { validateStartsWith("field", "sup'!", []string{}, map[string]interface{}{}) })
}

func TestValidateEndsWith(t *testing.T) {
	assert.True(t, validateEndsWith("field", "hello world", []string{"world"}, map[string]interface{}{}))
	assert.True(t, validateEndsWith("field", "oh hi mark", []string{"ross", "mark", "bruce"}, map[string]interface{}{}))
	assert.False(t, validateEndsWith("field", "sup' bro!", []string{"ross", "mark", "bruce"}, map[string]interface{}{}))
	assert.Panics(t, func() { validateEndsWith("field", "sup'!", []string{}, map[string]interface{}{}) })
}

func TestValidateIn(t *testing.T) {
	assert.True(t, validateIn("field", "dolor", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{}))
	assert.False(t, validateIn("field", "dolors", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{}))
	assert.False(t, validateIn("field", "hello world", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{}))

	assert.True(t, validateIn("field", 2.5, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{}))
	assert.False(t, validateIn("field", 2.51, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{}))

	assert.False(t, validateIn("field", []string{"1"}, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{}))
	assert.Panics(t, func() { validateIn("field", "hi", []string{}, map[string]interface{}{}) })
}

func TestValidateNotIn(t *testing.T) {
	assert.False(t, validateNotIn("field", "dolor", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{}))
	assert.True(t, validateNotIn("field", "dolors", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{}))
	assert.True(t, validateNotIn("field", "hello world", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{}))

	assert.False(t, validateNotIn("field", 2.5, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{}))
	assert.True(t, validateNotIn("field", 2.51, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{}))

	assert.False(t, validateNotIn("field", []string{"1"}, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{}))
	assert.Panics(t, func() { validateNotIn("field", "hi", []string{}, map[string]interface{}{}) })
}
