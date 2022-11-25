package validation

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
