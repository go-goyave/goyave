package validation

import (
	"reflect"
)

type ArrayValidator struct {
	BaseValidator
}

func (v *ArrayValidator) Validate(ctx *ContextV5) bool {
	if GetFieldType(ctx.Value) != FieldTypeArray {
		return false
	}
	ctx.Value = convertArray(ctx.Value, reflect.TypeOf(ctx.Parent))
	return true
}

func (v *ArrayValidator) Name() string { return "array" }

func Array() *ArrayValidator {
	return &ArrayValidator{}
}
