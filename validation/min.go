package validation

import (
	"fmt"
)

// MinValidator validates the field under validation depending on its type.
//   - Numbers are directly compared if they fit in `float64`. If they don't the rule doesn't pass.
//   - Strings must have a length of at least n characters (calculated based on the number of grapheme clusters)
//   - Arrays must have at least n elements
//   - Objects must have at least n keys
//   - Files must weight at least n KiB (for multi-files, all files must match this criteria). The number of KiB of each file is rounded up (ceil).
type MinValidator struct {
	BaseValidator
	Min float64
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *MinValidator) Validate(ctx *Context) bool {
	fl, ok, err := numberAsFloat64(ctx.Value)
	if ok {
		return fl >= v.Min
	}
	if err != nil {
		return false
	}
	return validateSize(ctx.Value, func(size int) bool {
		return float64(size) >= v.Min
	})
}

// Name returns the string name of the validator.
func (v *MinValidator) Name() string { return "min" }

// IsTypeDependent returns true
func (v *MinValidator) IsTypeDependent() bool { return true }

// MessagePlaceholders returns the ":min" placeholder.
func (v *MinValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":min", fmt.Sprintf("%v", v.Min),
	}
}

// Min validates the field under validation depending on its type.
//   - Numbers are directly compared if they fit in `float64`. If they don't the rule doesn't pass.
//   - Strings must have a length of at least n characters (calculated based on the number of grapheme clusters)
//   - Arrays must have at least n elements
//   - Objects must have at least n keys
//   - Files must weight at least n KiB (for multi-files, all files must match this criteria). The number of KiB of each file is rounded up (ceil).
func Min(min float64) *MinValidator {
	return &MinValidator{Min: min}
}
