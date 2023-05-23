package validation

import "net/mail"

// EmailValidator the field under validation must be a string that can be parsed
// using Go's standard `mail.ParseAddress` function.
//
// The email address format is defined by RFC 5322. For example:
//   - Barry Gibbs <bg@example.com>
//   - foo@example.com
//
// This validator is not enough in itself to properly validate an email address.
// The only way to ensure an email address is valid is by sending a confirmation email.
//
// On successful validation, converts the value to `*mail.Address`.
type EmailValidator struct {
	BaseValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *EmailValidator) Validate(ctx *Context) bool {
	if _, ok := ctx.Value.(*mail.Address); ok {
		return true
	}
	val, ok := ctx.Value.(string)
	if !ok {
		return false
	}

	email, err := mail.ParseAddress(val)
	if err == nil {
		ctx.Value = email
	}

	return err == nil
}

// IsType returns true.
func (v *EmailValidator) IsType() bool {
	return true
}

// Name returns the string name of the validator.
func (v *EmailValidator) Name() string { return "email" }

// Email the field under validation must be a string that can be parsed using Go's standard
// `mail.ParseAddress` function.
//
// The email address format is defined by RFC 5322. For example:
//   - Barry Gibbs <bg@example.com>
//   - foo@example.com
//
// This validator is not enough in itself to properly validate an email address.
// The only way to ensure an email address is valid is by sending a confirmation email.
//
// On successful validation, converts the value to `*mail.Address`.
func Email() *EmailValidator {
	return &EmailValidator{}
}
