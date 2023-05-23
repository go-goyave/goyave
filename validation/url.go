package validation

import "net/url"

// URLValidator the field under validation must be a string representing
// a valid URL as per `url.ParseRequestURI()`.
// If validation passes, the value is converted to `*url.URL`.
type URLValidator struct{ BaseValidator }

// Validate checks the field under validation satisfies this validator's criteria.
func (v *URLValidator) Validate(ctx *Context) bool {
	if _, ok := ctx.Value.(*url.URL); ok {
		return true
	}
	val, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	url, err := url.ParseRequestURI(val)
	if err != nil {
		return false
	}
	ctx.Value = url
	return true
}

// Name returns the string name of the validator.
func (v *URLValidator) Name() string { return "url" }

// IsType returns true.
func (v *URLValidator) IsType() bool { return true }

// URL the field under validation must be a representing
// a valid URL as per `url.ParseRequestURI()`.
// If validation passes, the value is converted to `*url.URL`.
func URL() *URLValidator {
	return &URLValidator{}
}
