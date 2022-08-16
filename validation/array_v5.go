package validation

import (
	"reflect"
)

type Array struct {
	BaseValidator
}

func (v *Array) Validate(ctx *ContextV5) bool {
	if GetFieldType(ctx.Value) != "array" {
		return false
	}
	ctx.Value = convertArray(ctx.Value, reflect.TypeOf(ctx.Parent))
	return true
}

func (v *Array) Name() string { return "array" }
