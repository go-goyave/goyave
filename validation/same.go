package validation

import (
	"fmt"
	"reflect"

	"goyave.dev/goyave/v4/util/walk"
)

// SameValidator validates the field under validation is strictly equal to the field identified
// by the given path. Values of different types are never equal. Files are not checked and will never pass this validator.
// For arrays, objects and numbers, the values are compared using `reflect.DeepEqual()`.
type SameValidator struct {
	BaseValidator
	Path *walk.Path
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *SameValidator) Validate(ctx *ContextV5) bool {
	fieldType := GetFieldType(ctx.Value)
	ok := true
	v.Path.Walk(ctx.Data, func(c *walk.Context) {
		if c.Path.Type == walk.PathTypeArray && c.Found == walk.ElementNotFound {
			return
		}

		if c.Found != walk.Found {
			ok = false
			c.Break()
			return
		}

		switch fieldType {
		case FieldTypeString:
			str, okStr := c.Value.(string)
			ok = okStr && ctx.Value.(string) == str
		case FieldTypeBool:
			b, okBool := c.Value.(bool)
			ok = okBool && ctx.Value.(bool) == b
		case FieldTypeArray, FieldTypeObject, FieldTypeNumeric:
			ok = reflect.DeepEqual(ctx.Value, c.Value)
		default:
			// We don't check the other types
			ok = false
		}

		if !ok {
			c.Break()
		}
	})
	return ok
}

// Name returns the string name of the validator.
func (v *SameValidator) Name() string { return "same" }

// MessagePlaceholders returns the ":other" placeholder.
func (v *SameValidator) MessagePlaceholders(ctx *ContextV5) []string {
	return []string{
		":other", GetFieldName(v.Lang(), v.Path),
	}
}

// Same validates the field under validation is strictly equal to the field identified
// by the given path. Values of different types are never equal. Files are not checked
// and will never pass this validator.
// For arrays, objects and numbers, the values are compared using `reflect.DeepEqual()`.
func Same(path string) *SameValidator {
	p, err := walk.Parse(path)
	if err != nil {
		panic(fmt.Errorf("validation.Same: path parse error: %w", err))
	}
	return &SameValidator{Path: p}
}
