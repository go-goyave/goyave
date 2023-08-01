package validation

import (
	"fmt"
	"reflect"

	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/walk"
)

// DifferentValidator validates the field under validation is not equal to the field identified
// by the given path. Values of different types are never equal. Files are not checked and will never pass this validator.
// For arrays, objects and numbers, the values are compared using `reflect.DeepEqual()`.
// For numbers, make sure the two compared numbers have the same type. A `uint` with value `1` will be considered
// different from an `int` with value `1`.
type DifferentValidator struct {
	BaseValidator
	Path *walk.Path
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *DifferentValidator) Validate(ctx *Context) bool {
	fieldType := GetFieldType(ctx.Value)
	ok := true

	if fieldType == FieldTypeUnsupported {
		// We cannot validate this field
		return true
	}

	v.Path.Walk(ctx.Data, func(c *walk.Context) {
		lastParent := c.Path.LastParent()
		if lastParent != nil && lastParent.Type == walk.PathTypeArray && c.Found == walk.ElementNotFound {
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
			ok = !okStr || ctx.Value.(string) != str
		case FieldTypeBool:
			b, okBool := c.Value.(bool)
			ok = !okBool || ctx.Value.(bool) != b
		case FieldTypeArray, FieldTypeObject, FieldTypeNumeric:
			ok = !reflect.DeepEqual(ctx.Value, c.Value)
		}

		if !ok {
			c.Break()
		}
	})
	return ok
}

// Name returns the string name of the validator.
func (v *DifferentValidator) Name() string { return "different" }

// MessagePlaceholders returns the ":other" placeholder.
func (v *DifferentValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":other", GetFieldName(v.Lang(), v.Path),
	}
}

// Different validates the field under validation is not equal to the field identified
// by the given path. Values of different types are never equal. Files are not checked
// and will never pass this validator.
// For arrays, objects and numbers, the values are compared using `reflect.DeepEqual()`.
// For numbers, make sure the two compared numbers have the same type. A `uint` with value `1` will be considered
// different from an `int` with value `1`.
func Different(path string) *DifferentValidator {
	p, err := walk.Parse(path)
	if err != nil {
		panic(errors.NewSkip(fmt.Errorf("validation.Different: path parse error: %w", err), 3))
	}
	return &DifferentValidator{Path: p}
}
