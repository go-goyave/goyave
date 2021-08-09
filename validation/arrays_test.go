package validation

import (
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidateArray(t *testing.T) {
	assert.True(t, validateArray(newTestContext("field", []string{"test"}, []string{}, map[string]interface{}{})))
	assert.True(t, validateArray(newTestContext("field", []int{5}, []string{}, map[string]interface{}{})))
	assert.True(t, validateArray(newTestContext("field", []float64{5.5}, []string{}, map[string]interface{}{})))
	assert.True(t, validateArray(newTestContext("field", []bool{true}, []string{}, map[string]interface{}{})))
	assert.False(t, validateArray(newTestContext("field", map[string]string{}, []string{}, map[string]interface{}{})))
	assert.False(t, validateArray(newTestContext("field", "test", []string{}, map[string]interface{}{})))
	assert.False(t, validateArray(newTestContext("field", 5, []string{}, map[string]interface{}{})))
	assert.False(t, validateArray(newTestContext("field", 5.0, []string{}, map[string]interface{}{})))
	assert.False(t, validateArray(newTestContext("field", true, []string{}, map[string]interface{}{})))

	// With type validation
	assert.Panics(t, func() {
		validateArray(newTestContext("field", []float64{5.5}, []string{"file"}, map[string]interface{}{}))
	})
	assert.Panics(t, func() {
		validateArray(newTestContext("field", []float64{5.5}, []string{"array"}, map[string]interface{}{}))
	})
	assert.Panics(t, func() {
		validateArray(newTestContext("field", []float64{5.5}, []string{"not a type"}, map[string]interface{}{}))
	})
	assert.Panics(t, func() {
		validationRules["unsupported_type"] = &RuleDefinition{IsType: true}
		defer delete(validationRules, "unsupported_type")
		validateArray(newTestContext("field", []float64{5.5}, []string{"unsupported_type"}, map[string]interface{}{}))
	})
	assert.False(t, validateArray(newTestContext("field", []string{"0.5", "not numeric"}, []string{"numeric"}, map[string]interface{}{})))

	data := map[string]interface{}{
		"field": "",
	}
	ctx := newTestContext("field", []string{"0.5", "1.42"}, []string{"numeric"}, data)
	assert.True(t, validateArray(ctx))
	arr, ok := ctx.Value.([]float64)
	if assert.True(t, ok) {
		assert.Equal(t, 0.5, arr[0])
		assert.Equal(t, 1.42, arr[1])
	}

	data = map[string]interface{}{"field": ""}
	ctx = newTestContext("field", []float64{0.5, 1.42}, []string{"numeric"}, data)
	assert.True(t, validateArray(ctx))
	arr, ok = ctx.Value.([]float64)
	if assert.True(t, ok) {
		assert.Equal(t, 0.5, arr[0])
		assert.Equal(t, 1.42, arr[1])
	}

	data = map[string]interface{}{"field": ""}
	ctx = newTestContext("field", []string{"12", "42"}, []string{"integer"}, data)
	assert.True(t, validateArray(ctx))
	arrInt, ok := ctx.Value.([]int)
	if assert.True(t, ok) {
		assert.Equal(t, 12, arrInt[0])
		assert.Equal(t, 42, arrInt[1])
	}

	data = map[string]interface{}{"field": ""}
	ctx = newTestContext("field", []string{"UTC", "America/New_York"}, []string{"timezone"}, data)
	assert.True(t, validateArray(ctx))
	arrLoc, ok := ctx.Value.([]*time.Location)
	if assert.True(t, ok) {
		assert.Equal(t, time.UTC, arrLoc[0])
	}

	data = map[string]interface{}{"field": ""}
	ctx = newTestContext("field", []string{"127.0.0.1", "::1"}, []string{"ip"}, data)
	assert.True(t, validateArray(ctx))
	arrIP, ok := ctx.Value.([]net.IP)
	if assert.True(t, ok) {
		assert.Equal(t, "127.0.0.1", arrIP[0].String())
	}

	data = map[string]interface{}{"field": ""}
	ctx = newTestContext("field", []string{"5", "{\"test\":\"string\"}"}, []string{"json"}, data)
	assert.True(t, validateArray(ctx))
	arrJSON, ok := ctx.Value.([]interface{})
	if assert.True(t, ok) {
		assert.Equal(t, 5.0, arrJSON[0])
		mp, okMap := arrJSON[1].(map[string]interface{})
		assert.True(t, okMap)
		assert.Equal(t, "string", mp["test"])
	}

	data = map[string]interface{}{"field": ""}
	ctx = newTestContext("field", []string{"http://google.com", "https://systemglitch.me"}, []string{"url"}, data)
	assert.True(t, validateArray(ctx))
	arrURL, ok := ctx.Value.([]*url.URL)
	assert.True(t, ok)
	if ok {
		assert.Equal(t, "http://google.com", arrURL[0].String())
		assert.Equal(t, "https://systemglitch.me", arrURL[1].String())
	}

	data = map[string]interface{}{"field": ""}
	ctx = newTestContext("field", []string{"fdda765f-fc57-5604-a269-52a7df8164ec"}, []string{"uuid", "5"}, data)
	assert.True(t, validateArray(ctx))
	arrUUID, ok := ctx.Value.([]uuid.UUID)
	if assert.True(t, ok) {
		assert.Equal(t, "fdda765f-fc57-5604-a269-52a7df8164ec", arrUUID[0].String())
	}

	data = map[string]interface{}{"field": ""}
	ctx = newTestContext("field", []interface{}{"yes", true, false}, []string{"bool"}, data)
	assert.True(t, validateArray(ctx))
	arrBool, ok := ctx.Value.([]bool)
	if assert.True(t, ok) {
		assert.True(t, arrBool[0])
		assert.True(t, arrBool[1])
		assert.False(t, arrBool[2])
	}

	data = map[string]interface{}{"field": ""}
	ctx = newTestContext("field", []string{"2019-12-05"}, []string{"date"}, data)
	assert.True(t, validateArray(ctx))
	arrDate, ok := ctx.Value.([]time.Time)
	if assert.True(t, ok) {
		assert.Equal(t, "2019-12-05 00:00:00 +0000 UTC", arrDate[0].String())
	}

	data = map[string]interface{}{"field": ""}
	ctx = newTestContext("field", []string{"test"}, []string{"string"}, data)
	assert.True(t, validateArray(ctx))
	arrStr, ok := ctx.Value.([]string)
	if assert.True(t, ok) {
		assert.Equal(t, "test", arrStr[0])
	}

	data = map[string]interface{}{"field": ""}
	ctx = newTestContext("field", []map[string]interface{}{{"test": "success"}}, []string{"object"}, data)
	assert.True(t, validateArray(ctx))
	arrObject, ok := ctx.Value.([]map[string]interface{})
	if assert.True(t, ok) {
		assert.Equal(t, map[string]interface{}{"test": "success"}, arrObject[0])
	}
}

func TestValidateArrayInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"array": []string{"0.5", "1.42"},
		},
	}
	ctx := newTestContext("object.array", []string{"0.5", "1.42"}, []string{"numeric"}, data)
	assert.True(t, validateArray(ctx))
	arr, ok := ctx.Value.([]float64)
	if assert.True(t, ok) {
		assert.Equal(t, 0.5, arr[0])
		assert.Equal(t, 1.42, arr[1])
	}
}

