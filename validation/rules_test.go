package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestContext(field string, value interface{}, parameters []string, form map[string]interface{}) *Context {
	return &Context{
		Data:   form,
		Value:  value,
		Parent: form,
		Name:   field,
		Rule: &Rule{
			Params: parameters,
		},
	}
}

func newTestContextWithField(fieldName string, value interface{}, parameters []string, form map[string]interface{}, field *Field) *Context {
	ctx := newTestContext(fieldName, value, parameters, form)
	ctx.Field = field
	return ctx
}

func TestValidateRequired(t *testing.T) {
	assert.True(t, validateRequired(newTestContextWithField("field", "not empty", []string{}, map[string]interface{}{"field": "not empty"}, &Field{})))
	assert.True(t, validateRequired(newTestContextWithField("field", 1, []string{}, map[string]interface{}{"field": 1}, &Field{})))
	assert.True(t, validateRequired(newTestContextWithField("field", 2.5, []string{}, map[string]interface{}{"field": 2.5}, &Field{})))
	assert.True(t, validateRequired(newTestContextWithField("field", []string{}, []string{}, map[string]interface{}{"field": []string{}}, &Field{})))
	assert.True(t, validateRequired(newTestContextWithField("field", []float64{}, []string{}, map[string]interface{}{"field": []float64{}}, &Field{})))
	assert.True(t, validateRequired(newTestContextWithField("field", 0, []string{}, map[string]interface{}{"field": 0}, &Field{})))
	assert.True(t, validateRequired(newTestContextWithField("field", nil, []string{}, map[string]interface{}{"field": nil}, &Field{isNullable: true})))
	assert.True(t, validateRequired(newTestContextWithField("field", "", []string{}, map[string]interface{}{"field": ""}, &Field{})))
	assert.False(t, validateRequired(newTestContextWithField("field", nil, []string{}, map[string]interface{}{"field": nil}, &Field{})))

	data := map[string]interface{}{
		"object": map[string]interface{}{
			"key": "value",
		},
	}
	assert.True(t, validateRequired(newTestContextWithField("object.key", "value", []string{}, data, &Field{})))
	assert.False(t, validateRequired(newTestContextWithField("object.notakey", nil, []string{}, data, &Field{})))
}

func TestValidateMin(t *testing.T) {
	assert.True(t, validateMin(newTestContext("field", "not numeric", []string{"2"}, map[string]interface{}{})))
	assert.False(t, validateMin(newTestContext("field", "not numeric", []string{"20"}, map[string]interface{}{})))
	assert.False(t, validateMin(newTestContext("field", "🇩🇪", []string{"2"}, map[string]interface{}{})))
	assert.False(t, validateMin(newTestContext("field", "👍🏼", []string{"2"}, map[string]interface{}{})))

	assert.True(t, validateMin(newTestContext("field", 2, []string{"1"}, map[string]interface{}{})))
	assert.False(t, validateMin(newTestContext("field", 10, []string{"20"}, map[string]interface{}{})))

	assert.True(t, validateMin(newTestContext("field", 2.0, []string{"2"}, map[string]interface{}{})))
	assert.False(t, validateMin(newTestContext("field", 10.0, []string{"20"}, map[string]interface{}{})))
	assert.True(t, validateMin(newTestContext("field", 3.7, []string{"2.5"}, map[string]interface{}{})))
	assert.False(t, validateMin(newTestContext("field", 10.0, []string{"20.4"}, map[string]interface{}{})))

	assert.True(t, validateMin(newTestContext("field", []int{5, 4}, []string{"2"}, map[string]interface{}{})))
	assert.False(t, validateMin(newTestContext("field", []int{5, 4, 3, 2}, []string{"20"}, map[string]interface{}{})))

	assert.True(t, validateMin(newTestContext("field", []string{"5", "4"}, []string{"2"}, map[string]interface{}{})))
	assert.False(t, validateMin(newTestContext("field", []string{"5", "4", "3", "2"}, []string{"20"}, map[string]interface{}{})))

	assert.True(t, validateMin(newTestContext("field", true, []string{"2"}, map[string]interface{}{})))
	assert.Panics(t, func() { validateMin(newTestContext("field", true, []string{"test"}, map[string]interface{}{})) })

	assert.True(t, validateMin(newTestContext("file", createTestFiles(largeLogoPath), []string{"2"}, map[string]interface{}{})))
	assert.True(t, validateMin(newTestContext("file", createTestFiles(mediumLogoPath, largeLogoPath), []string{"1"}, map[string]interface{}{})))
	assert.False(t, validateMin(newTestContext("file", createTestFiles(logoPath), []string{"1"}, map[string]interface{}{})))
	assert.False(t, validateMin(newTestContext("file", createTestFiles(logoPath, largeLogoPath), []string{"1"}, map[string]interface{}{})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "min"},
			},
		}
		field.Check()
	})
}

