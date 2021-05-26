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
	assert.True(t, validateArray("field", []string{"test"}, []string{}, map[string]interface{}{}))
	assert.True(t, validateArray("field", []int{5}, []string{}, map[string]interface{}{}))
	assert.True(t, validateArray("field", []float64{5.5}, []string{}, map[string]interface{}{}))
	assert.True(t, validateArray("field", []bool{true}, []string{}, map[string]interface{}{}))
	assert.False(t, validateArray("field", map[string]string{}, []string{}, map[string]interface{}{}))
	assert.False(t, validateArray("field", "test", []string{}, map[string]interface{}{}))
	assert.False(t, validateArray("field", 5, []string{}, map[string]interface{}{}))
	assert.False(t, validateArray("field", 5.0, []string{}, map[string]interface{}{}))
	assert.False(t, validateArray("field", true, []string{}, map[string]interface{}{}))

	// With type validation
	assert.Panics(t, func() {
		validateArray("field", []float64{5.5}, []string{"file"}, map[string]interface{}{})
	})
	assert.Panics(t, func() {
		validateArray("field", []float64{5.5}, []string{"array"}, map[string]interface{}{})
	})
	assert.Panics(t, func() {
		validateArray("field", []float64{5.5}, []string{"not a type"}, map[string]interface{}{})
	})
	assert.False(t, validateArray("field", []string{"0.5", "not numeric"}, []string{"numeric"}, map[string]interface{}{}))

	data := map[string]interface{}{
		"field": "",
	}
	assert.True(t, validateArray("field", []string{"0.5", "1.42"}, []string{"numeric"}, data))
	arr, ok := data["field"].([]float64)
	assert.True(t, ok)
	if ok {
		assert.Equal(t, 0.5, arr[0])
		assert.Equal(t, 1.42, arr[1])
	}

	data = map[string]interface{}{"field": ""}
	assert.True(t, validateArray("field", []float64{0.5, 1.42}, []string{"numeric"}, data))
	arr, ok = data["field"].([]float64)
	assert.True(t, ok)
	if ok {
		assert.Equal(t, 0.5, arr[0])
		assert.Equal(t, 1.42, arr[1])
	}

	data = map[string]interface{}{"field": ""}
	assert.True(t, validateArray("field", []string{"12", "42"}, []string{"integer"}, data))
	arrInt, ok := data["field"].([]int)
	assert.True(t, ok)
	if ok {
		assert.Equal(t, 12, arrInt[0])
		assert.Equal(t, 42, arrInt[1])
	}

	data = map[string]interface{}{"field": ""}
	assert.True(t, validateArray("field", []string{"UTC", "America/New_York"}, []string{"timezone"}, data))
	arrLoc, ok := data["field"].([]*time.Location)
	assert.True(t, ok)
	if ok {
		assert.Equal(t, time.UTC, arrLoc[0])
	}

	data = map[string]interface{}{"field": ""}
	assert.True(t, validateArray("field", []string{"127.0.0.1", "::1"}, []string{"ip"}, data))
	arrIP, ok := data["field"].([]net.IP)
	assert.True(t, ok)
	if ok {
		assert.Equal(t, "127.0.0.1", arrIP[0].String())
	}

	data = map[string]interface{}{"field": ""}
	assert.True(t, validateArray("field", []string{"5", "{\"test\":\"string\"}"}, []string{"json"}, data))
	arrJSON, ok := data["field"].([]interface{})
	assert.True(t, ok)
	if ok {
		assert.Equal(t, 5.0, arrJSON[0])
		mp, okMap := arrJSON[1].(map[string]interface{})
		assert.True(t, okMap)
		assert.Equal(t, "string", mp["test"])
	}

	data = map[string]interface{}{"field": ""}
	assert.True(t, validateArray("field", []string{"http://google.com", "https://systemglitch.me"}, []string{"url"}, data))
	arrURL, ok := data["field"].([]*url.URL)
	assert.True(t, ok)
	if ok {
		assert.Equal(t, "http://google.com", arrURL[0].String())
		assert.Equal(t, "https://systemglitch.me", arrURL[1].String())
	}

	data = map[string]interface{}{"field": ""}
	assert.True(t, validateArray("field", []string{"fdda765f-fc57-5604-a269-52a7df8164ec"}, []string{"uuid", "5"}, data))
	arrUUID, ok := data["field"].([]uuid.UUID)
	assert.True(t, ok)
	if ok {
		assert.Equal(t, "fdda765f-fc57-5604-a269-52a7df8164ec", arrUUID[0].String())
	}

	data = map[string]interface{}{"field": ""}
	assert.True(t, validateArray("field", []interface{}{"yes", true, false}, []string{"bool"}, data))
	arrBool, ok := data["field"].([]bool)
	assert.True(t, ok)
	if ok {
		assert.True(t, arrBool[0])
		assert.True(t, arrBool[1])
		assert.False(t, arrBool[2])
	}

	data = map[string]interface{}{"field": ""}
	assert.True(t, validateArray("field", []string{"2019-12-05"}, []string{"date"}, data))
	arrDate, ok := data["field"].([]time.Time)
	assert.True(t, ok)
	if ok {
		assert.Equal(t, "2019-12-05 00:00:00 +0000 UTC", arrDate[0].String())
	}

	data = map[string]interface{}{"field": ""}
	assert.True(t, validateArray("field", []string{"test"}, []string{"string"}, data))
	arrStr, ok := data["field"].([]string)
	assert.True(t, ok)
	if ok {
		assert.Equal(t, "test", arrStr[0])
	}
}

func TestValidateArrayInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"array": []string{"0.5", "1.42"},
		},
	}
	assert.True(t, validateArray("object.array", []string{"0.5", "1.42"}, []string{"numeric"}, data))
	arr, ok := data["object"].(map[string]interface{})["array"].([]float64)
	assert.True(t, ok)
	if ok {
		assert.Equal(t, 0.5, arr[0])
		assert.Equal(t, 1.42, arr[1])
	}
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

func TestValidateIn(t *testing.T) {
	assert.True(t, validateIn("field", "dolor", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{}))
	assert.False(t, validateIn("field", "dolors", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{}))
	assert.False(t, validateIn("field", "hello world", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{}))

	assert.True(t, validateIn("field", 2.5, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{}))
	assert.False(t, validateIn("field", 2.51, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{}))

	assert.False(t, validateIn("field", []string{"1"}, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "in"},
			},
		}
		field.check()
	})
}

func TestValidateNotIn(t *testing.T) {
	assert.False(t, validateNotIn("field", "dolor", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{}))
	assert.True(t, validateNotIn("field", "dolors", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{}))
	assert.True(t, validateNotIn("field", "hello world", []string{"lorem", "ipsum", "sit", "dolor", "amet"}, map[string]interface{}{}))

	assert.False(t, validateNotIn("field", 2.5, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{}))
	assert.True(t, validateNotIn("field", 2.51, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{}))

	assert.False(t, validateNotIn("field", []string{"1"}, []string{"1", "2.4", "2.65", "87", "2.5"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "not_in"},
			},
		}
		field.check()
	})
}

func TestValidateInArray(t *testing.T) {
	assert.True(t, validateInArray("field", "dolor", []string{"other"}, map[string]interface{}{"field": "dolors", "other": []string{"lorem", "ipsum", "sit", "dolor", "amet"}}))
	assert.True(t, validateInArray("field", 4, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []int{1, 2, 3, 4, 5}}))
	assert.True(t, validateInArray("field", 2.2, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []float64{1.1, 2.2, 3.3, 4.4, 5.5}}))
	assert.True(t, validateInArray("field", false, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []bool{true, false}}))

	assert.False(t, validateInArray("field", "dolors", []string{"other"}, map[string]interface{}{"field": "dolors", "other": []string{"lorem", "ipsum", "sit", "dolor", "amet"}}))
	assert.False(t, validateInArray("field", 1, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []string{"lorem", "ipsum", "sit", "dolor", "amet"}}))
	assert.False(t, validateInArray("field", 6, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []int{1, 2, 3, 4, 5}}))
	assert.False(t, validateInArray("field", 2.3, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []float64{1.1, 2.2, 3.3, 4.4, 5.5}}))
	assert.False(t, validateInArray("field", false, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []bool{true}}))
	assert.False(t, validateInArray("field", []string{"test"}, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []bool{true}}))
	assert.False(t, validateInArray("field", []string{"test"}, []string{"other"}, map[string]interface{}{"field": "dolors", "other": 1}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "in_array"},
			},
		}
		field.check()
	})

	// Objects
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"field": "dolors",
			"other": []string{"lorem", "ipsum", "sit", "dolor", "amet"},
		},
	}
	assert.True(t, validateInArray("object.field", "dolor", []string{"object.other"}, data))
	assert.False(t, validateInArray("object.field", "dolors", []string{"object.other"}, data))
}

func TestValidateNotInArray(t *testing.T) {
	assert.False(t, validateNotInArray("field", "dolor", []string{"other"}, map[string]interface{}{"field": "dolors", "other": []string{"lorem", "ipsum", "sit", "dolor", "amet"}}))
	assert.False(t, validateNotInArray("field", 4, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []int{1, 2, 3, 4, 5}}))
	assert.False(t, validateNotInArray("field", 2.2, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []float64{1.1, 2.2, 3.3, 4.4, 5.5}}))
	assert.False(t, validateNotInArray("field", false, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []bool{true, false}}))
	assert.False(t, validateNotInArray("field", []string{"test"}, []string{"other"}, map[string]interface{}{"field": "dolors", "other": 1}))

	assert.True(t, validateNotInArray("field", "dolors", []string{"other"}, map[string]interface{}{"field": "dolors", "other": []string{"lorem", "ipsum", "sit", "dolor", "amet"}}))
	assert.True(t, validateNotInArray("field", 1, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []string{"lorem", "ipsum", "sit", "dolor", "amet"}}))
	assert.True(t, validateNotInArray("field", 6, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []int{1, 2, 3, 4, 5}}))
	assert.True(t, validateNotInArray("field", 2.3, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []float64{1.1, 2.2, 3.3, 4.4, 5.5}}))
	assert.True(t, validateNotInArray("field", false, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []bool{true}}))
	assert.True(t, validateNotInArray("field", []string{"test"}, []string{"other"}, map[string]interface{}{"field": "dolors", "other": []bool{true}}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "not_in_array"},
			},
		}
		field.check()
	})

	// Objects
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"field": "dolors",
			"other": []string{"lorem", "ipsum", "sit", "dolor", "amet"},
		},
	}
	assert.False(t, validateNotInArray("object.field", "dolor", []string{"object.other"}, data))
	assert.True(t, validateNotInArray("object.field", "dolors", []string{"object.other"}, data))
}
