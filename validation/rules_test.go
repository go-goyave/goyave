package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestValidateBool(t *testing.T) {
	assert.True(t, validateBool("field", 1, []string{}, map[string]interface{}{}))
	assert.True(t, validateBool("field", 0, []string{}, map[string]interface{}{}))
	assert.True(t, validateBool("field", "on", []string{}, map[string]interface{}{}))
	assert.True(t, validateBool("field", "off", []string{}, map[string]interface{}{}))
	assert.True(t, validateBool("field", "true", []string{}, map[string]interface{}{}))
	assert.True(t, validateBool("field", "false", []string{}, map[string]interface{}{}))
	assert.True(t, validateBool("field", "yes", []string{}, map[string]interface{}{}))
	assert.True(t, validateBool("field", "no", []string{}, map[string]interface{}{}))
	assert.True(t, validateBool("field", true, []string{}, map[string]interface{}{}))
	assert.True(t, validateBool("field", false, []string{}, map[string]interface{}{}))

	assert.False(t, validateBool("field", 0.0, []string{}, map[string]interface{}{}))
	assert.False(t, validateBool("field", 1.0, []string{}, map[string]interface{}{}))
	assert.False(t, validateBool("field", []string{"true"}, []string{}, map[string]interface{}{}))
	assert.False(t, validateBool("field", -1, []string{}, map[string]interface{}{}))
}

func TestValidateBoolConvert(t *testing.T) {
	form := map[string]interface{}{"field": "on"}
	assert.True(t, validateBool("field", form["field"], []string{}, form))
	b, ok := form["field"].(bool)
	assert.True(t, ok)
	assert.True(t, b)

	form = map[string]interface{}{"field": "off"}
	assert.True(t, validateBool("field", form["field"], []string{}, form))
	b, ok = form["field"].(bool)
	assert.True(t, ok)
	assert.False(t, b)

	form = map[string]interface{}{"field": 1}
	assert.True(t, validateBool("field", form["field"], []string{}, form))
	b, ok = form["field"].(bool)
	assert.True(t, ok)
	assert.True(t, b)

	form = map[string]interface{}{"field": 0}
	assert.True(t, validateBool("field", form["field"], []string{}, form))
	b, ok = form["field"].(bool)
	assert.True(t, ok)
	assert.False(t, b)
}
