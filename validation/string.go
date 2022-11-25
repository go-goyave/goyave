package validation

import (
	"strings"

	"github.com/samber/lo"
)

// StringValidator the field under validation must be a string.
type StringValidator struct{ BaseValidator }

// Validate checks the field under validation satisfies this validator's criteria.
func (v *StringValidator) Validate(ctx *ContextV5) bool {
	_, ok := ctx.Value.(string)
	return ok
}

// Name returns the string name of the validator.
func (v *StringValidator) Name() string { return "string" }

// IsType returns true.
func (v *StringValidator) IsType() bool { return true }

// String the field under validation must be a string.
func String() *StringValidator {
	return &StringValidator{}
}

//------------------------------

// StartsWithValidator the field under validation must be a string starting
// with at least one of the specified prefixes.
type StartsWithValidator struct {
	BaseValidator
	Prefix []string
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *StartsWithValidator) Validate(ctx *ContextV5) bool {
	val, ok := ctx.Value.(string)
	return ok && lo.ContainsBy(v.Prefix, func(prefix string) bool {
		return strings.HasPrefix(val, prefix)
	})
}

// Name returns the string name of the validator.
func (v *StartsWithValidator) Name() string { return "starts_with" }

// MessagePlaceholders returns the ":values" placeholder.
func (v *StartsWithValidator) MessagePlaceholders(ctx *ContextV5) []string {
	return []string{
		":values", strings.Join(v.Prefix, ", "),
	}
}

// StartsWith the field under validation must be a string starting
// with at least one of the specified prefixes.
func StartsWith(prefix ...string) *StartsWithValidator {
	return &StartsWithValidator{Prefix: prefix}
}

//------------------------------

// EndsWithValidator the field under validation must be a string ending
// with at least one of the specified suffixes.
type EndsWithValidator struct {
	BaseValidator
	Suffix []string
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *EndsWithValidator) Validate(ctx *ContextV5) bool {
	val, ok := ctx.Value.(string)
	return ok && lo.ContainsBy(v.Suffix, func(prefix string) bool {
		return strings.HasSuffix(val, prefix)
	})
}

// Name returns the string name of the validator.
func (v *EndsWithValidator) Name() string { return "ends_with" }

// MessagePlaceholders returns the ":values" placeholder.
func (v *EndsWithValidator) MessagePlaceholders(ctx *ContextV5) []string {
	return []string{
		":values", strings.Join(v.Suffix, ", "),
	}
}

// EndsWith the field under validation must be a string ending
// with at least one of the specified prefixes.
func EndsWith(suffix ...string) *EndsWithValidator {
	return &EndsWithValidator{Suffix: suffix}
}
