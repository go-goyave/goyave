package validation

import (
	"reflect"
	"strconv"
	"strings"

	"goyave.dev/goyave/v4/util/typeutil"
)

func validateNumeric(ctx *Context) bool {
	rv := reflect.ValueOf(ctx.Value)
	kind := rv.Kind().String()
	switch {
	case strings.HasPrefix(kind, "float"):
		return true
	case strings.HasPrefix(kind, "int"), strings.HasPrefix(kind, "uint") && kind != "uintptr":
		floatVal, err := typeutil.ToFloat64(ctx.Value)
		ok := err == nil
		if ok {
			ctx.Value = floatVal
		}
		return ok
	case kind == "string":
		floatVal, err := strconv.ParseFloat(ctx.Value.(string), 64)
		ok := err == nil
		if ok {
			ctx.Value = floatVal
		}
		return ok
	default:
		return false
	}
}

func validateInteger(ctx *Context) bool {
	rv := reflect.ValueOf(ctx.Value)
	kind := rv.Kind().String()
	switch {
	case strings.HasPrefix(kind, "int"), strings.HasPrefix(kind, "uint") && kind != "uintptr":
		return true
	case strings.HasPrefix(kind, "float"):
		if kind == "float64" {
			val, _ := ctx.Value.(float64)
			if val-float64(int(val)) > 0 {
				return false
			}
			ctx.Value = int(val)
			return true
		}

		val, _ := ctx.Value.(float32)
		if val-float32(int(val)) > 0 {
			return false
		}
		ctx.Value = int(val)
		return true
	case kind == "string":
		intVal, err := strconv.Atoi(ctx.Value.(string))
		if err == nil {
			ctx.Value = intVal
		}
		return err == nil
	default:
		return false
	}
}
