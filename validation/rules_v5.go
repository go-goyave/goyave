package validation

import (
	"fmt"
	"reflect"

	"github.com/Code-Hex/uniseg"
	"goyave.dev/goyave/v4/util/fsutil"
	"goyave.dev/goyave/v4/util/typeutil"
	"goyave.dev/goyave/v4/util/walk"
)

type RequiredValidator struct{ BaseValidator }

func (v *RequiredValidator) Validate(ctx *ContextV5) bool {
	if !ctx.Field.IsNullable() && ctx.Value == nil {
		return false
	}
	return true
}

func (v *RequiredValidator) Name() string { return "required" }

func Required() *RequiredValidator {
	return &RequiredValidator{}
}

type NullableValidator struct{ BaseValidator }

func (v *NullableValidator) Validate(ctx *ContextV5) bool {
	return true
}

func (v *NullableValidator) Name() string { return "nullable" }

func Nullable() *NullableValidator {
	return &NullableValidator{}
}

// TODO design: Message() in interface? or register rules as before? -> would allow more computation for more complex messages instead of placeholders (and remove the difference between validation placeholders and regular placeholders)

type BetweenValidator struct {
	BaseValidator
	Min int
	Max int
}

func (v *BetweenValidator) Validate(ctx *ContextV5) bool {
	switch GetFieldType(ctx.Value) {
	case FieldTypeNumeric:
		floatValue, _ := typeutil.ToFloat64(ctx.Value)
		return floatValue >= float64(v.Min) && floatValue <= float64(v.Max)
	case FieldTypeString:
		length := uniseg.GraphemeClusterCount(ctx.Value.(string))
		return length >= v.Min && length <= v.Max
	case FieldTypeArray, FieldTypeObject: // TODO test for object (validates the number of keys)
		list := reflect.ValueOf(ctx.Value)
		length := list.Len()
		return length >= v.Min && length <= v.Max
	case FieldTypeFile:
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

func (v *BetweenValidator) Name() string          { return "between" }
func (v *BetweenValidator) IsTypeDependent() bool { return true }

func Between(min, max int) *BetweenValidator {
	return &BetweenValidator{Min: min, Max: max}
}

type GreaterThanValidator struct {
	BaseValidator
	Path string // TODO no need to keep the orignal path once walk.Path implements String()
	path *walk.Path
}

func (v *GreaterThanValidator) Validate(ctx *ContextV5) bool {
	valueType := GetFieldType(ctx.Value)

	ok := true
	v.path.Walk(ctx.Data, func(c walk.Context) {
		if !ok {
			return
		}
		if c.Path.Type == walk.PathTypeArray && c.Found == walk.ElementNotFound {
			return
		}

		if c.Found != walk.Found || valueType != GetFieldType(c.Value) {
			ok = false
			return // Can't compare two different types or missing field
		}

		switch valueType {
		case "numeric":
			floatValue, _ := typeutil.ToFloat64(ctx.Value)
			comparedFloatValue, _ := typeutil.ToFloat64(c.Value)
			ok = floatValue > comparedFloatValue
		case "string":
			ok = uniseg.GraphemeClusterCount(ctx.Value.(string)) > uniseg.GraphemeClusterCount(c.Value.(string))
		case "array":
			ok = reflect.ValueOf(ctx.Value).Len() > reflect.ValueOf(c.Value).Len()
		case "file":
			files, _ := ctx.Value.([]fsutil.File)
			comparedFiles, _ := c.Value.([]fsutil.File)
			for _, file := range files {
				for _, comparedFile := range comparedFiles {
					if file.Header.Size <= comparedFile.Header.Size {
						ok = false
						return
					}
				}
			}
		}
	})
	return ok
}

func (v *GreaterThanValidator) Name() string          { return "greater_than" }
func (v *GreaterThanValidator) IsTypeDependent() bool { return true }
func (v *GreaterThanValidator) ComparesWith() string  { return v.Path }

func GreaterThan(path string) *GreaterThanValidator {
	p, err := walk.Parse(path)
	if err != nil {
		panic(fmt.Errorf("validation.GreaterThan: path parse error: %w", err))
	}
	return &GreaterThanValidator{Path: path, path: p}
}
