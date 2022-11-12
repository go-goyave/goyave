package validation

type ObjectValidator struct{ BaseValidator }

func (v *ObjectValidator) Validate(ctx *ContextV5) bool {
	_, ok := ctx.Value.(map[string]any)
	if !ok {
		if (&JSONValidator{}).Validate(ctx) {
			_, ok = ctx.Value.(map[string]any)
		}
	}
	return ok
}

func (v *ObjectValidator) Name() string { return "object" }
func (v *ObjectValidator) IsType() bool { return true }

func Object() *ObjectValidator {
	return &ObjectValidator{}
}
