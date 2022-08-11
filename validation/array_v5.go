package validation

import (
	"reflect"
)

type Array struct {
	BaseValidator
	ElementsType Validator
}

func (v *Array) Validate(ctx *ContextV5) bool {
	if GetFieldType(ctx.Value) == "array" {

		parentType := reflect.TypeOf(ctx.Parent)

		if v.ElementsType == nil {
			ctx.Value = convertArray(ctx.Value, parentType)
			return true
		}

		if v.ElementsType.Name() == "array" {
			panic("Cannot use array type for array validation. Use \"fieldName[]\" instead")
		}

		list := reflect.ValueOf(ctx.Value)
		length := list.Len()
		arr := createArray(ArrayType(v.ElementsType.Name()), length)
		if !arr.Type().AssignableTo(parentType.Elem()) {
			arr = list
		}

		for i := 0; i < length; i++ {
			val := list.Index(i).Interface()
			tmpCtx := &ContextV5{
				Options: ctx.Options,
				Data:    ctx.Data,
				Value:   val,
				Extra:   ctx.Extra,
				Parent:  ctx.Value,
				Now:     ctx.Now,
			}
			ok := v.ElementsType.Validate(tmpCtx)
			if len(tmpCtx.errors) > 0 {
				ctx.AddError(tmpCtx.errors...)
				return false
			}
			if !ok {
				return false
			}
			arr.Index(i).Set(reflect.ValueOf(tmpCtx.Value))
		}

		ctx.Value = arr.Interface()
		return true
	}

	return false
}

func (v *Array) Name() string { return "array" }
