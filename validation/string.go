package validation

import (
	"strings"

	"github.com/samber/lo"
)

// StringValidator the field under validation must be a string.
type StringValidator struct{ BaseValidator }

// Validate checks the field under validation satisfies this validator's criteria.
func (v *StringValidator) Validate(ctx *Context) bool {
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
func (v *StartsWithValidator) Validate(ctx *Context) bool {
	val, ok := ctx.Value.(string)
	return ok && lo.ContainsBy(v.Prefix, func(prefix string) bool {
		return strings.HasPrefix(val, prefix)
	})
}

// Name returns the string name of the validator.
func (v *StartsWithValidator) Name() string { return "starts_with" }

// MessagePlaceholders returns the ":values" placeholder.
func (v *StartsWithValidator) MessagePlaceholders(_ *Context) []string {
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
func (v *EndsWithValidator) Validate(ctx *Context) bool {
	val, ok := ctx.Value.(string)
	return ok && lo.ContainsBy(v.Suffix, func(prefix string) bool {
		return strings.HasSuffix(val, prefix)
	})
}

// Name returns the string name of the validator.
func (v *EndsWithValidator) Name() string { return "ends_with" }

// MessagePlaceholders returns the ":values" placeholder.
func (v *EndsWithValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":values", strings.Join(v.Suffix, ", "),
	}
}

// EndsWith the field under validation must be a string ending
// with at least one of the specified prefixes.
func EndsWith(suffix ...string) *EndsWithValidator {
	return &EndsWithValidator{Suffix: suffix}
}

//------------------------------

// DoesntStartWithValidator the field under validation must be a string not starting
// with any of the specified prefixes.
type DoesntStartWithValidator struct {
	BaseValidator
	Prefix []string
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *DoesntStartWithValidator) Validate(ctx *Context) bool {
	val, ok := ctx.Value.(string)
	return ok && !lo.ContainsBy(v.Prefix, func(prefix string) bool {
		return strings.HasPrefix(val, prefix)
	})
}

// Name returns the string name of the validator.
func (v *DoesntStartWithValidator) Name() string { return "doesnt_start_with" }

// MessagePlaceholders returns the ":values" placeholder.
func (v *DoesntStartWithValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":values", strings.Join(v.Prefix, ", "),
	}
}

// DoesntStartWith the field under validation must be a string not starting
// with any of the specified prefixes.
func DoesntStartWith(prefix ...string) *DoesntStartWithValidator {
	return &DoesntStartWithValidator{Prefix: prefix}
}

//------------------------------

// DoesntEndWithValidator the field under validation must be a string not ending
// with any of the specified prefixes.
type DoesntEndWithValidator struct {
	BaseValidator
	Suffix []string
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *DoesntEndWithValidator) Validate(ctx *Context) bool {
	val, ok := ctx.Value.(string)
	return ok && !lo.ContainsBy(v.Suffix, func(prefix string) bool {
		return strings.HasSuffix(val, prefix)
	})
}

// Name returns the string name of the validator.
func (v *DoesntEndWithValidator) Name() string { return "doesnt_end_with" }

// MessagePlaceholders returns the ":values" placeholder.
func (v *DoesntEndWithValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":values", strings.Join(v.Suffix, ", "),
	}
}

// DoesntEndWith the field under validation must be a string not ending
// with any of the specified prefixes.
func DoesntEndWith(suffix ...string) *DoesntEndWithValidator {
	return &DoesntEndWithValidator{Suffix: suffix}
}
