package validation

import (
	"strings"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5/util/fsutil"
)

// MIMEValidator validates the field under validation must be a file
// and match one of the given MIME types.
// Multi-files are supported (all files must satisfy the criteria).
type MIMEValidator struct {
	BaseValidator
	MIMETypes []string
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *MIMEValidator) Validate(ctx *Context) bool {
	files, ok := ctx.Value.([]fsutil.File)
	if ok {
		for _, file := range files {
			mime := file.MIMEType
			if i := strings.Index(mime, ";"); i != -1 { // Ignore MIME settings (example: "text/plain; charset=utf-8")
				mime = mime[:i]
			}
			if !lo.Contains(v.MIMETypes, mime) {
				return false
			}
		}
		return true
	}
	return false
}

// Name returns the string name of the validator.
func (v *MIMEValidator) Name() string { return "mime" }

// MessagePlaceholders returns the ":values" placeholder.
func (v *MIMEValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":values", strings.Join(v.MIMETypes, ", "),
	}
}

// MIME the field under validation must be a file and match one of the given
// MIME types. Multi-files are supported (all files must satisfy the criteria).
func MIME(mimeTypes ...string) *MIMEValidator {
	return &MIMEValidator{MIMETypes: mimeTypes}
}

//------------------------------

// ImageMIMETypes MIME types accepted by `ImageValidator`.
var ImageMIMETypes = []string{"image/jpeg", "image/png", "image/gif", "image/bmp", "image/svg+xml", "image/webp"}

// ImageValidator validates the field under validation must be an image file.
// Multi-files are supported (all files must satisfy the criteria).
type ImageValidator struct {
	MIMEValidator
}

// Name returns the string name of the validator.
func (v *ImageValidator) Name() string { return "image" }

// Image the field under validation must be an image file.
// Multi-files are supported (all files must satisfy the criteria).
//
// Accepted MIME types are defined by `ImageMIMETypes`.
func Image() *ImageValidator {
	return &ImageValidator{MIMEValidator: MIMEValidator{MIMETypes: ImageMIMETypes}}
}
