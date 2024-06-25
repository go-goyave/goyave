package validation

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/walk"
)

// InValidator validates the field under validation must be a one of the given values.
type InValidator[T comparable] struct {
	BaseValidator
	Values []T
}

// Validate checks the field under validation satisfies this validator's criteria.
// Always return false if the validated value is not of type `T`.
func (v *InValidator[T]) Validate(ctx *Context) bool {
	val, ok := ctx.Value.(T)
	if !ok {
		return false
	}
	return lo.ContainsBy(v.Values, func(v T) bool {
		return val == v
	})
}

// Name returns the string name of the validator.
func (v *InValidator[T]) Name() string { return "in" }

// MessagePlaceholders returns the ":values placeholder.
func (v *InValidator[T]) MessagePlaceholders(_ *Context) []string {
	return []string{
		":values", strings.Join(lo.Map(v.Values, func(v T, _ int) string { return fmt.Sprintf("%v", v) }), ", "),
	}
}

// In the field under validation must be a one of the given values.
func In[T comparable](values []T) *InValidator[T] {
	return &InValidator[T]{Values: values}
}

//------------------------------

// NotInValidator validates the field undervalidation must not be a one of the given values.
type NotInValidator[T comparable] struct {
	BaseValidator
	Values []T
}

// Validate checks the field under validation satisfies this validator's criteria.
// Always return false if the validated value is not of type `T`or the matched arrays
// are not of type `[]T`.
func (v *NotInValidator[T]) Validate(ctx *Context) bool {
	val, ok := ctx.Value.(T)
	if !ok {
		return false
	}
	return !lo.ContainsBy(v.Values, func(v T) bool {
		return val == v
	})
}

// Name returns the string name of the validator.
func (v *NotInValidator[T]) Name() string { return "not_in" }

// MessagePlaceholders returns the ":values placeholder.
func (v *NotInValidator[T]) MessagePlaceholders(_ *Context) []string {
	return []string{
		":values", strings.Join(lo.Map(v.Values, func(v T, _ int) string { return fmt.Sprintf("%v", v) }), ", "),
	}
}

// NotIn the field under validation must not be a one of the given values.
func NotIn[T comparable](values []T) *NotInValidator[T] {
	return &NotInValidator[T]{Values: values}
}

//------------------------------

// InFieldValidator validates the field under validation must be in at least one
// of the arrays matched by the specified path.
type InFieldValidator[T comparable] struct {
	BaseValidator
	Path *walk.Path
}

// Validate checks the field under validation satisfies this validator's criteria.
// Always return false if the validated value is not of type `T` or the matched arrays
// are not of type `[]T`.
func (v *InFieldValidator[T]) Validate(ctx *Context) bool {
	val, ok := ctx.Value.(T)
	if !ok {
		return false
	}

	ok = false
	v.Path.Walk(ctx.Data, func(c *walk.Context) {
		if c.Path.Type == walk.PathTypeArray && c.Found == walk.ElementNotFound {
			return
		}

		list, okList := c.Value.([]T)
		if !okList || c.Found != walk.Found {
			return
		}

		if lo.Contains(list, val) {
			ok = true
			c.Break()
		}
	})
	return ok
}

// Name returns the string name of the validator.
func (v *InFieldValidator[T]) Name() string { return "in_field" }

// MessagePlaceholders returns the ":other" placeholder.
func (v *InFieldValidator[T]) MessagePlaceholders(_ *Context) []string {
	return []string{
		":other", GetFieldName(v.Lang(), v.Path),
	}
}

// InField the field under validation must be in at least one
// of the arrays matched by the specified path.
func InField[T comparable](path string) *InFieldValidator[T] {
	p, err := walk.Parse(path)
	if err != nil {
		panic(errors.NewSkip(fmt.Errorf("validation.InField: path parse error: %w", err), 3))
	}
	return &InFieldValidator[T]{Path: p}
}

//------------------------------

// NotInFieldValidator validates the field under validation must not be in any
// of the arrays matched by the specified path.
type NotInFieldValidator[T comparable] struct {
	InFieldValidator[T]
}

// Validate checks the field under validation satisfies this validator's criteria.
// Always return false if the validated value is not of type `T`.
func (v *NotInFieldValidator[T]) Validate(ctx *Context) bool {
	_, ok := ctx.Value.(T)
	if !ok {
		return false
	}

	return !v.InFieldValidator.Validate(ctx)
}

// Name returns the string name of the validator.
func (v *NotInFieldValidator[T]) Name() string { return "not_in_field" }

// NotInField the field under validation must not be in any
// of the arrays matched by the specified path.
func NotInField[T comparable](path string) *NotInFieldValidator[T] {
	p, err := walk.Parse(path)
	if err != nil {
		panic(errors.NewSkip(fmt.Errorf("validation.NotInField: path parse error: %w", err), 3))
	}
	return &NotInFieldValidator[T]{InFieldValidator: InFieldValidator[T]{Path: p}}
}
