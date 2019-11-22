package validation

import (
	"reflect"

	"github.com/System-Glitch/goyave/helper"
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
		if helper.Contains(found, v) {
			return false
		}
		found = append(found, v)
	}

	return true
}

func checkInNumeric(parameters []string, value interface{}) bool {
	for _, v := range parameters {
		floatVal, _ := helper.ToFloat64(value)
		other, err := helper.ToFloat64(v)
		if err == nil && floatVal == other { // Compare only values of the same type
			return true
		}
	}
	return false
}

func validateIn(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("in", parameters, 1)
	switch getFieldType(value) {
	case "numeric":
		return checkInNumeric(parameters, value)
	case "string":
		return helper.Contains(parameters, value)
	}
	// Don't check arrays and files
	return false
}

func validateNotIn(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("not_in", parameters, 1)
	switch getFieldType(value) {
	case "numeric":
		return !checkInNumeric(parameters, value)
	case "string":
		return !helper.Contains(parameters, value)
	}
	// Don't check arrays and files
	return false
}

func validateInArray(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("in_array", parameters, 1)
	other, exists := form[parameters[0]]
	if exists && getFieldType(other) == "array" {
		return helper.Contains(other, value)
	}
	return false
}

func validateNotInArray(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("not_in_array", parameters, 1)
	other, exists := form[parameters[0]]
	if exists && getFieldType(other) == "array" {
		return !helper.Contains(other, value)
	}
	return false
}