func TestValidateDistinct(t *testing.T) {
	assert.True(t, validateDistinct(newTestContext("field", []string{"test", "test2", "test3"}, []string{}, map[string]interface{}{})))
	assert.True(t, validateDistinct(newTestContext("field", []int{1, 2, 3}, []string{}, map[string]interface{}{})))
	assert.True(t, validateDistinct(newTestContext("field", []float64{1.2, 4.3, 2.4, 3.5, 4.5, 4.30001}, []string{}, map[string]interface{}{})))
	assert.True(t, validateDistinct(newTestContext("field", []bool{true, false}, []string{}, map[string]interface{}{})))

	assert.False(t, validateDistinct(newTestContext("field", []string{"test", "test2", "test3", "test2"}, []string{}, map[string]interface{}{})))
	assert.False(t, validateDistinct(newTestContext("field", []int{1, 4, 2, 3, 4}, []string{}, map[string]interface{}{})))
	assert.False(t, validateDistinct(newTestContext("field", []float64{1.2, 4.3, 2.4, 3.5, 4.5, 4.30001, 4.3}, []string{}, map[string]interface{}{})))

	// Not array
	assert.False(t, validateDistinct(newTestContext("field", 8, []string{}, map[string]interface{}{})))
	assert.False(t, validateDistinct(newTestContext("field", 8.0, []string{}, map[string]interface{}{})))
	assert.False(t, validateDistinct(newTestContext("field", "string", []string{}, map[string]interface{}{})))
}

func TestValidateIn(t *testing.T) {
	assert.True(t, validateIn(newTestContext("field", "dolor", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{})))
	assert.False(t, validateIn(newTestContext("field", "dolors", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{})))
	assert.False(t, validateIn(newTestContext("field", "hello world", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{})))

	assert.True(t, validateIn(newTestContext("field", 2.5, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{})))
	assert.False(t, validateIn(newTestContext("field", 2.51, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{})))

	assert.False(t, validateIn(newTestContext("field", []string{"1"}, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "in"},
			},
		}
		field.Check()
	})
}

