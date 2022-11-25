package validation

// ObjectValidator the field under validation must be an object (`map[string]any`).
// If the value of the field under validation is a valid JSON string that can be unmarshalled
// into a `map[string]any`, converts the value to `map[string]any`.
type ObjectValidator struct{ BaseValidator }

// Validate checks the field under validation satisfies this validator's criteria.
func (v *ObjectValidator) Validate(ctx *ContextV5) bool {
	_, ok := ctx.Value.(map[string]any)
	if !ok {
		if (&JSONValidator{}).Validate(ctx) {
			_, ok = ctx.Value.(map[string]any)
		}
	}
	return ok
}

// Name returns the string name of the validator.
func (v *ObjectValidator) Name() string { return "object" }

// IsType returns true.
func (v *ObjectValidator) IsType() bool { return true }

// Object the field under validation must be an object (`map[string]any`).
// If the value of the field under validation is a valid JSON string that can be unmarshalled
// into a `map[string]any`, converts the value to `map[string]any`.
func Object() *ObjectValidator {
	return &ObjectValidator{}
}