func TestValidateMax(t *testing.T) {
	assert.True(t, validateMax(newTestContext("field", "not numeric", []string{"12"}, map[string]interface{}{})))
	assert.False(t, validateMax(newTestContext("field", "not numeric", []string{"5"}, map[string]interface{}{})))
	assert.True(t, validateMax(newTestContext("field", "🇩🇪🇩🇪👍🏼", []string{"5"}, map[string]interface{}{})))
	assert.True(t, validateMax(newTestContext("field", "🇩🇪🇩🇪👍🏼👍🏼👍🏼", []string{"5"}, map[string]interface{}{})))
	assert.False(t, validateMax(newTestContext("field", "🇩🇪🇩🇪👍🏼👍🏼👍🏼👍🏼", []string{"5"}, map[string]interface{}{})))

	assert.True(t, validateMax(newTestContext("field", 1, []string{"2"}, map[string]interface{}{})))
	assert.False(t, validateMax(newTestContext("field", 20, []string{"10"}, map[string]interface{}{})))

	assert.True(t, validateMax(newTestContext("field", 2.0, []string{"2"}, map[string]interface{}{})))
	assert.False(t, validateMax(newTestContext("field", 10.0, []string{"5"}, map[string]interface{}{})))
	assert.True(t, validateMax(newTestContext("field", 2.5, []string{"3.7"}, map[string]interface{}{})))
	assert.False(t, validateMax(newTestContext("field", 20.4, []string{"10.0"}, map[string]interface{}{})))

	assert.True(t, validateMax(newTestContext("field", []int{5, 4}, []string{"2"}, map[string]interface{}{})))
	assert.False(t, validateMax(newTestContext("field", []int{5, 4, 3, 2}, []string{"3"}, map[string]interface{}{})))

	assert.True(t, validateMax(newTestContext("field", []string{"5", "4"}, []string{"3"}, map[string]interface{}{})))
	assert.False(t, validateMax(newTestContext("field", []string{"5", "4", "3", "2"}, []string{"2"}, map[string]interface{}{})))

	assert.True(t, validateMax(newTestContext("field", true, []string{"2"}, map[string]interface{}{})))
	assert.Panics(t, func() { validateMax(newTestContext("field", true, []string{"test"}, map[string]interface{}{})) })

	assert.False(t, validateMax(newTestContext("file", createTestFiles(largeLogoPath), []string{"2"}, map[string]interface{}{})))
	assert.False(t, validateMax(newTestContext("file", createTestFiles(mediumLogoPath, largeLogoPath), []string{"1"}, map[string]interface{}{})))
	assert.True(t, validateMax(newTestContext("file", createTestFiles(logoPath), []string{"1"}, map[string]interface{}{})))
	assert.True(t, validateMax(newTestContext("file", createTestFiles(logoPath, configPath), []string{"1"}, map[string]interface{}{})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "max"},
			},
		}
		field.Check()
	})
}

