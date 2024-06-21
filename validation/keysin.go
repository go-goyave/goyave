package validation

import "strings"

// KeysInValidator the field under validation must be an object and all its keys must
// be equal to one of the given values.
type KeysInValidator struct {
	BaseValidator
	Keys []string
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *KeysInValidator) Validate(ctx *Context) bool {
	obj, ok := ctx.Value.(map[string]any)
	if !ok {
		return false
	}

	allowedKeys := make(map[string]struct{})
	for _, key := range v.Keys {
		allowedKeys[key] = struct{}{}
	}

	for key := range obj {
		if _, ok := allowedKeys[key]; !ok {
			return false
		}
	}
	return true
}

// Name returns the string name of the validator.
func (v *KeysInValidator) Name() string { return "keys_in" }

// MessagePlaceholders returns the ":values" placeholder.
func (v *KeysInValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":values", strings.Join(v.Keys, ", "),
	}
}

// KeysIn the field under validation must be an object and all its keys must
// be equal to one of the given values.
func KeysIn(keys ...string) *KeysInValidator {
	return &KeysInValidator{Keys: keys}
}
