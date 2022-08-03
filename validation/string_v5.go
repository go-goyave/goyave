package validation

type String struct{ BaseValidator }

func (v *String) Validate(ctx *ContextV5) bool {
	_, ok := ctx.Value.(string)
	return ok
}

func (v *String) Name() string { return "string" }
func (v *String) IsType() bool { return true }
