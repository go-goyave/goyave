package validation

// BoolValidator the field under validation must be a bool or one of the following values:
//   - "1" / "0"
//   - "true" / "false"
//   - "yes" / "no"
//   - "on" / "off"
//   - a number different from 0 is converetd to `true`, a number equals to 0 is converted to `false`
//
// This rule converts the field to `bool` if it passes.
type BoolValidator struct {
	BaseValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *BoolValidator) Validate(ctx *Context) bool {
	switch val := ctx.Value.(type) {
	case bool:
		return true
	case string:
		switch val {
		case "1", "on", "true", "yes":
			ctx.Value = true
			return true
		case "0", "off", "false", "no":
			ctx.Value = false
			return true
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		f, _, _ := numberAsFloat64(val)
		ctx.Value = f != 0
		return true
	}
	return false
}

// Name returns the string name of the validator.
func (v *BoolValidator) Name() string { return "bool" }

// IsType returns true
func (v *BoolValidator) IsType() bool { return true }

// Bool the field under validation must be a bool or one of the following values:
//   - "1" / "0"
//   - "true" / "false"
//   - "yes" / "no"
//   - "on" / "off"
//   - a number different from 0 is converetd to `true`, a number equals to 0 is converted to `false`
//
// This rule converts the field to `bool` if it passes.
func Bool() *BoolValidator {
	return &BoolValidator{}
}
