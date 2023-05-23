package validation

import "encoding/json"

// JSONValidator validates the field under validation must be a valid JSON string.
type JSONValidator struct{ BaseValidator }

// Validate checks the field under validation satisfies this validator's criteria.
func (v *JSONValidator) Validate(ctx *Context) bool {
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

// Name returns the string name of the validator.
func (v *JSONValidator) Name() string { return "json" }

// IsType returns true.
func (v *JSONValidator) IsType() bool { return true }

// JSON the field under validation must be a valid JSON string.
// Unmarshals the string and sets the field value to the unmarshalled result.
func JSON() *JSONValidator {
	return &JSONValidator{}
}
