package validation

import "regexp"

var emailRegex = regexp.MustCompile(`^[^@\r\n\t]{1,64}@[^\s]+$`)

// EmailValidator the field under validation must be a string that
// matches a basic email regexp.
// This validator is not enough in itself to properly validate an email address.
// The only way to ensure an email address is valid is by sending a confirmation email.
type EmailValidator struct {
	RegexValidator
}

// Name returns the string name of the validator.
func (v *EmailValidator) Name() string { return "email" }

// Email the field under validation must be a string that matches a basic email regexp.
// This validator is not enough in itself to properly validate an email address.
// The only way to ensure an email address is valid is by sending a confirmation email.
func Email() *EmailValidator {
	return &EmailValidator{RegexValidator: RegexValidator{Regexp: emailRegex}}
}
