package validation

// RequiredValidator the field under validation is required.
// If a field is absent from the input data, subsequent validators
// will not be executed.
//
// If a field is `nil` and has the `Nullable` validator, this validator passes.
// As non-nullable fields are removed if they have a `nil` value, this validator
// doesn't pass if a field is `nil` and doesn't have the `Nullable` validator.
type RequiredValidator struct{ BaseValidator }

// Validate checks the field under validation satisfies this validator's criteria.
func (v *RequiredValidator) Validate(ctx *Context) bool {
	if !ctx.Field.IsNullable() && ctx.Value == nil {
		return false
	}
	return true
}

// Name returns the string name of the validator.
func (v *RequiredValidator) Name() string { return "required" }

// Required the field under validation is required.
// If a field is absent from the input data, subsequent validators
// will not be executed.
//
// If a field is `nil` and has the `Nullable` validator, this validator passes.
// As non-nullable fields are removed if they have a `nil` value, this validator
// doesn't pass if a field is `nil` and doesn't have the `Nullable` validator.
func Required() *RequiredValidator {
	return &RequiredValidator{}
}

//------------------------------

// RequiredIfValidator is the same as `RequiredValidator` but only applies the behavior
// described if the specified `Condition` function returns true.
type RequiredIfValidator struct {
	RequiredValidator
	Condition func(*Context) bool
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *RequiredIfValidator) Validate(ctx *Context) bool {
	if !v.Condition(ctx) {
		return true
	}
	return v.RequiredValidator.Validate(ctx)
}

// RequiredIf is the same as `Required` but only applies the behavior
// described if the specified condition function returns true.
func RequiredIf(condition func(*Context) bool) *RequiredIfValidator {
	return &RequiredIfValidator{Condition: condition}
}
