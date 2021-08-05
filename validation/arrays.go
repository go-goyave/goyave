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
	case "object":
		newArray := make([]map[string]interface{}, 0, length)
		arr = reflect.ValueOf(&newArray).Elem()
	}
	// TODO only works with built-in type rules
	return arr
}

func validateArray(ctx *Context) bool {
	if GetFieldType(ctx.Value) == "array" {

		if len(ctx.Rule.Params) == 0 {
			return true
		}

		if ctx.Rule.Params[0] == "array" {
			panic("Cannot use array type for array validation. Use \">array\" instead")
		}

		if !validationRules[ctx.Rule.Params[0]].IsType {
			panic(fmt.Sprintf("Rule %s is not converting, cannot use it for array validation", ctx.Rule.Params[0]))
		}

		list := reflect.ValueOf(ctx.Value)
		length := list.Len()
		arr := createArray(ctx.Rule.Params[0], length)

		for i := 0; i < length; i++ {
			val := list.Index(i).Interface()
			tmpCtx := &Context{
				Value: val,
				Rule: &Rule{
					Name:   ctx.Rule.Params[0],
					Params: ctx.Rule.Params[1:],
				},
				Data: ctx.Data,
			}
			if !validationRules[ctx.Rule.Params[0]].Function(tmpCtx) {
				return false
			}
			arr.Set(reflect.Append(arr, reflect.ValueOf(tmpCtx.Value)))
		}

		ctx.Value = arr.Interface()
		return true
	}

	return false
}

func validateDistinct(ctx *Context) bool {
	if GetFieldType(ctx.Value) != "array" {
		return false // Can't validate if not an array
	}

	found := []interface{}{}
	list := reflect.ValueOf(ctx.Value)
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

func validateIn(ctx *Context) bool {
	switch GetFieldType(ctx.Value) {
	case "numeric":
		return checkInNumeric(ctx.Rule.Params, ctx.Value)
	case "string":
		return helper.Contains(ctx.Rule.Params, ctx.Value)
	}
	// Don't check arrays and files
	return false
}

func validateNotIn(ctx *Context) bool {
	switch GetFieldType(ctx.Value) {
	case "numeric":
		return !checkInNumeric(ctx.Rule.Params, ctx.Value)
	case "string":
		return !helper.ContainsStr(ctx.Rule.Params, ctx.Value.(string))
	}
	// Don't check arrays and files
	return false
}

func validateInArray(ctx *Context) bool {
	_, other, _, exists := GetFieldFromName(ctx.Rule.Params[0], ctx.Data)
	if exists && GetFieldType(other) == "array" {
		return helper.Contains(other, ctx.Value)
	}
	return false
}

func validateNotInArray(ctx *Context) bool {
	_, other, _, exists := GetFieldFromName(ctx.Rule.Params[0], ctx.Data)
	if exists && GetFieldType(other) == "array" {
		return !helper.Contains(other, ctx.Value)
	}
	return false
}
