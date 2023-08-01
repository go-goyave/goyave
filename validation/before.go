package validation

import (
	"fmt"
	"time"

	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/walk"
)

// BeforeValidator validates the field under validation must be a date (`time.Time`) before
// the specified date.
type BeforeValidator struct {
	DateComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *BeforeValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(t1, t2 time.Time) bool {
		return t1.Before(t2)
	})
}

// Name returns the string name of the validator.
func (v *BeforeValidator) Name() string { return "before" }

// Before the field under validation must be a date (`time.Time`) before the given date.
func Before(date time.Time) *BeforeValidator {
	return &BeforeValidator{DateComparisonValidator: DateComparisonValidator{Date: date}}
}

//------------------------------

// BeforeEqualValidator validates the field under validation must be a date (`time.Time`) before
// or equal to the specified date.
type BeforeEqualValidator struct {
	DateComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *BeforeEqualValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(t1, t2 time.Time) bool {
		return t1.Before(t2) || t1.Equal(t2)
	})
}

// Name returns the string name of the validator.
func (v *BeforeEqualValidator) Name() string { return "before_equal" }

// BeforeEqual the field under validation must be a date (`time.Time`) before or equal to the given date.
func BeforeEqual(date time.Time) *BeforeEqualValidator {
	return &BeforeEqualValidator{DateComparisonValidator: DateComparisonValidator{Date: date}}
}

//------------------------------

// BeforeFieldValidator validates the field under validation must be a date (`time.Time`) before
// all the other dates matched by the specified path.
type BeforeFieldValidator struct {
	DateFieldComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *BeforeFieldValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(t1, t2 time.Time) bool {
		return t1.Before(t2)
	})
}

// Name returns the string name of the validator.
func (v *BeforeFieldValidator) Name() string { return "before" }

// BeforeField the field under validation must be a date (`time.Time`) before the date field identified
// by the given path.
func BeforeField(path string) *BeforeFieldValidator {
	p, err := walk.Parse(path)
	if err != nil {
		panic(errors.NewSkip(fmt.Errorf("validation.BeforeField: path parse error: %w", err), 3))
	}
	return &BeforeFieldValidator{DateFieldComparisonValidator: DateFieldComparisonValidator{Path: p}}
}

//------------------------------

// BeforeEqualFieldValidator validates the field under validation must be a date (`time.Time`) before
// or equal to all the other dates matched by the specified path.
type BeforeEqualFieldValidator struct {
	DateFieldComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *BeforeEqualFieldValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(t1, t2 time.Time) bool {
		return t1.Before(t2) || t1.Equal(t2)
	})
}

// Name returns the string name of the validator.
func (v *BeforeEqualFieldValidator) Name() string { return "before_equal" }

// BeforeEqualField the field under validation must be a date (`time.Time`) before or equal to the date field identified
// by the given path.
func BeforeEqualField(path string) *BeforeEqualFieldValidator {
	p, err := walk.Parse(path)
	if err != nil {
		panic(errors.NewSkip(fmt.Errorf("validation.BeforeEqualField: path parse error: %w", err), 3))
	}
	return &BeforeEqualFieldValidator{DateFieldComparisonValidator: DateFieldComparisonValidator{Path: p}}
}