func TestValidateNotIn(t *testing.T) {
	assert.False(t, validateNotIn(newTestContext("field", "dolor", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{})))
	assert.True(t, validateNotIn(newTestContext("field", "dolors", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{})))
	assert.True(t, validateNotIn(newTestContext("field", "hello world", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{})))

	assert.False(t, validateNotIn(newTestContext("field", 2.5, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{})))
	assert.True(t, validateNotIn(newTestContext("field", 2.51, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{})))

	assert.False(t, validateNotIn(newTestContext("field", []string{"1"}, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "not_in"},
			},
		}
		field.Check()
	})
}

func TestValidateInArray(t *testing.T) {
	assert.True(t, validateInArray(newTestContext("field", "dolor", []string{"other"}, map[string]interface{}{"field": "dolors", "other": []string{"lorem", "ipsum", "sit", "dolor", "amet"}})))
	assert.True(t, validateInArray(newTestContext("field", 4, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []int{1, 2, 3, 4, 5}})))
	assert.True(t, validateInArray(newTestContext("field", 2.2, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []float64{1.1, 2.2, 3.3, 4.4, 5.5}})))
	assert.True(t, validateInArray(newTestContext("field", false, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []bool{true, false}})))

	assert.False(t, validateInArray(newTestContext("field", "dolors", []string{"other"}, map[string]interface{}{"field": "dolors", "other": []string{"lorem", "ipsum", "sit", "dolor", "amet"}})))
	assert.False(t, validateInArray(newTestContext("field", 1, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []string{"lorem", "ipsum", "sit", "dolor", "amet"}})))
	assert.False(t, validateInArray(newTestContext("field", 6, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []int{1, 2, 3, 4, 5}})))
	assert.False(t, validateInArray(newTestContext("field", 2.3, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []float64{1.1, 2.2, 3.3, 4.4, 5.5}})))
	assert.False(t, validateInArray(newTestContext("field", false, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []bool{true}})))
	assert.False(t, validateInArray(newTestContext("field", []string{"test"}, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []bool{true}})))
	assert.False(t, validateInArray(newTestContext("field", []string{"test"}, []string{"other"}, map[string]interface{}{"field": "dolors", "other": 1})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "in_array"},
			},
		}
		field.Check()
	})

	// Objects
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"field": "dolors",
			"other": []string{"lorem", "ipsum", "sit", "dolor", "amet"},
		},
	}
	assert.True(t, validateInArray(newTestContext("object.field", "dolor", []string{"object.other"}, data)))
	assert.False(t, validateInArray(newTestContext("object.field", "dolors", []string{"object.other"}, data)))
}

func TestValidateNotInArray(t *testing.T) {
	assert.False(t, validateNotInArray(newTestContext("field", "dolor", []string{"other"}, map[string]interface{}{"field": "dolors", "other": []string{"lorem", "ipsum", "sit", "dolor", "amet"}})))
	assert.False(t, validateNotInArray(newTestContext("field", 4, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []int{1, 2, 3, 4, 5}})))
	assert.False(t, validateNotInArray(newTestContext("field", 2.2, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []float64{1.1, 2.2, 3.3, 4.4, 5.5}})))
	assert.False(t, validateNotInArray(newTestContext("field", false, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []bool{true, false}})))
	assert.False(t, validateNotInArray(newTestContext("field", []string{"test"}, []string{"other"}, map[string]interface{}{"field": "dolors", "other": 1})))

	assert.True(t, validateNotInArray(newTestContext("field", "dolors", []string{"other"}, map[string]interface{}{"field": "dolors", "other": []string{"lorem", "ipsum", "sit", "dolor", "amet"}})))
	assert.True(t, validateNotInArray(newTestContext("field", 1, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []string{"lorem", "ipsum", "sit", "dolor", "amet"}})))
	assert.True(t, validateNotInArray(newTestContext("field", 6, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []int{1, 2, 3, 4, 5}})))
	assert.True(t, validateNotInArray(newTestContext("field", 2.3, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []float64{1.1, 2.2, 3.3, 4.4, 5.5}})))
	assert.True(t, validateNotInArray(newTestContext("field", false, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []bool{true}})))
	assert.True(t, validateNotInArray(newTestContext("field", []string{"test"}, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []bool{true}})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "not_in_array"},
			},
		}
		field.Check()
	})

	// Objects
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"field": "dolors",
			"other": []string{"lorem", "ipsum", "sit", "dolor", "amet"},
		},
	}
	assert.False(t, validateNotInArray(newTestContext("object.field", "dolor", []string{"object.other"}, data)))
	assert.True(t, validateNotInArray(newTestContext("object.field", "dolors", []string{"object.other"}, data)))
}
