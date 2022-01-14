package validation

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"time"

	"github.com/google/uuid"
	"goyave.dev/goyave/v4/util/sliceutil"
	"goyave.dev/goyave/v4/util/typeutil"
)

// createArray create a slice of the same type as the given type.
func createArray(dataType string, length int) reflect.Value {
	var arr reflect.Value
	switch dataType {
	case "string":
		newArray := make([]string, length)
		arr = reflect.ValueOf(&newArray).Elem()
	case "numeric":
		newArray := make([]float64, length)
		arr = reflect.ValueOf(&newArray).Elem()
	case "integer":
		newArray := make([]int, length)
		arr = reflect.ValueOf(&newArray).Elem()
	case "timezone":
		newArray := make([]*time.Location, length)
		arr = reflect.ValueOf(&newArray).Elem()
	case "ip", "ipv4", "ipv6":
		newArray := make([]net.IP, length)
		arr = reflect.ValueOf(&newArray).Elem()
	case "json":
		newArray := make([]interface{}, length)
		arr = reflect.ValueOf(&newArray).Elem()
	case "url":
		newArray := make([]*url.URL, length)
		arr = reflect.ValueOf(&newArray).Elem()
	case "uuid":
		newArray := make([]uuid.UUID, length)
		arr = reflect.ValueOf(&newArray).Elem()
	case "bool":
		newArray := make([]bool, length)
		arr = reflect.ValueOf(&newArray).Elem()
	case "date":
		newArray := make([]time.Time, length)
		arr = reflect.ValueOf(&newArray).Elem()
	case "object":
		newArray := make([]map[string]interface{}, length)
		arr = reflect.ValueOf(&newArray).Elem()
	default:
		panic(fmt.Sprintf("Unsupported array type %q", dataType))
	}
	// TODO only works with built-in type rules
	return arr
}

// convertArray to its correct type based on its elements' type.
// If all elements have the same type, the array is converted to
// a slice of this type.
func convertArray(array interface{}, parentType reflect.Type) interface{} {
	list := reflect.ValueOf(array)
	length := list.Len()
	if length <= 0 {
		return array
	}

	elemVal := list.Index(0)
	if elemVal.Kind() != reflect.Interface {
		return array
	}
	elemType := elemVal.Elem().Type()
	for i := 1; i < length; i++ {
		if list.Index(i).Elem().Type() != elemType {
			// Not all elements have the same type, keep it []interface{}
			return array
		}
	}

	if !elemType.AssignableTo(parentType.Elem()) {
		return array
	}

	convertedArray := reflect.MakeSlice(reflect.SliceOf(elemType), 0, length)
	for i := 0; i < length; i++ {
		convertedArray = reflect.Append(convertedArray, list.Index(i).Elem())
	}

	return convertedArray.Interface()
}

func validateArray(ctx *Context) bool {
	if GetFieldType(ctx.Value) == "array" {

		parentType := reflect.TypeOf(ctx.Parent)

		if len(ctx.Rule.Params) == 0 {
			ctx.Value = convertArray(ctx.Value, parentType)
			return true
		}

		if ctx.Rule.Params[0] == "array" {
			panic("Cannot use array type for array validation. Use \"fieldName[]\" instead")
		}

		if !validationRules[ctx.Rule.Params[0]].IsType {
			panic(fmt.Sprintf("Rule %s is not converting, cannot use it for array validation", ctx.Rule.Params[0]))
		}

		list := reflect.ValueOf(ctx.Value)
		length := list.Len()
		arr := createArray(ctx.Rule.Params[0], length)
		if !arr.Type().AssignableTo(parentType.Elem()) {
			arr = list
		}

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
			arr.Index(i).Set(reflect.ValueOf(tmpCtx.Value))
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
		if sliceutil.Contains(found, v) {
			return false
		}
		found = append(found, v)
	}

	return true
}

func checkInNumeric(parameters []string, value interface{}) bool {
	for _, v := range parameters {
		floatVal, _ := typeutil.ToFloat64(value)
		other, err := typeutil.ToFloat64(v)
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
		return sliceutil.Contains(ctx.Rule.Params, ctx.Value)
	}
	// Don't check arrays and files
	return false
}

func validateNotIn(ctx *Context) bool {
	switch GetFieldType(ctx.Value) {
	case "numeric":
		return !checkInNumeric(ctx.Rule.Params, ctx.Value)
	case "string":
		return !sliceutil.ContainsStr(ctx.Rule.Params, ctx.Value.(string))
	}
	// Don't check arrays and files
	return false
}

func validateInArray(ctx *Context) bool {
	_, other, _, exists := GetFieldFromName(ctx.Rule.Params[0], ctx.Data)
	if exists && GetFieldType(other) == "array" {
		return sliceutil.Contains(other, ctx.Value)
	}
	return false
}

func validateNotInArray(ctx *Context) bool {
	_, other, _, exists := GetFieldFromName(ctx.Rule.Params[0], ctx.Data)
	if exists && GetFieldType(other) == "array" {
		return !sliceutil.Contains(other, ctx.Value)
	}
	return false
}
