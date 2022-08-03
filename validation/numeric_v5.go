package validation

import (
	"reflect"
	"strconv"
	"strings"
)

type Integer struct{ BaseValidator }

func (v *Integer) Validate(ctx *ContextV5) bool {
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

func (v *Integer) Name() string { return "integer" }
func (v *Integer) IsType() bool { return true }
