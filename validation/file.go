package validation

import (
	"strconv"

	"goyave.dev/goyave/v5/util/fsutil"
)

// FileValidator validates the field under validation must be a file.
// Multi-files are supported.
type FileValidator struct {
	BaseValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *FileValidator) Validate(ctx *Context) bool {
	_, ok := ctx.Value.([]fsutil.File)
	return ok
}

// Name returns the string name of the validator.
func (v *FileValidator) Name() string { return "file" }

// IsType returns true.
func (v *FileValidator) IsType() bool { return true }

// File the field under validation must be a file. Multi-files are supported.
func File() *FileValidator {
	return &FileValidator{}
}

//------------------------------

// FileCountValidator validates the field under validation must be a multi-files
// with exactly the specified number of files.
type FileCountValidator struct {
	BaseValidator
	Count uint
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *FileCountValidator) Validate(ctx *Context) bool {
	files, ok := ctx.Value.([]fsutil.File)
	return ok && uint(len(files)) == v.Count
}

// Name returns the string name of the validator.
func (v *FileCountValidator) Name() string { return "file_count" }

// MessagePlaceholders returns the ":value" placeholder.
func (v *FileCountValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":value", strconv.FormatUint(uint64(v.Count), 10),
	}
}

// FileCount the field under validation must be a multi-files
// with exactly the specified number of files.
func FileCount(count uint) *FileCountValidator {
	return &FileCountValidator{Count: count}
}

//------------------------------

// MinFileCountValidator validates the field under validation must be a multi-files
// with at least the specified number of files.
type MinFileCountValidator struct {
	BaseValidator
	Min uint
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *MinFileCountValidator) Validate(ctx *Context) bool {
	files, ok := ctx.Value.([]fsutil.File)
	return ok && uint(len(files)) >= v.Min
}

// Name returns the string name of the validator.
func (v *MinFileCountValidator) Name() string { return "min_file_count" }

// MessagePlaceholders returns the ":min" placeholder.
func (v *MinFileCountValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":min", strconv.FormatUint(uint64(v.Min), 10),
	}
}

// MinFileCount the field under validation must be a multi-files
// with at least the specified number of files.
func MinFileCount(min uint) *MinFileCountValidator {
	return &MinFileCountValidator{Min: min}
}

//------------------------------

// MaxFileCountValidator validates the field under validation must be a multi-files
// with at most the specified number of files.
type MaxFileCountValidator struct {
	BaseValidator
	Max uint
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *MaxFileCountValidator) Validate(ctx *Context) bool {
	files, ok := ctx.Value.([]fsutil.File)
	return ok && uint(len(files)) <= v.Max
}

// Name returns the string name of the validator.
func (v *MaxFileCountValidator) Name() string { return "max_file_count" }

// MessagePlaceholders returns the ":max" placeholder.
func (v *MaxFileCountValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":max", strconv.FormatUint(uint64(v.Max), 10),
	}
}

// MaxFileCount the field under validation must be a multi-files
// with at most the specified number of files.
func MaxFileCount(max uint) *MaxFileCountValidator {
	return &MaxFileCountValidator{Max: max}
}

//------------------------------

// FileCountBetweenValidator validates the field under validation must be a multi-files
// with a number of files between the specified min and max.
type FileCountBetweenValidator struct {
	BaseValidator
	Min uint
	Max uint
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *FileCountBetweenValidator) Validate(ctx *Context) bool {
	files, ok := ctx.Value.([]fsutil.File)
	return ok && uint(len(files)) >= v.Min && uint(len(files)) <= v.Max
}

// Name returns the string name of the validator.
func (v *FileCountBetweenValidator) Name() string { return "file_count_between" }

// MessagePlaceholders returns the ":min" and ":max" placeholders.
func (v *FileCountBetweenValidator) MessagePlaceholders(_ *Context) []string {
	return []string{
		":min", strconv.FormatUint(uint64(v.Min), 10),
		":max", strconv.FormatUint(uint64(v.Max), 10),
	}
}

// FileCountBetween the field under validation must be a multi-files
// with a number of files between the specified min and max.
func FileCountBetween(min, max uint) *FileCountBetweenValidator {
	return &FileCountBetweenValidator{Min: min, Max: max}
}
