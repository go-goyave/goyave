package validation

import (
	"fmt"
	"time"

	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/walk"
)

// DateValidator validates the field under validation must be a string representing a date.
type DateValidator struct {
	BaseValidator
	Formats []string
}

func (v *DateValidator) parseDate(date any) (time.Time, bool) {
	if d, ok := date.(time.Time); ok {
		return d, true
	}
	str, ok := date.(string)
	if ok {
		for _, format := range v.Formats {
			t, err := time.Parse(format, str)
			if err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *DateValidator) Validate(ctx *Context) bool {
	date, ok := v.parseDate(ctx.Value)
	if ok {
		ctx.Value = date
	}
	return ok
}

// Name returns the string name of the validator.
func (v *DateValidator) Name() string { return "date" }

// IsType returns true.
func (v *DateValidator) IsType() bool { return true }

// Date the field under validation must be a string representing a date.
// On successful validation, converts the value to `time.Time`.
//
// The date must match at least one of the provided date formats (by order of preference).
// The format uses the same syntax as Go's standard datetime format.
// If no format is given the "2006-01-02" format is used.
func Date(acceptedFormats ...string) *DateValidator {
	if len(acceptedFormats) == 0 {
		acceptedFormats = []string{time.DateOnly}
	}
	return &DateValidator{Formats: acceptedFormats}
}

//------------------------------

// DateComparisonValidator factorized date comparison validator for static dates (before, after, etc.)
type DateComparisonValidator struct {
	BaseValidator
	Date time.Time
}

func (v *DateComparisonValidator) validate(ctx *Context, comparisonFunc func(time.Time, time.Time) bool) bool {
	date, ok := ctx.Value.(time.Time)
	if !ok {
		return false
	}
	return comparisonFunc(date, v.Date)
}

// MessagePlaceholders returns the ":date" placeholder.
func (v *DateComparisonValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":date", v.Date.Format(time.RFC3339),
	}
}

//------------------------------

// DateFieldComparisonValidator factorized date comparison validator for field dates (before field, after field, etc.)
type DateFieldComparisonValidator struct {
	BaseValidator
	Path *walk.Path
}

func (v *DateFieldComparisonValidator) validate(ctx *Context, comparisonFunc func(time.Time, time.Time) bool) bool {
	date, ok := ctx.Value.(time.Time)
	if !ok {
		return false
	}

	ok = true
	v.Path.Walk(ctx.Data, func(c *walk.Context) {
		lastParent := c.Path.LastParent()
		if lastParent != nil && lastParent.Type == walk.PathTypeArray && c.Found == walk.ElementNotFound {
			return
		}

		otherDate, isDate := c.Value.(time.Time)
		if !isDate || c.Found != walk.Found {
			ok = false
			c.Break()
			return // Can't compare two different types or missing field
		}

		if !comparisonFunc(date, otherDate) {
			ok = false
			c.Break()
		}
	})
	return ok
}

// MessagePlaceholders returns the ":date" placeholder.
func (v *DateFieldComparisonValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":date", GetFieldName(v.Lang(), v.Path),
	}
}

//------------------------------

// DateEqualsValidator validates the field under validation must be a date (`time.Time`)
// equal to the specified date.
type DateEqualsValidator struct {
	DateComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *DateEqualsValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(t1, t2 time.Time) bool {
		return t1.Equal(t2)
	})
}

// Name returns the string name of the validator.
func (v *DateEqualsValidator) Name() string { return "date_equals" }

// DateEquals the field under validation must be a date (`time.Time`) equal to the given date.
func DateEquals(date time.Time) *DateEqualsValidator {
	return &DateEqualsValidator{DateComparisonValidator: DateComparisonValidator{Date: date}}
}

//------------------------------

// DateEqualsFieldValidator validates the field under validation must be a date (`time.Time`)
// equal to all the other dates matched by the specified path.
type DateEqualsFieldValidator struct {
	DateFieldComparisonValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *DateEqualsFieldValidator) Validate(ctx *Context) bool {
	return v.validate(ctx, func(t1, t2 time.Time) bool {
		return t1.Equal(t2)
	})
}

// Name returns the string name of the validator.
func (v *DateEqualsFieldValidator) Name() string { return "date_equals" }

// DateEqualsField the field under validation must be a date (`time.Time`) equal to the date field identified
// by the given path.
func DateEqualsField(path string) *DateEqualsFieldValidator {
	p, err := walk.Parse(path)
	if err != nil {
		panic(errors.NewSkip(fmt.Errorf("validation.DateEqualsField: path parse error: %w", err), 3))
	}
	return &DateEqualsFieldValidator{DateFieldComparisonValidator: DateFieldComparisonValidator{Path: p}}
}
