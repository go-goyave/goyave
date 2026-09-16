package validation

import (
	"context"
)

// UniqueValidator validates the field under validation must not match an existing record in the database.
type UniqueValidator[T any] struct {
	fn func(ctx context.Context, key T) (bool, error)
	BaseValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *UniqueValidator[T]) Validate(ctx *Context) bool {
	value, ok := ctx.Value.(T)
	if ctx.Invalid || !ok {
		return true
	}
	unique, err := v.fn(ctx.Context, value)
	if err != nil {
		ctx.AddError(err)
		return false
	}
	return unique
}

// Name returns the string name of the validator.
func (v *UniqueValidator[T]) Name() string { return "unique" }

// Unique validates the field under validation must not match an existing record in database
// according to the provided function. The function should be a service method returning
// true if no record match the given key.
//
// It is recommended to use a count query, or [database.Unique.Check].
func Unique[T any](fn func(ctx context.Context, key T) (bool, error)) *UniqueValidator[T] {
	return &UniqueValidator[T]{fn: fn}
}

//------------------------------

// ExistsValidator validates the field under validation must match an existing record in the database.
type ExistsValidator[T any] struct {
	fn func(ctx context.Context, key T) (bool, error)
	BaseValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *ExistsValidator[T]) Validate(ctx *Context) bool {
	value, ok := ctx.Value.(T)
	if ctx.Invalid || !ok {
		return true
	}
	exists, err := v.fn(ctx.Context, value)
	if err != nil {
		ctx.AddError(err)
		return false
	}
	return exists
}

// Name returns the string name of the validator.
func (v *ExistsValidator[T]) Name() string { return "exists" }

// Exists validates the field under validation must match an existing record in database according
// to the provided function. The function should be a service method returning
// true if a record matches the given key.
//
// It is recommended to use a count query, or [database.Exist.Check].
func Exists[T any](fn func(ctx context.Context, key T) (bool, error)) *ExistsValidator[T] {
	return &ExistsValidator[T]{fn: fn}
}

//------------------------------

// ExistsArrayValidator validates the field under validation must be an array and all
// of its elements must have a matching record in database.
// The type `T` is the type of the elements of the array under validation.
type ExistsArrayValidator[T any] struct {
	fn func(ctx context.Context, keys []T) ([]int, error)
	BaseValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *ExistsArrayValidator[T]) Validate(ctx *Context) bool {
	values, ok := ctx.Value.([]T)
	if ctx.Invalid || !ok {
		return true
	}

	results, err := v.fn(ctx.Context, values)
	if err != nil {
		ctx.AddError(err)
	}

	ctx.AddArrayElementValidationErrors(results...)
	return true
}

// Name returns the string name of the validator.
func (v *ExistsArrayValidator[T]) Name() string { return "exists" }

// ExistsArray validates the field under validation must be an array and all of its elements
// must have a matching record in database according to the provided function.
// The function should be a service method returning the indexes of the "keys" that do NOT
// match any existing record.
//
// This is preferable to use this validation rule on the array instead of [Exists] on
// each array element because this rule is designed to only execute a single SQL query instead of
// as many as there are elements in the array.
//
// It is recommended to use a count query, or [database.Exist.CheckSlice].
func ExistsArray[T any](fn func(ctx context.Context, keys []T) ([]int, error)) *ExistsArrayValidator[T] {
	return &ExistsArrayValidator[T]{fn: fn}
}

//------------------------------

// UniqueArrayValidator validates the field under validation must be an array and all
// of its elements must NOT have a matching record in database.
// The type `T` is the type of the elements of the array under validation.
type UniqueArrayValidator[T any] struct {
	ExistsArrayValidator[T]
}

// Name returns the string name of the validator.
func (v *UniqueArrayValidator[T]) Name() string { return "unique" }

// UniqueArray validates the field under validation must be an array and all of its elements
// must NOT have a matching record in database according to the provided function.
// The function should be a service method returning the indexes of the "keys" that
// match an existing record.
//
// This is preferable to use this validation rule on the array instead of [Unique] on
// each array element because this rule is designed to only execute a single SQL query instead of
// as many as there are elements in the array.
//
// It is recommended to use a count query, or [database.Unique.CheckSlice].
func UniqueArray[T any](fn func(ctx context.Context, keys []T) ([]int, error)) *UniqueArrayValidator[T] {
	return &UniqueArrayValidator[T]{fn: fn}
}
