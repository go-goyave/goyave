package validation

import (
	"fmt"
	"reflect"

	"github.com/Code-Hex/uniseg"
	"goyave.dev/goyave/v4/util/fsutil"
	"goyave.dev/goyave/v4/util/typeutil"
	"goyave.dev/goyave/v4/util/walk"
)

type GreaterThanValidatorProto struct {
	BaseValidator
	Path *walk.Path
}

func (v *GreaterThanValidatorProto) Validate(ctx *ContextV5) bool {
	valueType := GetFieldType(ctx.Value)

	ok := true
	v.Path.Walk(ctx.Data, func(c *walk.Context) {
		if c.Path.Type == walk.PathTypeArray && c.Found == walk.ElementNotFound {
			return
		}

		if c.Found != walk.Found || valueType != GetFieldType(c.Value) {
			ok = false
			c.Break()
			// TODO maybe we could? For example comparing a numeric with an array length (following the "size" rule principle)
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
						c.Break()
						return
					}
				}
			}
		}
		if !ok {
			c.Break()
		}
	})
	return ok
}

func (v *GreaterThanValidatorProto) Name() string          { return "greater_than" }
func (v *GreaterThanValidatorProto) IsTypeDependent() bool { return true }
func (v *GreaterThanValidatorProto) MessagePlaceholders(ctx *ContextV5) []string {
	return []string{
		":other", GetFieldName(v.Lang(), v.Path),
	}
}

func GreaterThanProto(path string) *GreaterThanValidatorProto {
	p, err := walk.Parse(path)
	if err != nil {
		panic(fmt.Errorf("validation.GreaterThan: path parse error: %w", err))
	}
	return &GreaterThanValidatorProto{Path: p}
}

// TODO implement more rules
// After that I think the new validation system will be complete
// Also will need a lot of documentation and testing
// TODO one file per rule
