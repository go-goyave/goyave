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
	assert.True(t, validateRequired("field", nil, []string{}, map[string]interface{}{"field": nil}))
	assert.False(t, validateRequired("field", "", []string{}, map[string]interface{}{"field": ""}))

	data := map[string]interface{}{
		"object": map[string]interface{}{
			"key": "value",
		},
	}
	assert.True(t, validateRequired("object.key", "value", []string{}, data))
	assert.False(t, validateRequired("object.notakey", nil, []string{}, data))
}

func TestValidateMin(t *testing.T) {
	assert.True(t, validateMin("field", "not numeric", []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateMin("field", "not numeric", []string{"20"}, map[string]interface{}{}))
	assert.False(t, validateMin("field", "🇩🇪", []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateMin("field", "👍🏼", []string{"2"}, map[string]interface{}{}))

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

	assert.True(t, validateMin("file", createTestFiles(largeLogoPath), []string{"2"}, map[string]interface{}{}))
	assert.True(t, validateMin("file", createTestFiles(mediumLogoPath, largeLogoPath), []string{"1"}, map[string]interface{}{}))
	assert.False(t, validateMin("file", createTestFiles(logoPath), []string{"1"}, map[string]interface{}{}))
	assert.False(t, validateMin("file", createTestFiles(logoPath, largeLogoPath), []string{"1"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "min"},
			},
		}
		field.check()
	})
}

func TestValidateMax(t *testing.T) {
	assert.True(t, validateMax("field", "not numeric", []string{"12"}, map[string]interface{}{}))
	assert.False(t, validateMax("field", "not numeric", []string{"5"}, map[string]interface{}{}))
	assert.True(t, validateMax("field", "🇩🇪🇩🇪👍🏼", []string{"5"}, map[string]interface{}{}))
	assert.True(t, validateMax("field", "🇩🇪🇩🇪👍🏼👍🏼👍🏼", []string{"5"}, map[string]interface{}{}))
	assert.False(t, validateMax("field", "🇩🇪🇩🇪👍🏼👍🏼👍🏼👍🏼", []string{"5"}, map[string]interface{}{}))

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

	assert.False(t, validateMax("file", createTestFiles(largeLogoPath), []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateMax("file", createTestFiles(mediumLogoPath, largeLogoPath), []string{"1"}, map[string]interface{}{}))
	assert.True(t, validateMax("file", createTestFiles(logoPath), []string{"1"}, map[string]interface{}{}))
	assert.True(t, validateMax("file", createTestFiles(logoPath, configPath), []string{"1"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "max"},
			},
		}
		field.check()
	})
}

func TestValidateBetween(t *testing.T) {
	assert.True(t, validateBetween("field", "not numeric", []string{"5", "12"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", "not numeric", []string{"12", "20"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", "not numeric", []string{"5", "6"}, map[string]interface{}{}))

	assert.False(t, validateBetween("field", "🇩🇪", []string{"2", "5"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", "👍🏼", []string{"2", "5"}, map[string]interface{}{}))
	assert.True(t, validateBetween("field", "👍🏼🇩🇪", []string{"2", "5"}, map[string]interface{}{}))
	assert.True(t, validateBetween("field", "🇩🇪🇩🇪👍🏼👍🏼👍🏼", []string{"2", "5"}, map[string]interface{}{}))
	assert.False(t, validateBetween("field", "🇩🇪🇩🇪👍🏼👍🏼👍🏼👍🏼", []string{"2", "5"}, map[string]interface{}{}))

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

	assert.True(t, validateBetween("file", createTestFiles(largeLogoPath), []string{"2", "50"}, map[string]interface{}{}))
	assert.True(t, validateBetween("file", createTestFiles(mediumLogoPath, largeLogoPath), []string{"8", "42"}, map[string]interface{}{}))
	assert.False(t, validateBetween("file", createTestFiles(logoPath), []string{"5", "10"}, map[string]interface{}{}))
	assert.False(t, validateBetween("file", createTestFiles(logoPath, mediumLogoPath), []string{"5", "10"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "between"},
			},
		}
		field.check()
	})

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "between", Params: []string{"2"}},
			},
		}
		field.check()
	})
}

