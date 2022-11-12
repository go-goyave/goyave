package validation

type StringValidator struct{ BaseValidator }

func (v *StringValidator) Validate(ctx *ContextV5) bool {
	_, ok := ctx.Value.(string)
	return ok
}

func (v *StringValidator) Name() string { return "string" }
func (v *StringValidator) IsType() bool { return true }

func String() *StringValidator {
	return &StringValidator{}
}
