package validation

import (
	"reflect"
)

// ArrayValidator validates the field under validation must be an array.
type ArrayValidator struct {
	BaseValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *ArrayValidator) Validate(ctx *ContextV5) bool {
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
