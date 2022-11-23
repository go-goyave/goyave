package validation

type RequiredValidator struct{ BaseValidator }

func (v *RequiredValidator) Validate(ctx *ContextV5) bool {
	if !ctx.Field.IsNullable() && ctx.Value == nil {
		return false
	}
	return true
}

func (v *RequiredValidator) Name() string { return "required" }

func Required() *RequiredValidator {
	return &RequiredValidator{}
}

type RequiredIfValidator struct {
	RequiredValidator
	Condition func(*ContextV5) bool
}

func (v *RequiredIfValidator) Validate(ctx *ContextV5) bool {
	if !v.Condition(ctx) {
		return true
	}
	return v.RequiredValidator.Validate(ctx)
}

func RequiredIf(condition func(*ContextV5) bool) *RequiredIfValidator {
	return &RequiredIfValidator{Condition: condition}
}
