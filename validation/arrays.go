package validation

import (
	"log"
	"net"
	"net/url"
	"reflect"
	"time"

	"github.com/System-Glitch/goyave/v2/helper"
	"github.com/google/uuid"
)

func validateArray(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	if GetFieldType(value) == "array" {

		if len(parameters) == 0 {
			return true
		}

		if !isArrayType(parameters[0]) {
			log.Panicf("Rule %s is not converting, cannot use it for array validation", parameters[0])
		}

		list := reflect.ValueOf(value)
		length := list.Len()
		var arr reflect.Value

		switch parameters[0] {
		case "string":
			newArray := make([]string, 0, length)
			arr = reflect.ValueOf(&newArray).Elem()
		case "numeric":
			newArray := make([]float64, 0, length)
			arr = reflect.ValueOf(&newArray).Elem()
		case "integer":
			newArray := make([]int, 0, length)
			arr = reflect.ValueOf(&newArray).Elem()
		case "timezone":
			newArray := make([]*time.Location, 0, length)
			arr = reflect.ValueOf(&newArray).Elem()
		case "ip", "ipv4", "ipv6":
			newArray := make([]net.IP, 0, length)
			arr = reflect.ValueOf(&newArray).Elem()
		case "json":
			newArray := make([]interface{}, 0, length)
			arr = reflect.ValueOf(&newArray).Elem()
		case "url":
			newArray := make([]*url.URL, 0, length)
			arr = reflect.ValueOf(&newArray).Elem()
		case "uuid":
			newArray := make([]uuid.UUID, 0, length)
			arr = reflect.ValueOf(&newArray).Elem()
		case "bool":
			newArray := make([]bool, 0, length)
			arr = reflect.ValueOf(&newArray).Elem()
		case "date":
			newArray := make([]time.Time, 0, length)
			arr = reflect.ValueOf(&newArray).Elem()
		}

		params := parameters[1:]

		for i := 0; i < length; i++ {
			val := list.Index(i).Interface()
			tmpData := map[string]interface{}{field: val}
			if !validationRules[parameters[0]](field, val, params, tmpData) {
				return false
			}
			arr.Set(reflect.Append(arr, reflect.ValueOf(tmpData[field])))
		}

		form[field] = arr.Interface()
		return true
	}

	return false
}

func validateDistinct(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	if GetFieldType(value) != "array" {
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
	switch GetFieldType(value) {
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
	switch GetFieldType(value) {
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
	if exists && GetFieldType(other) == "array" {
		return helper.Contains(other, value)
	}
	return false
}

func validateNotInArray(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("not_in_array", parameters, 1)
	other, exists := form[parameters[0]]
	if exists && GetFieldType(other) == "array" {
		return !helper.Contains(other, value)
	}
	return false
}
