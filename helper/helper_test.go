package helper

import (
	"encoding/json"
	"fmt"
	"math"
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

func TestIndexOf(t *testing.T) {
	assert.Equal(t, 1, IndexOf([]string{"val1", "val2", "val3"}, "val2"))
	assert.Equal(t, -1, IndexOf([]string{"val1", "val2", "val3"}, "val4"))
	assert.Equal(t, 1, IndexOf([]int{1, 2, 3}, 2))
	assert.Equal(t, -1, IndexOf([]int{1, 2, 3}, 4))
	assert.Equal(t, 1, IndexOf([]float64{1, 2, 3}, float64(2)))
	assert.Equal(t, -1, IndexOf([]float64{1, 2, 3}, 2))
	assert.Equal(t, -1, IndexOf([]float64{1, 2, 3}, 4))
}

func TestContainsStr(t *testing.T) {
	assert.True(t, ContainsStr([]string{"val1", "val2", "val3"}, "val2"))
	assert.False(t, ContainsStr([]string{"val1", "val2", "val3"}, "val4"))
}

func TestIndexOfStr(t *testing.T) {
	assert.Equal(t, 1, IndexOfStr([]string{"val1", "val2", "val3"}, "val2"))
	assert.Equal(t, -1, IndexOfStr([]string{"val1", "val2", "val3"}, "val4"))
}

func TestToFloat64(t *testing.T) {
	v, err := ToFloat64(1)
	assert.Nil(t, err)
	assert.Equal(t, 1.0, v)

	v, err = ToFloat64("1")
	assert.Nil(t, err)
	assert.Equal(t, 1.0, v)

	v, err = ToFloat64(uint(5))
	assert.Nil(t, err)
	assert.Equal(t, 5.0, v)

	v, err = ToFloat64("1.58")
	assert.Nil(t, err)
	assert.Equal(t, 1.58, v)

	v, err = ToFloat64(1.5)
	assert.Nil(t, err)
	assert.Equal(t, 1.5, v)

	v, err = ToFloat64("NaN")
	assert.Nil(t, err)
	assert.True(t, math.IsNaN(v))

	v, err = ToFloat64([]string{})
	fmt.Println(err)
	assert.NotNil(t, err)
	assert.Equal(t, float64(0), v)
}

func TestToString(t *testing.T) {
	assert.Equal(t, "12", ToString(12))
	assert.Equal(t, "-12", ToString(-12))
	assert.Equal(t, "12.5", ToString(12.5))
	assert.Equal(t, "-12.5", ToString(-12.5))
	assert.Equal(t, "true", ToString(true))
	assert.Equal(t, "[test]", ToString([]string{"test"}))
}

func TestSliceEqual(t *testing.T) {
	assert.True(t, SliceEqual([]string{"one", "two", "three"}, []string{"one", "two", "three"}))
	assert.True(t, SliceEqual([]int{1, 2, 3}, []int{1, 2, 3}))
	assert.True(t, SliceEqual([]float64{1.1, 2.2, 3.3}, []float64{1.1, 2.2, 3.3}))

	assert.False(t, SliceEqual([]string{"one", "two", "three"}, []string{"one", "three", "two"}))
	assert.False(t, SliceEqual([]string{"one", "two", "three"}, []string{"one"}))
	assert.False(t, SliceEqual([]string{"one", "two", "three"}, []int{1, 2, 3}))
}

func TestParseMultiValuesHeader(t *testing.T) {
	expected := []HeaderValue{
		{Value: "text/html", Priority: 0.8},
		{Value: "text/*", Priority: 0.8},
		{Value: "*/*", Priority: 0.8},
	}
	result := ParseMultiValuesHeader("text/html;q=0.8,text/*;q=0.8,*/*;q=0.8")
	assert.True(t, SliceEqual(expected, result))

	result = ParseMultiValuesHeader("*/*;q=0.8,text/*;q=0.8,text/html;q=0.8")
	assert.True(t, SliceEqual(expected, result))

	expected = []HeaderValue{
		{Value: "text/html", Priority: 1},
		{Value: "*/*", Priority: 0.7},
		{Value: "text/*", Priority: 0.5},
	}
	result = ParseMultiValuesHeader("text/html,text/*;q=0.5,*/*;q=0.7")
	assert.True(t, SliceEqual(expected, result))

	expected = []HeaderValue{
		{Value: "fr", Priority: 1},
		{Value: "fr-FR", Priority: 0.8},
		{Value: "en-US", Priority: 0.5},
		{Value: "en-*", Priority: 0.3},
		{Value: "en", Priority: 0.3},
		{Value: "*", Priority: 0.3},
	}
	result = ParseMultiValuesHeader("fr , fr-FR;q=0.8, en-US ;q=0.5, *;q=0.3, en-*;q=0.3, en;q=0.3")
	assert.True(t, SliceEqual(expected, result))

	expected = []HeaderValue{{Value: "fr", Priority: 1}}
	result = ParseMultiValuesHeader("fr")
	assert.True(t, SliceEqual(expected, result))

	expected = []HeaderValue{{Value: "fr", Priority: 0.3}}
	result = ParseMultiValuesHeader("fr;q=0.3")
	assert.True(t, SliceEqual(expected, result))
}

func TestRemoveHiddenFields(t *testing.T) {
	type Model struct {
		Password string `model:"hide" json:",omitempty"`
		Username string
	}

	model := &Model{
		Password: "bcrypted password",
		Username: "Jeff",
	}

	RemoveHiddenFields(model)
	assert.Empty(t, model.Password)
	assert.Equal(t, "Jeff", model.Username)

	json, _ := json.Marshal(model)
	assert.Equal(t, "{\"Username\":\"Jeff\"}", string(json))
}

func TestRemoveHiddenPromotedFields(t *testing.T) {
	type PasswordHolder struct {
		Password     string `model:"hide" json:",omitempty"`
		privateField string `model:"hide"`
	}

	type Model struct {
		PasswordHolder
		Username string
	}

	model := &Model{
		PasswordHolder: PasswordHolder{
			Password:     "bcrypted password",
			privateField: "this is private",
		},
		Username: "Jeff",
	}

	RemoveHiddenFields(model)
	assert.Empty(t, model.Password)
	assert.Equal(t, "Jeff", model.Username)
	assert.Equal(t, "this is private", model.privateField)

	json, _ := json.Marshal(model)
	assert.Equal(t, "{\"Username\":\"Jeff\"}", string(json))
}

func TestRemoveHiddenFieldsNotStruct(t *testing.T) {
	assert.NotPanics(t, func() {
		RemoveHiddenFields(map[string]string{})
	})
	assert.NotPanics(t, func() {
		RemoveHiddenFields("test")
	})
	assert.NotPanics(t, func() {
		RemoveHiddenFields([]string{})
	})
}

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

func TestEscapeLike(t *testing.T) {
	assert.Equal(t, "se\\%r\\_h", EscapeLike("se%r_h"))
	assert.Equal(t, "se\\%r\\%\\_h\\_", EscapeLike("se%r%_h_"))
}
