package validation

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/System-Glitch/goyave/v3/helper"
)

func validateNumeric(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	rv := reflect.ValueOf(value)
	kind := rv.Kind().String()
	switch {
	case strings.HasPrefix(kind, "float"):
		return true
	case strings.HasPrefix(kind, "int"), strings.HasPrefix(kind, "uint") && kind != "uintptr":
		floatVal, err := helper.ToFloat64(value)
		ok := err == nil
		if ok {
			form[field] = floatVal
		}
		return ok
	case kind == "string":
		floatVal, err := strconv.ParseFloat(value.(string), 64)
		ok := err == nil
		if ok {
			form[field] = floatVal
		}
		return ok
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
