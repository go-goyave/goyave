package validation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createDate(date string) time.Time {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	return t
}

func createDateTime(date string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05", date)
	if err != nil {
		panic(err)
	}
	return t
}

func TestValidateDate(t *testing.T) {
	data := map[string]interface{}{
		"field": "",
	}
	assert.True(t, validateDate("field", "2019-11-02", []string{}, data))
	assert.False(t, validateDate("field", "2019-13-02", []string{}, data))
	assert.False(t, validateDate("field", "2019-12-32", []string{}, data))

	assert.True(t, validateDate("field", "2019-11-02 11:07:42", []string{"2006-01-02 03:04:05"}, data))
	assert.False(t, validateDate("field", "2019-11-02 24:07:42", []string{"2006-01-02 03:04:05"}, data))
	assert.False(t, validateDate("field", "2019-11-02 11:61:42", []string{"2006-01-02 03:04:05"}, data))
	assert.False(t, validateDate("field", "2019-11-02 11:61:61", []string{"2006-01-02 03:04:05"}, data))
	assert.False(t, validateDate("field", "hello", []string{}, data))
	assert.False(t, validateDate("field", 1, []string{"2006-01-02 03:04:05"}, data))
	assert.False(t, validateDate("field", 1.0, []string{"2006-01-02 03:04:05"}, data))
	assert.False(t, validateDate("field", true, []string{"2006-01-02 03:04:05"}, data))
	assert.False(t, validateDate("field", []string{}, []string{"2006-01-02 03:04:05"}, data))
}

