package validation

import (
	"strings"
)

// DoesntEndWithValidator validates that the string under validation doesn't end with the given suffix.
type DoesntEndWithValidator struct {
	BaseValidator
	Suffix string
}

// Validate checks if the field under validation satisfies this validator's criteria.
func (v *DoesntEndWithValidator) Validate(ctx *Context) bool {
	str, ok := ctx.Value.(string)
	if !ok {
		return false
	}

	return !strings.HasSuffix(str, v.Suffix)
}

// Name returns the string name of the validator.
func (v *DoesntEndWithValidator) Name() string { return "doesnt_end_with" }

// MessagePlaceholders returns the ":suffix" placeholder.
func (v *DoesntEndWithValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":suffix", v.Suffix,
	}
}

// DoesntEndWith creates a new DoesntEndWithValidator.
func DoesntEndWith(suffix string) *DoesntEndWithValidator {
	return &DoesntEndWithValidator{Suffix: suffix}
}
