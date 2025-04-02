package validation

// OnlyIfValidator execute a validator only if a condition is met.
type OnlyIfValidator struct {
	Validator
	Condition func(*Context) bool
}

// Validate executes the embedded validator only if the Condition returns true.
// Otherwise immediately returns true.
func (v *OnlyIfValidator) Validate(ctx *Context) bool {
	if !v.Condition(ctx) {
		return true
	}
	return v.Validator.Validate(ctx)
}

// OnlyIf execute the given Validator only if the condition returns true.
// This enables conditional validation.
//
// For example, if you want a validator to be executed only if another boolean field
// is "true":
//
//	  v.OnlyIf(func(ctx *v.Context) bool {
//		  other := walk.MustParse("other").First(ctx.Data)
//		  return other.Value == true
//	  }, MyValidator())
//
// This CANNOT be used with `Required()`, `RequiredIf()`, `Nullable()` or any type validator.
func OnlyIf(condition func(*Context) bool, validator Validator) *OnlyIfValidator {
	return &OnlyIfValidator{
		Validator: validator,
		Condition: condition,
	}
}
