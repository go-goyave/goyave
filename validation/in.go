package validation

import (
	"reflect"

	"github.com/samber/lo"
)

// InValidator validates the field under validation must be a one of the given values.
type InValidator[T comparable] struct {
	BaseValidator
	Values []T
}

// Validate checks the field under validation satisfies this validator's criteria.
// Always return false if the validated value is not comparable.
func (v *InValidator[T]) Validate(ctx *ContextV5) bool {
	t := reflect.TypeOf(ctx.Value)
	if !t.Comparable() {
		return false
	}
	return lo.ContainsBy(v.Values, func(v T) bool {
		return ctx.Value == v
	})
}

// Name returns the string name of the validator.
func (v *InValidator[T]) Name() string { return "in" }

// In the field under validation must be a one of the given values.
func In[T comparable](values []T) *InValidator[T] {
	return &InValidator[T]{Values: values}
}

// NotInValidator validates the field undervalidation must not be a one of the given values.
type NotInValidator[T comparable] struct {
	InValidator[T]
}

// Validate checks the field under validation satisfies this validator's criteria.
// Always return false if the validated value is not comparable.
func (v *NotInValidator[T]) Validate(ctx *ContextV5) bool {
	t := reflect.TypeOf(ctx.Value)
	if !t.Comparable() {
		return false
	}
	return !lo.ContainsBy(v.Values, func(v T) bool {
		return ctx.Value == v
	})
}

// Name returns the string name of the validator.
func (v *NotInValidator[T]) Name() string { return "not_in" }

// NotIn the field under validation must not be a one of the given values.
func NotIn[T comparable](values []T) *NotInValidator[T] {
	return &NotInValidator[T]{
		InValidator: InValidator[T]{Values: values},
	}
}
