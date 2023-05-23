package validation

import (
	"sync"
	"time"
)

var timezoneCache = sync.Map{}

// TimezoneValidator the field under validation must be a valid string
// reprensentation of a timezone.
// If validation passes, the value is converted to `*time.Location`
// using `time.LoadLocation()`.
// "Local" as an input is not accepted as a valid timezone.
//
// As `time.LoadLocation()` can be a slow operation, timezones are cached.
type TimezoneValidator struct{ BaseValidator }

// Validate checks the field under validation satisfies this validator's criteria.
func (v *TimezoneValidator) Validate(ctx *Context) bool {
	if _, ok := ctx.Value.(*time.Location); ok {
		return true
	}
	tz, ok := ctx.Value.(string)
	if !ok || tz == "Local" || tz == "" {
		return false
	}

	if loc, ok := timezoneCache.Load(tz); ok {
		ctx.Value = loc
		return true
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return false
	}

	timezoneCache.Store(tz, loc)
	ctx.Value = loc
	return true
}

// Name returns the string name of the validator.
func (v *TimezoneValidator) Name() string { return "timezone" }

// IsType returns true.
func (v *TimezoneValidator) IsType() bool { return true }

// Timezone the field under validation must be a valid string reprensentation of a timezone.
// If validation passes, the value is converted to `*time.Location` using `time.LoadLocation()`.
// "Local" as an input is not accepted as a valid timezone.
//
// As `time.LoadLocation()` can be a slow operation, timezones are cached.
func Timezone() *TimezoneValidator {
	return &TimezoneValidator{}
}
