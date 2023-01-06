package validation

import (
	"math"
	"reflect"
	"strconv"

	"github.com/Code-Hex/uniseg"
	"goyave.dev/goyave/v4/util/fsutil"
)

func validateSizeV5(value any, v func(size int) bool) bool {
	val := reflect.ValueOf(value)
	switch getFieldType(val) {
	case FieldTypeString:
		return v(uniseg.GraphemeClusterCount(value.(string)))
	case FieldTypeArray, FieldTypeObject: // TODO document it also works for objects (number of keys)
		return v(val.Len())
	case FieldTypeFile:
		files, _ := value.([]fsutil.File)
		for _, file := range files {
			if !v(int(math.Ceil(float64(file.Header.Size) / 1024.0))) {
				return false
			}
		}
	}
	return true // Pass if field type cannot be checked (bool, dates, ...)
}

// SizeValidator validates the field under validation depending on its type.
//   - Strings must have a length of n characters (calculated based on the number of grapheme clusters)
//   - Arrays must have n elements
//   - Objects must have n keys
//   - Files must weight n KiB (for multi-files, all files must match this criteria). The number of KiB is rounded up (ceil).
type SizeValidator struct {
	BaseValidator
	Size int
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *SizeValidator) Validate(ctx *ContextV5) bool {
	return validateSizeV5(ctx.Value, func(size int) bool {
		return size == v.Size
	})
}

// Name returns the string name of the validator.
func (v *SizeValidator) Name() string { return "size" }

// IsTypeDependent returns true
func (v *SizeValidator) IsTypeDependent() bool { return true }

// MessagePlaceholders returns the ":value" placeholder.
func (v *SizeValidator) MessagePlaceholders(ctx *ContextV5) []string {
	return []string{
		":value", strconv.Itoa(v.Size),
	}
}

// Size validates the field under validation depending on its type.
//   - Strings must have a length of n characters (calculated based on the number of grapheme clusters)
//   - Arrays must have n elements
//   - Objects must have n keys
//   - Files must weight n KiB (for multi-files, all files must match this criteria). The number of KiB is rounded up (ceil).
func Size(size int) *SizeValidator {
	return &SizeValidator{Size: size}
}