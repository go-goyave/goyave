package validation

import "regexp"

// RegexValidator the field under validation must be a string matching
// the specified `*regexp.Regexp`.
type RegexValidator struct {
	BaseValidator
	Regexp *regexp.Regexp
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *RegexValidator) Validate(ctx *Context) bool {
	val, ok := ctx.Value.(string)
	return ok && v.Regexp.MatchString(val)
}

// Name returns the string name of the validator.
func (v *RegexValidator) Name() string { return "regex" }

// MessagePlaceholders returns the ":regexp" placeholder.
func (v *RegexValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":regexp", v.Regexp.String(),
	}
}

// Regex the field under validation must be a string matching
// the specified `*regexp.Regexp`.
func Regex(regex *regexp.Regexp) *RegexValidator {
	return &RegexValidator{Regexp: regex}
}
