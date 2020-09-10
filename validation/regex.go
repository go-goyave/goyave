package validation

import (
	"regexp"
	"sync"
)

const (
	patternAlpha        string = "^[\\pL\\pM]+$"
	patternAlphaDash    string = "^[\\pL\\pM0-9_-]+$"
	patternAlphaNumeric string = "^[\\pL\\pM0-9]+$"
	patternDigits       string = "[^0-9]"
	patternEmail        string = "^[^@\\r\\n\\t]{1,64}@[^\\s]+$"
)

var (
	regexCache = make(map[string]*regexp.Regexp, 5)
	mu         = sync.RWMutex{}
)

func getRegex(pattern string) *regexp.Regexp {
	mu.RLock()
	regex, exists := regexCache[pattern]
	mu.RUnlock()
	if !exists {
		regex = regexp.MustCompile(pattern)
		mu.Lock()
		regexCache[pattern] = regex
		mu.Unlock()
	}
	return regex
}

// ClearRegexCache empties the validation regex cache.
// Note that if validation.Validate is subsequently called, regex will need
// to be recompiled.
func ClearRegexCache() {
	regexCache = make(map[string]*regexp.Regexp, 5)
}
