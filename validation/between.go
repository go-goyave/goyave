package validation

import (
	"fmt"
)

// BetweenValidator validates the field under validation depending on its type.
//   - Numbers are directly compared if they fit in `float64`. If they don't the rule doesn't pass.
//   - Strings must have a length between min and max characters (calculated based on the number of grapheme clusters)
//   - Arrays must have at between min and max elements
//   - Objects must have at between min and max keys
//   - Files must weight between min and max KiB (for multi-files, all files must match this criteria). The number of KiB of each file is rounded up (ceil).
//
// All comparisons are inclusive.
type BetweenValidator struct {
	BaseValidator
	Min float64
	Max float64
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *BetweenValidator) Validate(ctx *Context) bool {
	fl, ok, err := numberAsFloat64(ctx.Value)
	if ok {
		return fl >= v.Min && fl <= v.Max
	}
	if err != nil {
		return false
	}
	return validateSize(ctx.Value, func(size int) bool {
		s := float64(size)
		return s >= v.Min && s <= v.Max
	})
}

// Name returns the string name of the validator.
func (v *BetweenValidator) Name() string { return "between" }

// IsTypeDependent returns true
func (v *BetweenValidator) IsTypeDependent() bool { return true }

// MessagePlaceholders returns the ":min" and ":max" placeholder.
func (v *BetweenValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":min", fmt.Sprintf("%v", v.Min),
		":max", fmt.Sprintf("%v", v.Max),
	}
}

// Between validates the field under validation depending on its type.
//   - Numbers are directly compared if they fit in `float64`. If they don't the rule doesn't pass.
//   - Strings must have a length between min and max characters (calculated based on the number of grapheme clusters)
//   - Arrays must have at between min and max elements
//   - Objects must have at between min and max keys
//   - Files must weight between min and max KiB (for multi-files, all files must match this criteria). The number of KiB of each file is rounded up (ceil).
//
// All comparisons are inclusive.
func Between(min, max float64) *BetweenValidator {
	return &BetweenValidator{Min: min, Max: max}
}
