package validation

import (
	"fmt"
	"time"

	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/walk"
)

// AfterValidator validates the field under validation must be a date (`time.Time`) before
// the specified date.
type AfterValidator struct {
	DateComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *AfterValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(t1, t2 time.Time) bool {
		return t1.After(t2)
	})
}

// Name returns the string name of the validator.
func (v *AfterValidator) Name() string { return "after" }

// After the field under validation must be a date (`time.Time`) before the given date.
func After(date time.Time) *AfterValidator {
	return &AfterValidator{DateComparisonValidator: DateComparisonValidator{Date: date}}
}

//------------------------------

// AfterEqualValidator validates the field under validation must be a date (`time.Time`) after
// or equal to the specified date.
type AfterEqualValidator struct {
	DateComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *AfterEqualValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(t1, t2 time.Time) bool {
		return t1.After(t2) || t1.Equal(t2)
	})
}

// Name returns the string name of the validator.
func (v *AfterEqualValidator) Name() string { return "after_equal" }

// AfterEqual the field under validation must be a date (`time.Time`) after or equal to the given date.
func AfterEqual(date time.Time) *AfterEqualValidator {
	return &AfterEqualValidator{DateComparisonValidator: DateComparisonValidator{Date: date}}
}

//------------------------------

// AfterFieldValidator validates the field under validation must be a date (`time.Time`) before
// all the other dates matched by the specified path.
type AfterFieldValidator struct {
	DateFieldComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *AfterFieldValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(t1, t2 time.Time) bool {
		return t1.After(t2)
	})
}

// Name returns the string name of the validator.
func (v *AfterFieldValidator) Name() string { return "after" }

// AfterField the field under validation must be a date (`time.Time`) before the date field identified
// by the given path.
func AfterField(path string) *AfterFieldValidator {
	p, err := walk.Parse(path)
	if err != nil {
		panic(errors.NewSkip(fmt.Errorf("validation.AfterField: path parse error: %w", err), 3))
	}
	return &AfterFieldValidator{DateFieldComparisonValidator: DateFieldComparisonValidator{Path: p}}
}

//------------------------------

// AfterEqualFieldValidator validates the field under validation must be a date (`time.Time`) after
// or equal to all the other dates matched by the specified path.
type AfterEqualFieldValidator struct {
	DateFieldComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *AfterEqualFieldValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(t1, t2 time.Time) bool {
		return t1.After(t2) || t1.Equal(t2)
	})
}

// Name returns the string name of the validator.
func (v *AfterEqualFieldValidator) Name() string { return "after_equal" }

// AfterEqualField the field under validation must be a date (`time.Time`) after or equal to the date field identified
// by the given path.
func AfterEqualField(path string) *AfterEqualFieldValidator {
	p, err := walk.Parse(path)
	if err != nil {
		panic(errors.NewSkip(fmt.Errorf("validation.AfterEqualField: path parse error: %w", err), 3))
	}
	return &AfterEqualFieldValidator{DateFieldComparisonValidator: DateFieldComparisonValidator{Path: p}}
}