func TestValidateGreaterThan(t *testing.T) {
	assert.True(t, validateGreaterThan("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 2}))
	assert.False(t, validateGreaterThan("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 20}))

	assert.True(t, validateGreaterThan("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 2.0}))
	assert.False(t, validateGreaterThan("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 5.1}))

	assert.True(t, validateGreaterThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "str"}))
	assert.False(t, validateGreaterThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "other string"}))

	assert.True(t, validateGreaterThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "👍🏼👍🏼👍🏼👍🏼👍🏼"}))
	assert.False(t, validateGreaterThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "👍🏼👍🏼👍🏼👍🏼👍🏼👍🏼"}))

	assert.True(t, validateGreaterThan("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1}}))
	assert.False(t, validateGreaterThan("field", []int{6}, []string{"comparison"}, map[string]interface{}{"field": []int{6}, "comparison": []int{1, 2, 3}}))

	// Different type
	assert.False(t, validateGreaterThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": 2}))

	// Missing field
	assert.False(t, validateGreaterThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string"}))

	// Unsupported type
	test := "string"
	assert.True(t, validateGreaterThan("field", &test, []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": &test}))

	files := createTestFiles(largeLogoPath)
	otherFiles := createTestFiles(logoPath)
	assert.True(t, validateGreaterThan("file", files, []string{"otherFiles"}, map[string]interface{}{"file": files, "otherFiles": otherFiles}))
	assert.False(t, validateGreaterThan("file", otherFiles, []string{"file"}, map[string]interface{}{"file": files, "otherFiles": otherFiles}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "greater_than"},
			},
		}
		field.check()
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
	assert.True(t, validateGreaterThan("cart.count", 5, []string{"constraints.min_products"}, data))
	assert.False(t, validateGreaterThan("cart.count", 0, []string{"constraints.min_products"}, data))
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

	assert.True(t, validateGreaterThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "👍🏼👍🏼👍🏼👍🏼👍🏼"}))
	assert.True(t, validateGreaterThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "👍🏼👍🏼👍🏼👍🏼👍🏼👍🏼"}))
	assert.False(t, validateGreaterThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "👍🏼👍🏼👍🏼👍🏼👍🏼👍🏼👍🏼"}))

	assert.True(t, validateGreaterThanEqual("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1}}))
	assert.True(t, validateGreaterThanEqual("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1, 2}}))
	assert.False(t, validateGreaterThanEqual("field", []int{6}, []string{"comparison"}, map[string]interface{}{"field": []int{6}, "comparison": []int{1, 2, 3}}))

	// Different type
	assert.False(t, validateGreaterThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": 2}))

	// Missing field
	assert.False(t, validateGreaterThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string"}))

	// Unsupported type
	test := "string"
	assert.True(t, validateGreaterThanEqual("field", &test, []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": &test}))

	files := createTestFiles(largeLogoPath)
	otherFiles := createTestFiles(logoPath)
	assert.True(t, validateGreaterThanEqual("file", files, []string{"otherFiles"}, map[string]interface{}{"file": files, "otherFiles": otherFiles}))
	assert.False(t, validateGreaterThanEqual("file", otherFiles, []string{"file"}, map[string]interface{}{"file": files, "otherFiles": otherFiles}))

	files = createTestFiles(logoPath)
	otherFiles = createTestFiles(logoPath)
	assert.True(t, validateGreaterThanEqual("file", files, []string{"otherFiles"}, map[string]interface{}{"file": files, "otherFiles": otherFiles}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "greater_than_equal"},
			},
		}
		field.check()
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
	assert.True(t, validateGreaterThanEqual("cart.count", 5, []string{"constraints.min_products"}, data))
	assert.True(t, validateGreaterThanEqual("cart.count", 1, []string{"constraints.min_products"}, data))
	assert.False(t, validateGreaterThanEqual("cart.count", 0, []string{"constraints.min_products"}, data))
}

func TestValidateLowerThan(t *testing.T) {
	assert.True(t, validateLowerThan("field", 5, []string{"comparison"}, map[string]interface{}{"field": 5, "comparison": 7}))
	assert.False(t, validateLowerThan("field", 20, []string{"comparison"}, map[string]interface{}{"field": 20, "comparison": 5}))

	assert.True(t, validateLowerThan("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 7.0}))
	assert.False(t, validateLowerThan("field", 5.0, []string{"comparison"}, map[string]interface{}{"field": 5.0, "comparison": 4.9}))

	assert.True(t, validateLowerThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "other string"}))
	assert.False(t, validateLowerThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": "str"}))

	assert.True(t, validateLowerThan("field", "👍🏼👍🏼", []string{"comparison"}, map[string]interface{}{"field": "👍🏼👍🏼", "comparison": "str"}))
	assert.False(t, validateLowerThan("field", "st", []string{"comparison"}, map[string]interface{}{"field": "st", "comparison": "👍🏼"}))

	assert.True(t, validateLowerThan("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1, 2, 3}}))
	assert.False(t, validateLowerThan("field", []int{6, 7, 8}, []string{"comparison"}, map[string]interface{}{"field": []int{6, 7, 8}, "comparison": []int{1, 2}}))

	// Different type
	assert.False(t, validateLowerThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": 2}))

	// Missing field
	assert.False(t, validateLowerThan("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string"}))

	// Unsupported type
	test := "string"
	assert.True(t, validateLowerThan("field", &test, []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": &test}))

	files := createTestFiles(logoPath)
	otherFiles := createTestFiles(largeLogoPath)
	assert.True(t, validateLowerThan("file", files, []string{"otherFiles"}, map[string]interface{}{"file": files, "otherFiles": otherFiles}))
	assert.False(t, validateLowerThan("file", otherFiles, []string{"file"}, map[string]interface{}{"file": files, "otherFiles": otherFiles}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "lower_than"},
			},
		}
		field.check()
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
	assert.True(t, validateLowerThan("cart.count", 1, []string{"constraints.max_products"}, data))
	assert.False(t, validateLowerThan("cart.count", 5, []string{"constraints.max_products"}, data))
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

	assert.True(t, validateLowerThanEqual("field", "👍🏼👍🏼", []string{"comparison"}, map[string]interface{}{"field": "👍🏼👍🏼", "comparison": "str"}))
	assert.True(t, validateLowerThanEqual("field", "👍🏼👍🏼👍🏼", []string{"comparison"}, map[string]interface{}{"field": "👍🏼👍🏼👍🏼", "comparison": "str"}))
	assert.False(t, validateLowerThanEqual("field", "st", []string{"comparison"}, map[string]interface{}{"field": "st", "comparison": "👍🏼"}))

	assert.True(t, validateLowerThanEqual("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1, 2, 3}}))
	assert.True(t, validateLowerThanEqual("field", []int{5, 2}, []string{"comparison"}, map[string]interface{}{"field": []int{5, 2}, "comparison": []int{1, 2}}))
	assert.False(t, validateLowerThanEqual("field", []int{6, 7, 8}, []string{"comparison"}, map[string]interface{}{"field": []int{6, 7, 8}, "comparison": []int{1, 2}}))

	// Different type
	assert.False(t, validateLowerThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": 2}))

	// Missing field
	assert.False(t, validateLowerThanEqual("field", "string", []string{"comparison"}, map[string]interface{}{"field": "string"}))

	// Unsupported type
	test := "string"
	assert.True(t, validateLowerThanEqual("field", &test, []string{"comparison"}, map[string]interface{}{"field": "string", "comparison": &test}))

	files := createTestFiles(logoPath)
	otherFiles := createTestFiles(largeLogoPath)
	assert.True(t, validateLowerThanEqual("file", files, []string{"otherFiles"}, map[string]interface{}{"file": files, "otherFiles": otherFiles}))
	assert.False(t, validateLowerThanEqual("file", otherFiles, []string{"file"}, map[string]interface{}{"file": files, "otherFiles": otherFiles}))

	files = createTestFiles(logoPath)
	otherFiles = createTestFiles(logoPath)
	assert.True(t, validateLowerThanEqual("file", files, []string{"otherFiles"}, map[string]interface{}{"file": files, "otherFiles": otherFiles}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "lower_than_equal"},
			},
		}
		field.check()
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
	assert.True(t, validateLowerThanEqual("cart.count", 1, []string{"constraints.max_products"}, data))
	assert.True(t, validateLowerThanEqual("cart.count", 5, []string{"constraints.max_products"}, data))
	assert.False(t, validateLowerThanEqual("cart.count", 6, []string{"constraints.max_products"}, data))
}

func TestValidateBool(t *testing.T) {
	data := map[string]interface{}{
		"field": 1,
	}
	assert.True(t, validateBool("field", 1, []string{}, data))
	assert.True(t, validateBool("field", 0, []string{}, data))
	assert.True(t, validateBool("field", "on", []string{}, data))
	assert.True(t, validateBool("field", "off", []string{}, data))
	assert.True(t, validateBool("field", "true", []string{}, data))
	assert.True(t, validateBool("field", "false", []string{}, data))
	assert.True(t, validateBool("field", "yes", []string{}, data))
	assert.True(t, validateBool("field", "no", []string{}, data))
	assert.True(t, validateBool("field", true, []string{}, data))
	assert.True(t, validateBool("field", false, []string{}, data))

	assert.False(t, validateBool("field", 0.0, []string{}, data))
	assert.False(t, validateBool("field", 1.0, []string{}, data))
	assert.False(t, validateBool("field", []string{"true"}, []string{}, data))
	assert.False(t, validateBool("field", -1, []string{}, data))
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

func TestValidateBoolConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"bool": 1,
		},
	}

	set := RuleSet{
		"object":      {"required", "object"},
		"object.bool": {"required", "bool"},
	}

	errors := Validate(data, set, true, "en-US")
	assert.Empty(t, errors)
	_, ok := data["object"].(map[string]interface{})["bool"].(bool)
	assert.True(t, ok)
}

func TestValidateSame(t *testing.T) {
	assert.True(t, validateSame("field", "password", []string{"other"}, map[string]interface{}{"field": "password", "other": "password"}))
	assert.True(t, validateSame("field", 1, []string{"other"}, map[string]interface{}{"field": 1, "other": 1}))
	assert.True(t, validateSame("field", 1.2, []string{"other"}, map[string]interface{}{"field": 1.2, "other": 1.2}))
	assert.True(t, validateSame("field", []string{"one", "two", "three"}, []string{"other"}, map[string]interface{}{"field": []string{"one", "two", "three"}, "other": []string{"one", "two", "three"}}))

	assert.False(t, validateSame("field", 1, []string{"other"}, map[string]interface{}{"field": 1, "other": 2}))
	assert.False(t, validateSame("field", 1.1, []string{"other"}, map[string]interface{}{"field": 1.1, "other": 1}))
	assert.False(t, validateSame("field", "password", []string{"other"}, map[string]interface{}{"field": "password", "other": "not password"}))
	assert.False(t, validateSame("field", "no other", []string{"other"}, map[string]interface{}{"field": "no other"}))
	assert.False(t, validateSame("field", []string{"one", "two"}, []string{"other"}, map[string]interface{}{"field": []string{"one", "two"}, "other": []string{"one", "two", "three"}}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "same"},
			},
		}
		field.check()
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
	assert.True(t, validateSame("object", data["object"], []string{"other"}, data))
	assert.True(t, validateSame("object.deep", data["object"].(map[string]interface{})["deep"], []string{"other.deep"}, data))
	assert.False(t, validateSame("object", data["object"], []string{"other.deep"}, data))
}

func TestValidateDifferent(t *testing.T) {
	assert.False(t, validateDifferent("field", "password", []string{"other"}, map[string]interface{}{"field": "password", "other": "password"}))
	assert.False(t, validateDifferent("field", 1, []string{"other"}, map[string]interface{}{"field": 1, "other": 1}))
	assert.False(t, validateDifferent("field", 1.2, []string{"other"}, map[string]interface{}{"field": 1.2, "other": 1.2}))
	assert.False(t, validateDifferent("field", []string{"one", "two", "three"}, []string{"other"}, map[string]interface{}{"field": []string{"one", "two", "three"}, "other": []string{"one", "two", "three"}}))

	assert.True(t, validateDifferent("field", 1, []string{"other"}, map[string]interface{}{"field": 1, "other": 2}))
	assert.True(t, validateDifferent("field", 1.1, []string{"other"}, map[string]interface{}{"field": 1.1, "other": 1}))
	assert.True(t, validateDifferent("field", "password", []string{"other"}, map[string]interface{}{"field": "password", "other": "not password"}))
	assert.True(t, validateDifferent("field", "no other", []string{"other"}, map[string]interface{}{"field": "no other"}))
	assert.True(t, validateDifferent("field", []string{"one", "two"}, []string{"other"}, map[string]interface{}{"field": []string{"one", "two"}, "other": []string{"one", "two", "three"}}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "different"},
			},
		}
		field.check()
	})
}

func TestValidateConfirmed(t *testing.T) {
	assert.True(t, validateConfirmed("field", "password", []string{}, map[string]interface{}{"field": "password", "field_confirmation": "password"}))
	assert.True(t, validateConfirmed("field", 1, []string{}, map[string]interface{}{"field": 1, "field_confirmation": 1}))
	assert.True(t, validateConfirmed("field", 1.2, []string{}, map[string]interface{}{"field": 1.2, "field_confirmation": 1.2}))
	assert.True(t, validateConfirmed("field", []string{"one", "two", "three"}, []string{}, map[string]interface{}{"field": []string{"one", "two", "three"}, "field_confirmation": []string{"one", "two", "three"}}))

	assert.False(t, validateConfirmed("field", 1, []string{}, map[string]interface{}{"field": 1, "field_confirmation": 2}))
	assert.False(t, validateConfirmed("field", 1.1, []string{}, map[string]interface{}{"field": 1.1, "field_confirmation": 1}))
	assert.False(t, validateConfirmed("field", "password", []string{}, map[string]interface{}{"field": "password", "field_confirmation": "not password"}))
	assert.False(t, validateConfirmed("field", "no confirm", []string{}, map[string]interface{}{"field": "no confirm"}))
	assert.False(t, validateConfirmed("field", []string{"one", "two"}, []string{}, map[string]interface{}{"field": []string{"one", "two"}, "field_confirmation": []string{"one", "two", "three"}}))
}

func TestValidateSize(t *testing.T) {
	assert.True(t, validateSize("field", "123", []string{"3"}, map[string]interface{}{}))
	assert.True(t, validateSize("field", "", []string{"0"}, map[string]interface{}{}))
	assert.False(t, validateSize("field", "4567", []string{"5"}, map[string]interface{}{}))
	assert.False(t, validateSize("field", "4567", []string{"2"}, map[string]interface{}{}))

	assert.True(t, validateSize("field", "🇩🇪👍🏼", []string{"2"}, map[string]interface{}{}))
	assert.True(t, validateSize("field", "👍🏼!", []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateSize("field", "👍🏼", []string{"2"}, map[string]interface{}{}))

	assert.False(t, validateSize("field", 4567, []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateSize("field", 4567.8, []string{"2"}, map[string]interface{}{}))

	// Unsupported type
	assert.True(t, validateSize("field", true, []string{"2"}, map[string]interface{}{}))

	assert.Panics(t, func() { validateSize("field", "123", []string{"test"}, map[string]interface{}{}) })

	assert.True(t, validateSize("field", []string{"a", "b", "c"}, []string{"3"}, map[string]interface{}{}))
	assert.False(t, validateSize("field", []string{"a", "b", "c", "d"}, []string{"3"}, map[string]interface{}{}))

	assert.True(t, validateSize("field", 5, []string{"5"}, map[string]interface{}{}))
	assert.False(t, validateSize("field", 3, []string{"5"}, map[string]interface{}{}))

	assert.True(t, validateSize("file", createTestFiles(logoPath), []string{"1"}, map[string]interface{}{}))
	assert.True(t, validateSize("file", createTestFiles(largeLogoPath), []string{"42"}, map[string]interface{}{}))
	assert.False(t, validateSize("file", createTestFiles(logoPath), []string{"3"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "size"},
			},
		}
		field.check()
	})
}

func TestValidateObject(t *testing.T) {
	assert.False(t, validateObject("field", "123", []string{}, map[string]interface{}{}))
	assert.True(t, validateObject("field", map[string]interface{}{"hello": "world"}, []string{}, map[string]interface{}{}))
}
