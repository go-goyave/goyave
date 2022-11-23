package validation

type NullableValidator struct{ BaseValidator }

func (v *NullableValidator) Validate(ctx *ContextV5) bool {
	return true
}

func (v *NullableValidator) Name() string { return "nullable" }

func Nullable() *NullableValidator {
	return &NullableValidator{}
}
