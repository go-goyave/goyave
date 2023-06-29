package validation

import (
	"strings"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5/util/fsutil"
)

// ExtensionValidator validates the field under validation must be a file whose
// filename has one of the specified extensions as suffix.
// Multi-files are supported (all files must satisfy the criteria).
type ExtensionValidator struct {
	BaseValidator
	Extensions []string
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *ExtensionValidator) Validate(ctx *Context) bool {
	files, ok := ctx.Value.([]fsutil.File)
	if !ok {
		return false
	}

	for _, file := range files {
		i := strings.Index(file.Header.Filename, ".")
		if i == -1 || !lo.ContainsBy(v.Extensions, func(ext string) bool { return strings.HasSuffix(file.Header.Filename[i:], "."+ext) }) {
			return false
		}
	}
	return true
}

// Name returns the string name of the validator.
func (v *ExtensionValidator) Name() string { return "extension" }

// MessagePlaceholders returns the ":values" placeholder.
func (v *ExtensionValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":values", strings.Join(v.Extensions, ", "),
	}
}

// Extension the field under validation must be a file whose
// filename has one of the specified extensions as suffix.
// Don't include the dot in the extension.
// Composite extensions (e.g. "tar.gz") are supported.
//
// Multi-files are supported (all files must satisfy the criteria).
func Extension(extensions ...string) *ExtensionValidator {
	return &ExtensionValidator{Extensions: extensions}
}
