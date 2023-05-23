package validation

import (
	"fmt"
)

// MaxValidator validates the field under validation depending on its type.
//   - Numbers are directly compared if they fit in `float64`. If they don't the rule doesn't pass.
//   - Strings must have a length of at most n characters (calculated based on the number of grapheme clusters)
//   - Arrays must have at most n elements
//   - Objects must have at most n keys
//   - Files must weight at most n KiB (for multi-files, all files must match this criteria). The number of KiB of each file is rounded up (ceil).
type MaxValidator struct {
	BaseValidator
	Max float64
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *MaxValidator) Validate(ctx *Context) bool {
	fl, ok, err := numberAsFloat64(ctx.Value)
	if ok {
		return fl <= v.Max
	}
	if err != nil {
		return false
	}
	return validateSize(ctx.Value, func(size int) bool {
		return float64(size) <= v.Max
	})
}

// Name returns the string name of the validator.
func (v *MaxValidator) Name() string { return "max" }

// IsTypeDependent returns true
func (v *MaxValidator) IsTypeDependent() bool { return true }

// MessagePlaceholders returns the ":max" placeholder.
func (v *MaxValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":max", fmt.Sprintf("%v", v.Max),
	}
}

// Max validates the field under validation depending on its type.
//   - Numbers are directly compared if they fit in `float64`. If they don't the rule doesn't pass.
//   - Strings must have a length of at most n characters (calculated based on the number of grapheme clusters)
//   - Arrays must have at most n elements
//   - Objects must have at most n keys
//   - Files must weight at most n KiB (for multi-files, all files must match this criteria). The number of KiB of each file is rounded up (ceil).
func Max(max float64) *MaxValidator {
	return &MaxValidator{Max: max}
}
