package validation

// PreserveValidator wraps another Validator and ensures that the original value
// in the validation context is preserved, regardless of whether validation passes.
// This is useful when you want to run validation without altering or converting
// the input value.
type PreserveValidator struct{ BaseValidator, inner Validator }

// Validate runs the wrapped validator's validation logic while preserving the
// original context value. The context value is restored after validation,
// even if the wrapped validator modifies it.
func (v *PreserveValidator) Validate(ctx *Context) bool {
	if v.inner == nil {
		return true
	}
	original := ctx.Value
	ok := v.inner.Validate(ctx)
	ctx.Value = original
	return ok
}

// MessagePlaceholders returns the message placeholders from the wrapped validator.
func (v *PreserveValidator) MessagePlaceholders(ctx *Context) []string {
	if v.inner == nil {
		return nil
	}
	return v.inner.MessagePlaceholders(ctx)
}

// Name returns the string identifier of this validator.
func (v *PreserveValidator) Name() string { return "preserve" }

// IsTypeDependent returns true if the wrapped validator is type-dependent.
func (v *PreserveValidator) IsTypeDependent() bool {
	if v.inner == nil {
		return false
	}
	return v.inner.IsTypeDependent()
}

// IsType returns false
func (v *PreserveValidator) IsType() bool {
	return false
}

// Init initializes the validator with the given options.
func (v *PreserveValidator) Init(opts *Options) {
	v.BaseValidator.Init(opts)
	if v.inner != nil {
		v.inner.Init(opts)
	}
}

// Preserve creates a new PreserveValidator that wraps the given validator.
// The wrapped validator will run its checks without modifying the original
// value in the validation context.
func Preserve(validator Validator) *PreserveValidator {
	return &PreserveValidator{inner: validator}
}