func TestValidateBefore(t *testing.T) {
	assert.True(t, validateBefore("field", createDate("2019-11-02"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.True(t, validateBefore("field", createDateTime("2019-11-02T11:00:00"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateBefore("field", createDate("2019-11-03"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateBefore("field", createDateTime("2019-11-02T12:00:00"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateBefore("field", createDateTime("2019-11-02T13:00:00"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))

	assert.False(t, validateBefore("field", "hello", []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateBefore("field", 1.0, []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		validateBefore("field", createDate("2019-11-02"), []string{"invalid date and field doesn't exist"}, map[string]interface{}{})
	})

	assert.True(t, validateBefore("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": "2019-11-03"}))
	assert.True(t, validateBefore("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": createDate("2019-11-03")}))
	assert.False(t, validateBefore("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": createDate("2019-11-02")}))
	assert.False(t, validateBefore("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": "hello"}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "before"},
			},
		}
		field.check()
	})

	// Objects
	data := map[string]interface{}{
		"field": createDate("2019-11-02"),
		"object": map[string]interface{}{
			"other": createDate("2019-11-03"),
		},
	}
	assert.True(t, validateBefore("field", data["field"], []string{"object.other"}, data))
}

func TestValidateBeforeEqual(t *testing.T) {
	assert.True(t, validateBeforeEqual("field", createDate("2019-11-02"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.True(t, validateBeforeEqual("field", createDateTime("2019-11-02T11:00:00"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.True(t, validateBeforeEqual("field", createDateTime("2019-11-02T12:00:00"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateBeforeEqual("field", createDate("2019-11-03"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateBeforeEqual("field", createDateTime("2019-11-02T13:00:00"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))

	assert.False(t, validateBeforeEqual("field", "hello", []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateBeforeEqual("field", 1.0, []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		validateBeforeEqual("field", createDate("2019-11-02"), []string{"invalid date and field doesn't exist"}, map[string]interface{}{})
	})

	assert.True(t, validateBeforeEqual("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": "2019-11-03"}))
	assert.True(t, validateBeforeEqual("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": createDate("2019-11-03")}))
	assert.True(t, validateBeforeEqual("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": createDate("2019-11-02")}))
	assert.False(t, validateBeforeEqual("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": "hello"}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "before_equal"},
			},
		}
		field.check()
	})

	// Objects
	data := map[string]interface{}{
		"field": createDate("2019-11-02"),
		"object": map[string]interface{}{
			"other": createDate("2019-11-03"),
		},
	}
	assert.True(t, validateBeforeEqual("field", data["field"], []string{"object.other"}, data))

	data = map[string]interface{}{
		"field": createDate("2019-11-02"),
		"object": map[string]interface{}{
			"other": createDate("2019-11-02"),
		},
	}
	assert.True(t, validateBeforeEqual("field", data["field"], []string{"object.other"}, data))
}

func TestValidateAfter(t *testing.T) {
	assert.False(t, validateAfter("field", createDate("2019-11-02"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateAfter("field", createDateTime("2019-11-02T11:00:00"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateAfter("field", createDateTime("2019-11-02T12:00:00"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.True(t, validateAfter("field", createDate("2019-11-03"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.True(t, validateAfter("field", createDateTime("2019-11-02T13:00:00"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))

	assert.False(t, validateAfter("field", "hello", []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateAfter("field", 1.0, []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		validateAfter("field", createDate("2019-11-02"), []string{"invalid date and field doesn't exist"}, map[string]interface{}{})
	})

	assert.False(t, validateAfter("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": "2019-11-03"}))
	assert.True(t, validateAfter("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": "2019-11-01"}))
	assert.True(t, validateAfter("field", createDate("2019-11-04"), []string{"other"}, map[string]interface{}{"other": createDate("2019-11-03")}))
	assert.False(t, validateAfter("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": createDate("2019-11-02")}))
	assert.False(t, validateAfter("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": "hello"}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "after"},
			},
		}
		field.check()
	})

	// Objects
	data := map[string]interface{}{
		"field": createDate("2019-11-03"),
		"object": map[string]interface{}{
			"other": createDate("2019-11-02"),
		},
	}
	assert.True(t, validateAfter("field", data["field"], []string{"object.other"}, data))
}

func TestValidateAfterEqual(t *testing.T) {
	assert.False(t, validateAfterEqual("field", createDate("2019-11-02"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateAfterEqual("field", createDateTime("2019-11-02T11:00:00"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.True(t, validateAfterEqual("field", createDateTime("2019-11-02T12:00:00"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.True(t, validateAfterEqual("field", createDate("2019-11-03"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.True(t, validateAfterEqual("field", createDateTime("2019-11-02T13:00:00"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))

	assert.False(t, validateAfterEqual("field", "hello", []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateAfterEqual("field", 1.0, []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		validateAfterEqual("field", createDate("2019-11-02"), []string{"invalid date and field doesn't exist"}, map[string]interface{}{})
	})

	assert.False(t, validateAfterEqual("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": "2019-11-03"}))
	assert.True(t, validateAfterEqual("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": "2019-11-01"}))
	assert.False(t, validateAfterEqual("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": createDate("2019-11-03")}))
	assert.True(t, validateAfterEqual("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": createDate("2019-11-02")}))
	assert.False(t, validateAfterEqual("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": "hello"}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "after_equal"},
			},
		}
		field.check()
	})

	// Objects
	data := map[string]interface{}{
		"field": createDate("2019-11-03"),
		"object": map[string]interface{}{
			"other": createDate("2019-11-02"),
		},
	}
	assert.True(t, validateAfterEqual("field", data["field"], []string{"object.other"}, data))

	data = map[string]interface{}{
		"field": createDate("2019-11-02"),
		"object": map[string]interface{}{
			"other": createDate("2019-11-02"),
		},
	}
	assert.True(t, validateAfterEqual("field", data["field"], []string{"object.other"}, data))
}

func TestValidateDateEquals(t *testing.T) {
	assert.True(t, validateDateEquals("field", createDate("2019-11-02"), []string{"2019-11-02T00:00:00"}, map[string]interface{}{}))
	assert.False(t, validateDateEquals("field", createDate("2019-11-02"), []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))

	assert.False(t, validateDateEquals("field", "hello", []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))
	assert.False(t, validateDateEquals("field", 1.0, []string{"2019-11-02T12:00:00"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		validateDateEquals("field", createDate("2019-11-02"), []string{"invalid date and field doesn't exist"}, map[string]interface{}{})
	})

	assert.True(t, validateDateEquals("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": "2019-11-02"}))
	assert.True(t, validateDateEquals("field", createDateTime("2019-11-02T13:14:15"), []string{"other"}, map[string]interface{}{"other": createDateTime("2019-11-02T13:14:15")}))
	assert.False(t, validateDateEquals("field", createDate("2019-11-03"), []string{"other"}, map[string]interface{}{"other": createDateTime("2019-11-02T13:14:16")}))
	assert.False(t, validateDateEquals("field", createDateTime("2019-11-02T13:14:15"), []string{"other"}, map[string]interface{}{"other": createDateTime("2019-11-02T13:14:16")}))
	assert.False(t, validateDateEquals("field", createDate("2019-11-02"), []string{"other"}, map[string]interface{}{"other": "hello"}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "date_equals"},
			},
		}
		field.check()
	})

	// Objects
	data := map[string]interface{}{
		"field": createDate("2019-11-02"),
		"object": map[string]interface{}{
			"other": createDate("2019-11-02"),
		},
	}
	assert.True(t, validateDateEquals("field", data["field"], []string{"object.other"}, data))
}

func TestValidateDateBetween(t *testing.T) {
	assert.True(t, validateDateBetween("field", createDate("2019-11-02"), []string{"2019-11-01T00:00:00", "2019-11-03T00:00:00"}, map[string]interface{}{}))
	assert.True(t, validateDateBetween("field", createDate("2019-11-02"), []string{"2019-11-02T00:00:00", "2019-11-03T00:00:00"}, map[string]interface{}{}))
	assert.False(t, validateDateBetween("field", createDate("2019-11-04"), []string{"2019-11-02T00:00:00", "2019-11-03T00:00:00"}, map[string]interface{}{}))
	assert.False(t, validateDateBetween("field", createDate("2019-11-01"), []string{"2019-11-02T00:00:00", "2019-11-03T00:00:00"}, map[string]interface{}{}))

	assert.True(t, validateDateBetween("field", createDateTime("2019-11-02T13:14:15"), []string{"min", "max"}, map[string]interface{}{"min": createDateTime("2019-11-02T13:14:00"), "max": createDateTime("2019-11-02T14:14:00")}))
	assert.True(t, validateDateBetween("field", createDateTime("2019-11-02T13:14:15"), []string{"min", "2019-11-03T00:00:00"}, map[string]interface{}{"min": createDateTime("2019-11-02T13:14:00")}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "date_between"},
			},
		}
		field.check()
	})

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "date_between", Params: []string{"2019-11-03T00:00:00"}},
			},
		}
		field.check()
	})

	// Objects
	data := map[string]interface{}{
		"field": createDate("2019-11-03"),
		"object": map[string]interface{}{
			"min": createDate("2019-11-02"),
			"max": createDate("2019-11-04"),
		},
	}
	assert.True(t, validateDateBetween("field", data["field"], []string{"object.min", "object.max"}, data))
}

func TestValidateDateConvert(t *testing.T) {
	form := map[string]interface{}{"field": "2019-11-02"}
	assert.True(t, validateDate("field", form["field"], []string{}, form))

	_, ok := form["field"].(time.Time)
	assert.True(t, ok)
}

func TestValidateDateConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"time": "2019-11-02",
		},
	}

	set := RuleSet{
		"object":      {"required", "object"},
		"object.time": {"required", "date"},
	}

	errors := Validate(data, set, true, "en-US")
	assert.Empty(t, errors)
	_, ok := data["object"].(map[string]interface{})["time"].(time.Time)
	assert.True(t, ok)
}
