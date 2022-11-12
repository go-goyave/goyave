package validation

import "encoding/json"

type JSONValidator struct{ BaseValidator }

func (v *JSONValidator) Validate(ctx *ContextV5) bool {
	str, ok := ctx.Value.(string)
	if ok {
		var data any
		err := json.Unmarshal([]byte(str), &data)
		if err == nil {
			ctx.Value = data
			return true
		}
	}
	return false
}

func (v *JSONValidator) Name() string { return "json" }
func (v *JSONValidator) IsType() bool { return true }

func JSON() *JSONValidator {
	return &JSONValidator{}
}
