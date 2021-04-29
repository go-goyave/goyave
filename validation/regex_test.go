package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexCache(t *testing.T) {
	regex := getRegex(patternDigits)
	cached, exists := regexCache[patternDigits]

	assert.True(t, exists)
	assert.Equal(t, regex, cached)
	assert.Same(t, regex, cached)

	ClearRegexCache()
	assert.Empty(t, regexCache)
}
