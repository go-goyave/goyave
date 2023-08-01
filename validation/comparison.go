package validation

import (
	"fmt"

	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/walk"
)

// ComparisonValidator validates the field under validation is greater than field identified
// by the given path. Mixed types are supported, meaning you can use this rule for the following (non-exhaustive) cases:
//   - Compare the length of two strings
//   - Compare the value of two numeric fields
//   - Compare a numeric field with the length of a string or a string length with a numeric field
//   - Compare a numeric field with the number of elements in an array
//   - Compare the number of keys in an object with a numeric field
//   - Compare a file (or multifile) size with a numeric field. The number of KiB of each file is rounded up (ceil).
type ComparisonValidator struct {
	BaseValidator
	Path *walk.Path
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *ComparisonValidator) validate(ctx *Context, comparisonFunc func(size1, size2 float64) bool) bool {
	floatValue, isNumber, overflowErr := numberAsFloat64(ctx.Value)
	if overflowErr != nil {
		return false
	}

	ok := true
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

		comparedFloatValue, isComparedNumber, comparedOverflowErr := numberAsFloat64(c.Value)
		if comparedOverflowErr != nil {
			ok = false
			c.Break()
			return
		}

		if isNumber {
			if isComparedNumber {
				ok = comparisonFunc(floatValue, comparedFloatValue)
			} else {
				ok = validateSize(c.Value, func(size int) bool {
					return comparisonFunc(floatValue, float64(size))
				})
			}
		} else {
			if isComparedNumber {
				ok = validateSize(ctx.Value, func(size int) bool {
					return comparisonFunc(float64(size), comparedFloatValue)
				})
			} else {
				ok = validateSize(ctx.Value, func(size1 int) bool {
					return validateSize(c.Value, func(size2 int) bool {
						return comparisonFunc(float64(size1), float64(size2))
					})
				})
			}
		}

		if !ok {
			c.Break()
		}
	})
	return ok
}

// IsTypeDependent returns true
func (v *ComparisonValidator) IsTypeDependent() bool { return true }

// MessagePlaceholders returns the ":other" placeholder.
func (v *ComparisonValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":other", GetFieldName(v.Lang(), v.Path),
	}
}

//------------------------------

// GreaterThanValidator validates the field under validation is greater than the field identified
// by the given path. See `ComparisonValidator` for more details.
type GreaterThanValidator struct {
	ComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *GreaterThanValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(size1, size2 float64) bool {
		return size1 > size2
	})
}

// Name returns the string name of the validator.
func (v *GreaterThanValidator) Name() string { return "greater_than" }

// GreaterThan validates the field under validation is greater than the field identified
// by the given path. Mixed types are supported, meaning you can use this rule for the following (non-exhaustive) cases:
//   - Compare the length of two strings
//   - Compare the value of two numeric fields
//   - Compare a numeric field with the length of a string or a string length with a numeric field
//   - Compare a numeric field with the number of elements in an array
//   - Compare the number of keys in an object with a numeric field
//   - Compare a file (or multifile) size with a numeric field. The number of KiB of each file is rounded up (ceil).
func GreaterThan(path string) *GreaterThanValidator {
	p, err := walk.Parse(path)
	if err != nil {
		panic(errors.NewSkip(fmt.Errorf("validation.GreaterThan: path parse error: %w", err), 3))
	}
	return &GreaterThanValidator{ComparisonValidator: ComparisonValidator{Path: p}}
}

//------------------------------

// GreaterThanEqualValidator validates the field under validation is greater than the field identified
// by the given path. See `ComparisonValidator` for more details.
type GreaterThanEqualValidator struct {
	ComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *GreaterThanEqualValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(size1, size2 float64) bool {
		return size1 >= size2
	})
}

// Name returns the string name of the validator.
func (v *GreaterThanEqualValidator) Name() string { return "greater_than_equal" }

// GreaterThanEqual validates the field under validation is greater or equal to the field identified
// by the given path. Mixed types are supported, meaning you can use this rule for the following (non-exhaustive) cases:
//   - Compare the length of two strings
//   - Compare the value of two numeric fields
//   - Compare a numeric field with the length of a string or a string length with a numeric field
//   - Compare a numeric field with the number of elements in an array
//   - Compare the number of keys in an object with a numeric field
//   - Compare a file (or multifile) size with a numeric field. The number of KiB of each file is rounded up (ceil).
func GreaterThanEqual(path string) *GreaterThanEqualValidator {
	p, err := walk.Parse(path)
	if err != nil {
		panic(errors.NewSkip(fmt.Errorf("validation.GreaterThanEqual: path parse error: %w", err), 3))
	}
	return &GreaterThanEqualValidator{ComparisonValidator: ComparisonValidator{Path: p}}
}

//------------------------------

// LowerThanValidator validates the field under validation is lower than the field identified
// by the given path. See `ComparisonValidator` for more details.
type LowerThanValidator struct {
	ComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *LowerThanValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(size1, size2 float64) bool {
		return size1 < size2
	})
}

// Name returns the string name of the validator.
func (v *LowerThanValidator) Name() string { return "lower_than" }

// LowerThan validates the field under validation is lower than the field identified
// by the given path. Mixed types are supported, meaning you can use this rule for the following (non-exhaustive) cases:
//   - Compare the length of two strings
//   - Compare the value of two numeric fields
//   - Compare a numeric field with the length of a string or a string length with a numeric field
//   - Compare a numeric field with the number of elements in an array
//   - Compare the number of keys in an object with a numeric field
//   - Compare a file (or multifile) size with a numeric field. The number of KiB of each file is rounded up (ceil).
func LowerThan(path string) *LowerThanValidator {
	p, err := walk.Parse(path)
	if err != nil {
		panic(errors.NewSkip(fmt.Errorf("validation.LowerThan: path parse error: %w", err), 3))
	}
	return &LowerThanValidator{ComparisonValidator: ComparisonValidator{Path: p}}
}

//------------------------------

// LowerThanEqualValidator validates the field under validation is lower or equal to the field identified
// by the given path. See `ComparisonValidator` for more details.
type LowerThanEqualValidator struct {
	ComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *LowerThanEqualValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(size1, size2 float64) bool {
		return size1 <= size2
	})
}

// Name returns the string name of the validator.
func (v *LowerThanEqualValidator) Name() string { return "lower_than_equal" }

// LowerThanEqual validates the field under validation is lower or equal to the field identified
// by the given path. Mixed types are supported, meaning you can use this rule for the following (non-exhaustive) cases:
//   - Compare the length of two strings
//   - Compare the value of two numeric fields
//   - Compare a numeric field with the length of a string or a string length with a numeric field
//   - Compare a numeric field with the number of elements in an array
//   - Compare the number of keys in an object with a numeric field
//   - Compare a file (or multifile) size with a numeric field. The number of KiB of each file is rounded up (ceil).
func LowerThanEqual(path string) *LowerThanEqualValidator {
	p, err := walk.Parse(path)
	if err != nil {
		panic(errors.NewSkip(fmt.Errorf("validation.LowerThanEqual: path parse error: %w", err), 3))
	}
	return &LowerThanEqualValidator{ComparisonValidator: ComparisonValidator{Path: p}}
}
