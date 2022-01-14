package reflectutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnly(t *testing.T) {
	type Data struct {
		Field string
		Slice []float64
		Num   int
	}
	type Promote struct {
		Other string
		Data
	}
	type PromotePtr struct {
		*Data
		Other string
	}

	data := map[string]interface{}{
		"field": "value",
		"num":   42,
		"slice": []float64{2, 4, 8},
	}
	expected := map[string]interface{}{
		"field": "value",
		"slice": []float64{2, 4, 8},
	}
	res := Only(data, "field", "slice")
	assert.Equal(t, expected, res)
	assert.Equal(t, data["slice"], res["slice"])

	model := Data{
		Field: "value",
		Num:   42,
		Slice: []float64{3, 6, 9},
	}
	expected = map[string]interface{}{
		"Field": "value",
		"Slice": []float64{3, 6, 9},
	}
	res = Only(model, "Field", "Slice")
	assert.Equal(t, expected, res)
	assert.Equal(t, model.Slice, res["Slice"])

	res = Only(&model, "Field", "Slice")
	assert.Equal(t, expected, res)
	assert.Equal(t, model.Slice, res["Slice"])

	// Promoted fields
	promote := Promote{
		Data: Data{
			Field: "value",
			Num:   42,
			Slice: []float64{3, 6, 9},
		},
		Other: "test",
	}
	expected = map[string]interface{}{
		"Field": "value",
		"Slice": []float64{3, 6, 9},
		"Other": "test",
	}
	res = Only(promote, "Field", "Slice", "Other")
	assert.Equal(t, expected, res)
	assert.Equal(t, promote.Slice, res["Slice"])

	// Promoted fields ptr
	promotePtr := PromotePtr{
		Data: &Data{
			Field: "value",
			Num:   42,
			Slice: []float64{3, 6, 9},
		},
		Other: "test",
	}
	expected = map[string]interface{}{
		"Field": "value",
		"Slice": []float64{3, 6, 9},
		"Other": "test",
	}
	res = Only(promotePtr, "Field", "Slice", "Other")
	assert.Equal(t, expected, res)
	assert.Equal(t, promote.Slice, res["Slice"])

	// Promoted fields ptr nil
	promotePtr = PromotePtr{
		Other: "test",
	}
	expected = map[string]interface{}{
		"Other": "test",
	}
	res = Only(promotePtr, "Field", "Slice", "Other")
	assert.Equal(t, expected, res)
}

func TestOnlyError(t *testing.T) {
	dataInt := map[int]interface{}{
		1: "value",
		3: 42,
		4: []float64{2, 4, 8},
	}
	assert.Panics(t, func() {
		Only(dataInt, "3", "5")
	})

	assert.Panics(t, func() {
		Only("not a struct")
	})
}

func TestOnlyConflictingPromotedFields(t *testing.T) {
	type Data struct {
		Field string
	}
	type Promote struct {
		Field string
		Data
	}

	data := Promote{
		Data: Data{
			Field: "in data",
		},
		Field: "in promote",
	}
	expected := map[string]interface{}{
		"Field": "in promote",
	}
	res := Only(data, "Field")
	assert.Equal(t, expected, res)
}
