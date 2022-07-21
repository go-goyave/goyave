package validation

import (
	"reflect"

	"github.com/Code-Hex/uniseg"
	"goyave.dev/goyave/v4/util/fsutil"
	"goyave.dev/goyave/v4/util/typeutil"
)

type Required struct{ BaseValidator }

func (v *Required) Validate(ctx *ContextV5) bool {
	if !ctx.Field.IsNullable() && ctx.Value == nil {
		return false
	}
	return true
}

func (v *Required) Name() string { return "required" }

type Nullable struct{ BaseValidator }

func (v *Nullable) Validate(ctx *ContextV5) bool {
	return true
}

func (v *Nullable) Name() string { return "nullable" }

// TODO design: try with just a function (Name() may not be required)
// TODO design: Message() in interface? or register rules as before? -> would allow more computation for more complex messages instead of placeholders

type Between struct {
	BaseValidator
	Min int
	Max int
}

func (v *Between) Validate(ctx *ContextV5) bool {
	switch GetFieldType(ctx.Value) {
	case "numeric":
		floatValue, _ := typeutil.ToFloat64(ctx.Value)
		return floatValue >= float64(v.Min) && floatValue <= float64(v.Max)
	case "string":
		length := uniseg.GraphemeClusterCount(ctx.Value.(string))
		return length >= v.Min && length <= v.Max
	case "array":
		list := reflect.ValueOf(ctx.Value)
		length := list.Len()
		return length >= v.Min && length <= v.Max
	case "file":
		files, _ := ctx.Value.([]fsutil.File)
		for _, file := range files {
			minSize := int64(v.Min) * 1024
			maxSize := int64(v.Max) * 1024
			if file.Header.Size < minSize || file.Header.Size > maxSize {
				return false
			}
		}
		return true
	}

	return true // Pass if field type cannot be checked (bool, dates, ...)
}

func (v *Between) Name() string { return "between" }
