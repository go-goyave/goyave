package validation

import "strings"

// TrimValidator if the field under validation is a string, trims it using
// `strings.TrimSpace()`.
type TrimValidator struct{ BaseValidator }

// Validate always returns true. If the field under validation is a string,
// trims it using `strings.TrimSpace()`.
func (v *TrimValidator) Validate(ctx *Context) bool {
	str, ok := ctx.Value.(string)
	if ok {
		ctx.Value = strings.TrimSpace(str)
	}

	// This rule is just transforming, so we always return true.
	return true
}

// Name returns the string name of the validator.
func (v *TrimValidator) Name() string { return "trim" }

// Trim if the field under validation is a string, trims it using `strings.TrimSpace()`.
func Trim() *TrimValidator {
	return &TrimValidator{}
}
