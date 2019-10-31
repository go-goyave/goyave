package validation

import "regexp"

const (
	patternAlpha        string = "^[\\pL\\pM]+$"
	patternAlphaDash    string = "^[\\pL\\pM0-9_-]+$"
	patternAlphaNumeric string = "^[\\pL\\pM0-9]+$"
	patternDigits       string = "[^0-9]"
	patternEmail        string = "^[^@\\r\\n\\t]{1,64}@[^\\s]+$"
)

var (
	regexDigits = regexp.MustCompile(patternDigits)
)
