package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoesntEndWithValidator(t *testing.T) {
	tests := []struct {
		value    string
		suffix   string
		expected bool
	}{
		{"example", "ple", false},
		{"example", "exa", true},
		{"test", "st", false},
		{"test", "te", true},
		{"golang", "lang", false},
		{"golang", "go", true},
	}

	for _, test := range tests {
		validator := DoesntEndWith(test.suffix)
		ctx := &Context{Value: test.value}
		assert.Equal(t, test.expected, validator.Validate(ctx))
	}
}
