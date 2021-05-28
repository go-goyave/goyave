package validation

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"time"

	"github.com/google/uuid"
	"goyave.dev/goyave/v3/helper"
)

// createArray create a slice of the same type as the given type.
func createArray(dataType string, length int) reflect.Value {
	var arr reflect.Value
	switch dataType {
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
	return arr
}

func validateArray(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	if GetFieldType(value) == "array" {

		if len(parameters) == 0 {
			return true
		}

		if parameters[0] == "array" {
			panic("Cannot use array type for array validation. Use \">array\" instead")
		}

		if !validationRules[parameters[0]].IsType {
			panic(fmt.Sprintf("Rule %s is not converting, cannot use it for array validation", parameters[0]))
		}

		fieldName, _, parent, _ := GetFieldFromName(field, form)
		list := reflect.ValueOf(value)
		length := list.Len()
		arr := createArray(parameters[0], length)

		params := parameters[1:]

		for i := 0; i < length; i++ {
			val := list.Index(i).Interface()
			tmpData := map[string]interface{}{fieldName: val}
			if !validationRules[parameters[0]].Function(fieldName, val, params, tmpData) {
				return false
			}
			arr.Set(reflect.Append(arr, reflect.ValueOf(tmpData[fieldName])))
		}

		parent[fieldName] = arr.Interface()
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
	switch GetFieldType(value) {
	case "numeric":
		return !checkInNumeric(parameters, value)
	case "string":
		return !helper.ContainsStr(parameters, value.(string))
	}
	// Don't check arrays and files
	return false
}

func validateInArray(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	_, other, _, exists := GetFieldFromName(parameters[0], form)
	if exists && GetFieldType(other) == "array" {
		return helper.Contains(other, value)
	}
	return false
}

func validateNotInArray(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	_, other, _, exists := GetFieldFromName(parameters[0], form)
	if exists && GetFieldType(other) == "array" {
		return !helper.Contains(other, value)
	}
	return false
}