func TestValidateBetween(t *testing.T) {
	assert.True(t, validateBetween(newTestContext("field", "not numeric", []string{"5", "12"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", "not numeric", []string{"12", "20"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", "not numeric", []string{"5", "6"}, map[string]interface{}{})))

	assert.False(t, validateBetween(newTestContext("field", "🇩🇪", []string{"2", "5"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", "👍🏼", []string{"2", "5"}, map[string]interface{}{})))
	assert.True(t, validateBetween(newTestContext("field", "👍🏼🇩🇪", []string{"2", "5"}, map[string]interface{}{})))
	assert.True(t, validateBetween(newTestContext("field", "🇩🇪🇩🇪👍🏼👍🏼👍🏼", []string{"2", "5"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", "🇩🇪🇩🇪👍🏼👍🏼👍🏼👍🏼", []string{"2", "5"}, map[string]interface{}{})))

	assert.True(t, validateBetween(newTestContext("field", 1, []string{"0", "3"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", 20, []string{"5", "10"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", 20, []string{"21", "23"}, map[string]interface{}{})))

	assert.True(t, validateBetween(newTestContext("field", 2.0, []string{"2", "5"}, map[string]interface{}{})))
	assert.True(t, validateBetween(newTestContext("field", 2.0, []string{"1.0", "5.0"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", 10.0, []string{"5", "7"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", 10.0, []string{"15", "17"}, map[string]interface{}{})))
	assert.True(t, validateBetween(newTestContext("field", 2.5, []string{"1.7", "3.7"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", 20.4, []string{"10.0", "14.7"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", 20.4, []string{"25.0", "54.7"}, map[string]interface{}{})))

	assert.True(t, validateBetween(newTestContext("field", []int{5, 4}, []string{"1", "5"}, map[string]interface{}{})))
	assert.True(t, validateBetween(newTestContext("field", []int{5, 4}, []string{"2.2", "5.7"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", []int{5, 4, 3, 2}, []string{"1", "3"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", []int{5, 4, 3, 2}, []string{"5", "7"}, map[string]interface{}{})))

	assert.True(t, validateBetween(newTestContext("field", []string{"5", "4"}, []string{"1", "5"}, map[string]interface{}{})))
	assert.True(t, validateBetween(newTestContext("field", []string{"5", "4"}, []string{"2.2", "5.7"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", []string{"5", "4", "3", "2"}, []string{"1", "3"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("field", []string{"5", "4", "3", "2"}, []string{"5", "7"}, map[string]interface{}{})))

	assert.True(t, validateBetween(newTestContext("field", true, []string{"2", "3"}, map[string]interface{}{})))
	assert.Panics(t, func() { validateBetween(newTestContext("field", true, []string{"test"}, map[string]interface{}{})) })
	assert.Panics(t, func() { validateBetween(newTestContext("field", true, []string{"1"}, map[string]interface{}{})) })
	assert.Panics(t, func() {
		validateBetween(newTestContext("field", true, []string{"test", "2"}, map[string]interface{}{}))
	})
	assert.Panics(t, func() {
		validateBetween(newTestContext("field", true, []string{"2", "test"}, map[string]interface{}{}))
	})

	assert.True(t, validateBetween(newTestContext("file", createTestFiles(largeLogoPath), []string{"2", "50"}, map[string]interface{}{})))
	assert.True(t, validateBetween(newTestContext("file", createTestFiles(mediumLogoPath, largeLogoPath), []string{"8", "42"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("file", createTestFiles(logoPath), []string{"5", "10"}, map[string]interface{}{})))
	assert.False(t, validateBetween(newTestContext("file", createTestFiles(logoPath, mediumLogoPath), []string{"5", "10"}, map[string]interface{}{})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "between"},
			},
		}
		field.Check()
	})

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "between", Params: []string{"2"}},
			},
		}
		field.Check()
	})
}

func TestValidateGreaterThan(t *testing.T) {
	assert.True(t, validateGreaterThan(newTestContext("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 2})))
	assert.False(t, validateGreaterThan(newTestContext("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 20})))

	assert.True(t, validateGreaterThan(newTestContext("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 2.0})))
	assert.False(t, validateGreaterThan(newTestContext("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 5.1})))

	assert.True(t, validateGreaterThan(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "str"})))
	assert.False(t, validateGreaterThan(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "other string"})))

	assert.True(t, validateGreaterThan(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "👍🏼👍🏼👍🏼👍🏼👍🏼"})))
	assert.False(t, validateGreaterThan(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "👍🏼👍🏼👍🏼👍🏼👍🏼👍🏼"})))

	assert.True(t, validateGreaterThan(newTestContext("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1}})))
	assert.False(t, validateGreaterThan(newTestContext("field", []int{6}, []string{"comparison"}, map[string]interface{}{"field": []int{6}, "comparison": []int{1, 2, 3}})))

	// Different type
	assert.False(t, validateGreaterThan(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": 2})))

	// Missing field
	assert.False(t, validateGreaterThan(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string"})))

	// Unsupported type
	test := "string"
	assert.True(t, validateGreaterThan(newTestContext("field", &test, []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": &test})))

	files := createTestFiles(largeLogoPath)
	otherFiles := createTestFiles(logoPath)
	assert.True(t, validateGreaterThan(newTestContext("file", files, []string{"otherFiles"}, map[string]interface{}{"file": files, "otherFiles": otherFiles})))
	assert.False(t, validateGreaterThan(newTestContext("file", otherFiles, []string{"file"}, map[string]interface{}{"file": files, "otherFiles": otherFiles})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "greater_than"},
			},
		}
		field.Check()
	})

	// Objects
	data := map[string]interface{}{
		"cart": map[string]interface{}{
			"count": 5,
		},
		"constraints": map[string]interface{}{
			"min_products": 1,
		},
	}
	assert.True(t, validateGreaterThan(newTestContext("cart.count", 5, []string{"constraints.min_products"}, data)))
	assert.False(t, validateGreaterThan(newTestContext("cart.count", 0, []string{"constraints.min_products"}, data)))
}

func TestValidateGreaterThanEqual(t *testing.T) {
	assert.True(t, validateGreaterThanEqual(newTestContext("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 2})))
	assert.True(t, validateGreaterThanEqual(newTestContext("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 5})))
	assert.False(t, validateGreaterThanEqual(newTestContext("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 20})))
	assert.False(t, validateGreaterThanEqual(newTestContext("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 5.1})))

	assert.True(t, validateGreaterThanEqual(newTestContext("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 2.0})))
	assert.True(t, validateGreaterThanEqual(newTestContext("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 5.0})))
	assert.False(t, validateGreaterThanEqual(newTestContext("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 5.1})))

	assert.True(t, validateGreaterThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "str"})))
	assert.True(t, validateGreaterThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "gnirts"})))
	assert.False(t, validateGreaterThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "other string"})))

	assert.True(t, validateGreaterThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "👍🏼👍🏼👍🏼👍🏼👍🏼"})))
	assert.True(t, validateGreaterThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "👍🏼👍🏼👍🏼👍🏼👍🏼👍🏼"})))
	assert.False(t, validateGreaterThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "👍🏼👍🏼👍🏼👍🏼👍🏼👍🏼👍🏼"})))

	assert.True(t, validateGreaterThanEqual(newTestContext("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1}})))
	assert.True(t, validateGreaterThanEqual(newTestContext("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1, 2}})))
	assert.False(t, validateGreaterThanEqual(newTestContext("field", []int{6}, []string{"comparison"}, map[string]interface{}{"field": []int{6}, "comparison": []int{1, 2, 3}})))

	// Different type
	assert.False(t, validateGreaterThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": 2})))

	// Missing field
	assert.False(t, validateGreaterThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string"})))

	// Unsupported type
	test := "string"
	assert.True(t, validateGreaterThanEqual(newTestContext("field", &test, []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": &test})))

	files := createTestFiles(largeLogoPath)
	otherFiles := createTestFiles(logoPath)
	assert.True(t, validateGreaterThanEqual(newTestContext("file", files, []string{"otherFiles"}, map[string]interface{}{"file": files, "otherFiles": otherFiles})))
	assert.False(t, validateGreaterThanEqual(newTestContext("file", otherFiles, []string{"file"}, map[string]interface{}{"file": files, "otherFiles": otherFiles})))

	files = createTestFiles(logoPath)
	otherFiles = createTestFiles(logoPath)
	assert.True(t, validateGreaterThanEqual(newTestContext("file", files, []string{"otherFiles"}, map[string]interface{}{"file": files, "otherFiles": otherFiles})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "greater_than_equal"},
			},
		}
		field.Check()
	})

	// Objects
	data := map[string]interface{}{
		"cart": map[string]interface{}{
			"count": 5,
		},
		"constraints": map[string]interface{}{
			"min_products": 1,
		},
	}
	assert.True(t, validateGreaterThanEqual(newTestContext("cart.count", 5, []string{"constraints.min_products"}, data)))
	assert.True(t, validateGreaterThanEqual(newTestContext("cart.count", 1, []string{"constraints.min_products"}, data)))
	assert.False(t, validateGreaterThanEqual(newTestContext("cart.count", 0, []string{"constraints.min_products"}, data)))
}

func TestValidateLowerThan(t *testing.T) {
	assert.True(t, validateLowerThan(newTestContext("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 7})))
	assert.False(t, validateLowerThan(newTestContext("field", 20, []string{"comparison"}, map[string]interface{}{"field": 20, "comparison": 5})))

	assert.True(t, validateLowerThan(newTestContext("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 7.0})))
	assert.False(t, validateLowerThan(newTestContext("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 4.9})))

	assert.True(t, validateLowerThan(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "other string"})))
	assert.False(t, validateLowerThan(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "str"})))

	assert.True(t, validateLowerThan(newTestContext("field", "👍🏼👍🏼", []string{"comparison"}, map[string]interface{}{"field": "👍🏼👍🏼", "comparison": "str"})))
	assert.False(t, validateLowerThan(newTestContext("field", "st", []string{"comparison"}, map[string]interface{}{"field": "st", "comparison": "👍🏼"})))

	assert.True(t, validateLowerThan(newTestContext("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1, 2, 3}})))
	assert.False(t, validateLowerThan(newTestContext("field", []int{6, 7, 8}, []string{"comparison"}, map[string]interface{}{"field": []int{6, 7, 8}, "comparison": []int{1, 2}})))

	// Different type
	assert.False(t, validateLowerThan(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": 2})))

	// Missing field
	assert.False(t, validateLowerThan(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string"})))

	// Unsupported type
	test := "string"
	assert.True(t, validateLowerThan(newTestContext("field", &test, []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": &test})))

	files := createTestFiles(logoPath)
	otherFiles := createTestFiles(largeLogoPath)
	assert.True(t, validateLowerThan(newTestContext("file", files, []string{"otherFiles"}, map[string]interface{}{"file": files, "otherFiles": otherFiles})))
	assert.False(t, validateLowerThan(newTestContext("file", otherFiles, []string{"file"}, map[string]interface{}{"file": files, "otherFiles": otherFiles})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "lower_than"},
			},
		}
		field.Check()
	})

	// Objects
	data := map[string]interface{}{
		"cart": map[string]interface{}{
			"count": 1,
		},
		"constraints": map[string]interface{}{
			"max_products": 5,
		},
	}
	assert.True(t, validateLowerThan(newTestContext("cart.count", 1, []string{"constraints.max_products"}, data)))
	assert.False(t, validateLowerThan(newTestContext("cart.count", 5, []string{"constraints.max_products"}, data)))
}

func TestValidateLowerThanEqual(t *testing.T) {
	assert.True(t, validateLowerThanEqual(newTestContext("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 7})))
	assert.True(t, validateLowerThanEqual(newTestContext("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 5})))
	assert.False(t, validateLowerThanEqual(newTestContext("field", 20, []string{"comparison"}, map[string]interface{}{"field": 20, "comparison": 5})))
	assert.False(t, validateLowerThanEqual(newTestContext("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 4.9})))

	assert.True(t, validateLowerThanEqual(newTestContext("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 7.0})))
	assert.True(t, validateLowerThanEqual(newTestContext("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 5.0})))
	assert.False(t, validateLowerThanEqual(newTestContext("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 4.9})))

	assert.True(t, validateLowerThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "other string"})))
	assert.True(t, validateLowerThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "gnirts"})))
	assert.False(t, validateLowerThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "str"})))

	assert.True(t, validateLowerThanEqual(newTestContext("field", "👍🏼👍🏼", []string{"comparison"}, map[string]interface{}{"field": "👍🏼👍🏼", "comparison": "str"})))
	assert.True(t, validateLowerThanEqual(newTestContext("field", "👍🏼👍🏼👍🏼", []string{"comparison"}, map[string]interface{}{"field": "👍🏼👍🏼👍🏼", "comparison": "str"})))
	assert.False(t, validateLowerThanEqual(newTestContext("field", "st", []string{"comparison"}, map[string]interface{}{"field": "st", "comparison": "👍🏼"})))

	assert.True(t, validateLowerThanEqual(newTestContext("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1, 2, 3}})))
	assert.True(t, validateLowerThanEqual(newTestContext("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1, 2}})))
	assert.False(t, validateLowerThanEqual(newTestContext("field", []int{6, 7, 8}, []string{"comparison"}, map[string]interface{}{"field": []int{6, 7, 8}, "comparison": []int{1, 2}})))

	// Different type
	assert.False(t, validateLowerThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": 2})))

	// Missing field
	assert.False(t, validateLowerThanEqual(newTestContext("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string"})))

	// Unsupported type
	test := "string"
	assert.True(t, validateLowerThanEqual(newTestContext("field", &test, []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": &test})))

	files := createTestFiles(logoPath)
	otherFiles := createTestFiles(largeLogoPath)
	assert.True(t, validateLowerThanEqual(newTestContext("file", files, []string{"otherFiles"}, map[string]interface{}{"file": files, "otherFiles": otherFiles})))
	assert.False(t, validateLowerThanEqual(newTestContext("file", otherFiles, []string{"file"}, map[string]interface{}{"file": files, "otherFiles": otherFiles})))

	files = createTestFiles(logoPath)
	otherFiles = createTestFiles(logoPath)
	assert.True(t, validateLowerThanEqual(newTestContext("file", files, []string{"otherFiles"}, map[string]interface{}{"file": files, "otherFiles": otherFiles})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "lower_than_equal"},
			},
		}
		field.Check()
	})

	// Objects
	data := map[string]interface{}{
		"cart": map[string]interface{}{
			"count": 1,
		},
		"constraints": map[string]interface{}{
			"max_products": 5,
		},
	}
	assert.True(t, validateLowerThanEqual(newTestContext("cart.count", 1, []string{"constraints.max_products"}, data)))
	assert.True(t, validateLowerThanEqual(newTestContext("cart.count", 5, []string{"constraints.max_products"}, data)))
	assert.False(t, validateLowerThanEqual(newTestContext("cart.count", 6, []string{"constraints.max_products"}, data)))
}

func TestValidateBool(t *testing.T) {
	data := map[string]interface{}{
		"field": 1,
	}
	assert.True(t, validateBool(newTestContext("field", 1, []string{}, data)))
	assert.True(t, validateBool(newTestContext("field", 0, []string{}, data)))
	assert.True(t, validateBool(newTestContext("field", "on", []string{}, data)))
	assert.True(t, validateBool(newTestContext("field", "off", []string{}, data)))
	assert.True(t, validateBool(newTestContext("field", "true", []string{}, data)))
	assert.True(t, validateBool(newTestContext("field", "false", []string{}, data)))
	assert.True(t, validateBool(newTestContext("field", "yes", []string{}, data)))
	assert.True(t, validateBool(newTestContext("field", "no", []string{}, data)))
	assert.True(t, validateBool(newTestContext("field", true, []string{}, data)))
	assert.True(t, validateBool(newTestContext("field", false, []string{}, data)))

	assert.False(t, validateBool(newTestContext("field", 0.0, []string{}, data)))
	assert.False(t, validateBool(newTestContext("field", 1.0, []string{}, data)))
	assert.False(t, validateBool(newTestContext("field", []string{"true"}, []string{}, data)))
	assert.False(t, validateBool(newTestContext("field", -1, []string{}, data)))
}

func TestValidateBoolConvert(t *testing.T) {
	form := map[string]interface{}{"field": "on"}
	ctx := newTestContext("field", form["field"], []string{}, form)
	assert.True(t, validateBool(ctx))
	b, ok := ctx.Value.(bool)
	assert.True(t, ok)
	assert.True(t, b)

	form = map[string]interface{}{"field": "off"}
	ctx = newTestContext("field", form["field"], []string{}, form)
	assert.True(t, validateBool(ctx))
	b, ok = ctx.Value.(bool)
	assert.True(t, ok)
	assert.False(t, b)

	form = map[string]interface{}{"field": 1}
	ctx = newTestContext("field", form["field"], []string{}, form)
	assert.True(t, validateBool(ctx))
	b, ok = ctx.Value.(bool)
	assert.True(t, ok)
	assert.True(t, b)

	form = map[string]interface{}{"field": 0}
	ctx = newTestContext("field", form["field"], []string{}, form)
	assert.True(t, validateBool(ctx))
	b, ok = ctx.Value.(bool)
	assert.True(t, ok)
	assert.False(t, b)
}

func TestValidateBoolConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"bool": 1,
		},
	}

	set := RuleSet{
		"object":      List{"required", "object"},
		"object.bool": List{"required", "bool"},
	}

	errors := Validate(data, set, true, "en-US")
	assert.Empty(t, errors)
	_, ok := data["object"].(map[string]interface{})["bool"].(bool)
	assert.True(t, ok)
}

func TestValidateSame(t *testing.T) {
	assert.True(t, validateSame(newTestContext("field", "password", []string{"other"}, map[string]interface{}{"field": "password", "other": "password"})))
	assert.True(t, validateSame(newTestContext("field", 1, []string{"other"}, map[string]interface{}{"field": 1, "other": 1})))
	assert.True(t, validateSame(newTestContext("field", 1.2, []string{"other"}, map[string]interface{}{"field": 1.2, "other": 1.2})))
	assert.True(t, validateSame(newTestContext("field", []string{"one", "two", "three"}, []string{"other"}, map[string]interface{}{"field": []string{"one", "two", "three"}, "other": []string{"one", "two", "three"}})))

	assert.False(t, validateSame(newTestContext("field", 1, []string{"other"}, map[string]interface{}{"field": 1, "other": 2})))
	assert.False(t, validateSame(newTestContext("field", 1.1, []string{"other"}, map[string]interface{}{"field": 1.1, "other": 1})))
	assert.False(t, validateSame(newTestContext("field", "password", []string{"other"}, map[string]interface{}{"field": "password", "other": "not password"})))
	assert.False(t, validateSame(newTestContext("field", "no other", []string{"other"}, map[string]interface{}{"field": "no other"})))
	assert.False(t, validateSame(newTestContext("field", []string{"one", "two"}, []string{"other"}, map[string]interface{}{"field": []string{"one", "two"}, "other": []string{"one", "two", "three"}})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "same"},
			},
		}
		field.Check()
	})

	data := map[string]interface{}{
		"object": map[string]interface{}{
			"key": "value",
			"deep": map[string]interface{}{
				"a": 1,
			},
		},
		"other": map[string]interface{}{
			"key": "value",
			"deep": map[string]interface{}{
				"a": 1,
			},
		},
	}
	assert.True(t, validateSame(newTestContext("object", data["object"], []string{"other"}, data)))
	assert.True(t, validateSame(newTestContext("object.deep", data["object"].(map[string]interface{})["deep"], []string{"other.deep"}, data)))
	assert.False(t, validateSame(newTestContext("object", data["object"], []string{"other.deep"}, data)))
}

func TestValidateDifferent(t *testing.T) {
	assert.False(t, validateDifferent(newTestContext("field", "password", []string{"other"}, map[string]interface{}{"field": "password", "other": "password"})))
	assert.False(t, validateDifferent(newTestContext("field", 1, []string{"other"}, map[string]interface{}{"field": 1, "other": 1})))
	assert.False(t, validateDifferent(newTestContext("field", 1.2, []string{"other"}, map[string]interface{}{"field": 1.2, "other": 1.2})))
	assert.False(t, validateDifferent(newTestContext("field", []string{"one", "two", "three"}, []string{"other"}, map[string]interface{}{"field": []string{"one", "two", "three"}, "other": []string{"one", "two", "three"}})))

	assert.True(t, validateDifferent(newTestContext("field", 1, []string{"other"}, map[string]interface{}{"field": 1, "other": 2})))
	assert.True(t, validateDifferent(newTestContext("field", 1.1, []string{"other"}, map[string]interface{}{"field": 1.1, "other": 1})))
	assert.True(t, validateDifferent(newTestContext("field", "password", []string{"other"}, map[string]interface{}{"field": "password", "other": "not password"})))
	assert.True(t, validateDifferent(newTestContext("field", "no other", []string{"other"}, map[string]interface{}{"field": "no other"})))
	assert.True(t, validateDifferent(newTestContext("field", []string{"one", "two"}, []string{"other"}, map[string]interface{}{"field": []string{"one", "two"}, "other": []string{"one", "two", "three"}})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "different"},
			},
		}
		field.Check()
	})
}

func TestValidateSize(t *testing.T) {
	assert.True(t, validateSize(newTestContext("field", "123", []string{"3"}, map[string]interface{}{})))
	assert.True(t, validateSize(newTestContext("field", "", []string{"0"}, map[string]interface{}{})))
	assert.False(t, validateSize(newTestContext("field", "4567", []string{"5"}, map[string]interface{}{})))
	assert.False(t, validateSize(newTestContext("field", "4567", []string{"2"}, map[string]interface{}{})))

	assert.True(t, validateSize(newTestContext("field", "🇩🇪👍🏼", []string{"2"}, map[string]interface{}{})))
	assert.True(t, validateSize(newTestContext("field", "👍🏼!", []string{"2"}, map[string]interface{}{})))
	assert.False(t, validateSize(newTestContext("field", "👍🏼", []string{"2"}, map[string]interface{}{})))

	assert.False(t, validateSize(newTestContext("field", 4567, []string{"2"}, map[string]interface{}{})))
	assert.False(t, validateSize(newTestContext("field", 4567.8, []string{"2"}, map[string]interface{}{})))

	// Unsupported type
	assert.True(t, validateSize(newTestContext("field", true, []string{"2"}, map[string]interface{}{})))

	assert.Panics(t, func() { validateSize(newTestContext("field", "123", []string{"test"}, map[string]interface{}{})) })

	assert.True(t, validateSize(newTestContext("field", []string{"a", "b", "c"}, []string{"3"}, map[string]interface{}{})))
	assert.False(t, validateSize(newTestContext("field", []string{"a", "b", "c", "d"}, []string{"3"}, map[string]interface{}{})))

	assert.True(t, validateSize(newTestContext("field", 5, []string{"5"}, map[string]interface{}{})))
	assert.False(t, validateSize(newTestContext("field", 3, []string{"5"}, map[string]interface{}{})))

	assert.True(t, validateSize(newTestContext("file", createTestFiles(logoPath), []string{"1"}, map[string]interface{}{})))
	assert.True(t, validateSize(newTestContext("file", createTestFiles(largeLogoPath), []string{"42"}, map[string]interface{}{})))
	assert.False(t, validateSize(newTestContext("file", createTestFiles(logoPath), []string{"3"}, map[string]interface{}{})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "size"},
			},
		}
		field.Check()
	})
}

func TestValidateObject(t *testing.T) {
	assert.False(t, validateObject(newTestContext("field", "123", []string{}, map[string]interface{}{})))
	assert.True(t, validateObject(newTestContext("field", map[string]interface{}{"hello": "world"}, []string{}, map[string]interface{}{})))
}
