package validation

import "strings"

// KeysInValidator validates the field under validation must contain all the given keys.
type KeysInValidator struct {
	BaseValidator
	Keys []string
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *KeysInValidator) Validate(ctx *Context) bool {
	obj, ok := ctx.Value.(map[string]interface{})
	if !ok {
		return false
	}

	for _, key := range v.Keys {
		if _, ok := obj[key]; !ok {
			return false
		}
	}
	return true
}

// Name returns the string name of the validator.
func (v *KeysInValidator) Name() string { return "keysin" }

// MessagePlaceholders returns the ":keys placeholder.
func (v *KeysInValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":keys", strings.Join(v.Keys, ", "),
	}
}

// KeysIn the field under validation must contain all the given keys.
func KeysIn(keys ...string) *KeysInValidator {
	return &KeysInValidator{Keys: keys}
}
