package validation

import "net"

// IPValidator the field under validation must be a string representing
// a valid IPv4 or IPv6.
// If validation passes, the value is converted to `net.IP`.
type IPValidator struct{ BaseValidator }

// Validate checks the field under validation satisfies this validator's criteria.
func (v *IPValidator) Validate(ctx *Context) bool {
	if _, ok := ctx.Value.(net.IP); ok {
		return true
	}

	val, ok := ctx.Value.(string)
	if !ok {
		return false
	}

	ip := net.ParseIP(val)
	if ip == nil {
		return false
	}

	ctx.Value = ip
	return true
}

// Name returns the string name of the validator.
func (v *IPValidator) Name() string { return "ip" }

// IsType returns true.
func (v *IPValidator) IsType() bool { return true }

// IP the field under validation must be a string representing
// a valid IPv4 or IPv6.
// If validation passes, the value is converted to `net.IP`.
func IP() *IPValidator {
	return &IPValidator{}
}

//------------------------------

// IPv4Validator the field under validation must be a string representing
// a valid IPv4.
// If validation passes, the value is converted to `net.IP`.
type IPv4Validator struct{ IPValidator }

// Validate checks the field under validation satisfies this validator's criteria.
func (v *IPv4Validator) Validate(ctx *Context) bool {
	if !v.IPValidator.Validate(ctx) {
		return false
	}
	return ctx.Value.(net.IP).To4() != nil
}

// Name returns the string name of the validator.
func (v *IPv4Validator) Name() string { return "ipv4" }

// IPv4 the field under validation must be a string representing a valid IPv4.
// If validation passes, the value is converted to `net.IP`.
func IPv4() *IPv4Validator {
	return &IPv4Validator{}
}

//------------------------------

// IPv6Validator the field under validation must be a string representing
// a valid IPv6.
// If validation passes, the value is converted to `net.IP`.
type IPv6Validator struct{ IPValidator }

// Validate checks the field under validation satisfies this validator's criteria.
func (v *IPv6Validator) Validate(ctx *Context) bool {
	if !v.IPValidator.Validate(ctx) {
		return false
	}
	return ctx.Value.(net.IP).To4() == nil
}

// Name returns the string name of the validator.
func (v *IPv6Validator) Name() string { return "ipv6" }

// IPv6 the field under validation must be a string representing a valid IPv6.
// If validation passes, the value is converted to `net.IP`.
func IPv6() *IPv6Validator {
	return &IPv6Validator{}
}
