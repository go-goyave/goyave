package validation

// NullableValidator is a special validator indicating the `nil` values
// are accepted. If this field under validation is not
// nullable (is not validated by a `NullableValidator`) and its value is
// `nil`, it will be removed from the input data.
type NullableValidator struct{ BaseValidator }

// Validate returns true.
func (v *NullableValidator) Validate(_ *Context) bool {
	return true
}

// Name returns the string name of the validator.
func (v *NullableValidator) Name() string { return "nullable" }

// Nullable indicates `nil` values are accepted. If this field under validation is not
// nullable and its value is `nil`, it will be removed from the input data.
func Nullable() *NullableValidator {
	return &NullableValidator{}
}
