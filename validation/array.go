package validation

import (
	"reflect"
)

// ArrayValidator validates the field under validation must be an array.
type ArrayValidator struct {
	BaseValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *ArrayValidator) Validate(ctx *Context) bool {
	if GetFieldType(ctx.Value) != FieldTypeArray {
		return false
	}
	ctx.Value = convertArray(ctx.Value, reflect.TypeOf(ctx.Parent))
	return true
}

// Name returns the string name of the validator.
func (v *ArrayValidator) Name() string { return "array" }

// IsType returns true.
func (v *ArrayValidator) IsType() bool { return true }

// Array the field under validation must be an array.
// On successful validation and if possible, converts the array to its correct type
// based on its elements' type. If all elements have the same type, the array is converted to
// a slice of this type.
func Array() *ArrayValidator {
	return &ArrayValidator{}
}

// convertArray to its correct type based on its elements' type.
// If all elements have the same type, the array is converted to
// a slice of this type, otherwise the array is returned as-is.
func convertArray(array any, parentType reflect.Type) any {
	list := reflect.ValueOf(array)
	length := list.Len()
	if length <= 0 {
		return array
	}

	elemVal := list.Index(0)
	if elemVal.Kind() != reflect.Interface {
		return array
	}
	elemVal = elemVal.Elem()
	if !elemVal.IsValid() {
		// The first element is probably `nil`, avoid "call of reflect.Value.Interface on zero Value" error.
		return array
	}
	elemType := elemVal.Type()
	for i := 1; i < length; i++ {
		elem := list.Index(i).Elem()
		if !elem.IsValid() || elem.Type() != elemType {
			// Not all elements have the same type, keep it []any
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
