package validation

import (
	"reflect"
	"strconv"
	"strings"
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

// integer
// between
// min
// max

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
