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
