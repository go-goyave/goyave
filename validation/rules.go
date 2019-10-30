package validation

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/System-Glitch/goyave/helpers"
)

// Rule function defining a validation rule.
// Passing rules should return true, false otherwise.
//
// Rules can modifiy the validated value if needed.
// For example, the "numeric" rule converts the data to float64 if it's a string.
type Rule func(string, interface{}, []string, map[string]interface{}) bool

func validateRequired(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	val, ok := form[field]
	if ok {
		if val == nil {
			return false
		}
		if str, okStr := val.(string); okStr && str == "" {
			return false
		}
	}
	return ok
}

// -------------------------
// Numeric

func validateNumeric(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	rv := reflect.ValueOf(value)
	kind := rv.Kind().String()
	switch {
	case strings.HasPrefix(kind, "int"), strings.HasPrefix(kind, "uint") && kind != "uintptr", strings.HasPrefix(kind, "float"):
		return true
	case kind == "string":
		floatVal, err := strconv.ParseFloat(value.(string), 64)
		if err == nil {
			form[field] = floatVal
		}
		return err == nil
	default:
		return false
	}
}

func validateInteger(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	rv := reflect.ValueOf(value)
	kind := rv.Kind().String()
	switch {
	case strings.HasPrefix(kind, "int"), strings.HasPrefix(kind, "uint") && kind != "uintptr":
		return true
	case strings.HasPrefix(kind, "float"):
		if kind == "float64" {
			val, _ := value.(float64)
			if val-float64(int(val)) > 0 {
				return false
			}
			form[field] = int(val)
			return true
		}

		val, _ := value.(float32)
		if val-float32(int(val)) > 0 {
			return false
		}
		form[field] = int(val)
		return true
	case kind == "string":
		intVal, err := strconv.Atoi(value.(string))
		if err == nil {
			form[field] = intVal
		}
		return err == nil
	default:
		return false
	}
}

// between

func validateMin(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	min, err := strconv.ParseFloat(parameters[0], 64)
	if err != nil {
		panic(err)
	}
	switch getFieldType(value) {
	case "numeric":
		floatValue, err := helpers.ToFloat64(value)
		if err != nil {
			panic(err)
		}
		return floatValue >= min
	case "string":
		return len(value.(string)) >= int(min)
	case "array":
		list := reflect.ValueOf(value)
		return list.Len() >= int(min)
	case "file":
		return false // TODO implement file min size
	}

	return true // Pass if field type cannot be checked (bool, dates, ...)
}

func validateMax(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	max, err := strconv.ParseFloat(parameters[0], 64)
	if err != nil {
		panic(err)
	}
	switch getFieldType(value) {
	case "numeric":
		floatValue, err := helpers.ToFloat64(value)
		if err != nil {
			panic(err)
		}
		return floatValue <= max
	case "string":
		return len(value.(string)) <= int(max)
	case "array":
		list := reflect.ValueOf(value)
		return list.Len() <= int(max)
	case "file":
		return false // TODO implement file min size
	}

	return true // Pass if field type cannot be checked (bool, dates, ...)
}

// greater_than + greater_than_equal
// lower_than + lower_than_equal

// -------------------------
// Strings

func validateString(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	_, ok := value.(string)
	return ok
}

// digits
// alpha
// alpha_dash
// alphanumeric
// email
// starts_with
// ends_with

// length
// min
// max

// ip address
// json
// regex
// timezone
// url
// uuid

// -------------------------
// Arrays

// array
// distinct

// -------------------------
// Dates

// date:format

// date before + before or equal
// date after + after or equal
// between + between or equal
// date equals
//

// -------------------------
// Files

// file
// mime
// image
// size
// extension

// -------------------------
// Misc

// boolean (accept 1, "on", "true", "yes")
// different
// in + not_in
// in_array + not_in_array
// nullable
// confirmed
