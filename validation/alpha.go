package validation

import "regexp"

var (
	alphaRegex     = regexp.MustCompile(`^[\pL\pM]+$`)
	alphaNumRegex  = regexp.MustCompile(`^[\pL\pM0-9]+$`)
	alphaDashRegex = regexp.MustCompile(`^[\pL\pM0-9_-]+$`)
)

// AlphaValidator the field under validation must be an alphabetic string.
type AlphaValidator struct {
	RegexValidator
}

// Name returns the string name of the validator.
func (v *AlphaValidator) Name() string { return "alpha" }

// Alpha the field under validation must be an alphabetic string.
func Alpha() *AlphaValidator {
	return &AlphaValidator{RegexValidator: RegexValidator{Regexp: alphaRegex}}
}

//------------------------------

// AlphaNumValidator the field under validation must an alphabetic-numeric string.
type AlphaNumValidator struct {
	RegexValidator
}

// Name returns the string name of the validator.
func (v *AlphaNumValidator) Name() string { return "alpha_num" }

// AlphaNum the field under validation must an alphabetic-numeric string.
func AlphaNum() *AlphaNumValidator {
	return &AlphaNumValidator{RegexValidator: RegexValidator{Regexp: alphaNumRegex}}
}

//------------------------------

// AlphaDashValidator the field under validation must a string made
// of alphabetic-numeric characters, dashes or underscores.
type AlphaDashValidator struct {
	RegexValidator
}

// Name returns the string name of the validator.
func (v *AlphaDashValidator) Name() string { return "alpha_dash" }

// AlphaDash the field under validation must be a string made
// of alphabetic-numeric characters, dashes or underscores.
func AlphaDash() *AlphaDashValidator {
	return &AlphaDashValidator{RegexValidator: RegexValidator{Regexp: alphaDashRegex}}
}
