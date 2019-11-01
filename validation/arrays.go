package validation

import (
	"reflect"

	"github.com/System-Glitch/goyave/helpers"
)

func validateArray(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	return getFieldType(value) == "array"
}

func validateDistinct(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	if getFieldType(value) != "array" {
		return false // Can't validate if not an array
	}

	found := []interface{}{}
	list := reflect.ValueOf(value)
	for i := 0; i < list.Len(); i++ {
		v := list.Index(i).Interface()
		if helpers.Contains(found, v) {
			return false
		}
		found = append(found, v)
	}

	return true
}

func checkInNumeric(parameters []string, value interface{}) bool {
	for _, v := range parameters {
		floatVal, _ := helpers.ToFloat64(value)
		other, err := helpers.ToFloat64(v)
		if err == nil && floatVal == other { // Compare only values of the same type
			return true
		}
	}
	return false
}

func validateIn(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	requireParametersCount("in", parameters, 1)
	switch getFieldType(value) {
	case "numeric":
		return checkInNumeric(parameters, value)
	case "string":
		return helpers.Contains(parameters, value)
	}
	// Don't check arrays and files
	return false
}

func validateNotIn(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	requireParametersCount("in", parameters, 1)
	switch getFieldType(value) {
	case "numeric":
		return !checkInNumeric(parameters, value)
	case "string":
		return !helpers.Contains(parameters, value)
	}
	// Don't check arrays and files
	return false
}

// in_array
// not_in_array
