package validation

import (
	"goyave.dev/goyave/v4/util/fsutil"
)

// FileValidator validates the field under validation must be a file.
// Multi-files are supported.
type FileValidator struct {
	BaseValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *FileValidator) Validate(ctx *ContextV5) bool {
	_, ok := ctx.Value.([]fsutil.File)
	return ok
}

// Name returns the string name of the validator.
func (v *FileValidator) Name() string { return "file" }

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
func (v *FileCountValidator) Validate(ctx *ContextV5) bool {
	files, ok := ctx.Value.([]fsutil.File)
	return ok && uint(len(files)) == v.Count
}

// Name returns the string name of the validator.
func (v *FileCountValidator) Name() string { return "file_count" }

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
func (v *MinFileCountValidator) Validate(ctx *ContextV5) bool {
	files, ok := ctx.Value.([]fsutil.File)
	return ok && uint(len(files)) >= v.Min
}

// Name returns the string name of the validator.
func (v *MinFileCountValidator) Name() string { return "min_file_count" }

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
func (v *MaxFileCountValidator) Validate(ctx *ContextV5) bool {
	files, ok := ctx.Value.([]fsutil.File)
	return ok && uint(len(files)) <= v.Max
}

// Name returns the string name of the validator.
func (v *MaxFileCountValidator) Name() string { return "max_file_count" }

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
func (v *FileCountBetweenValidator) Validate(ctx *ContextV5) bool {
	files, ok := ctx.Value.([]fsutil.File)
	return ok && uint(len(files)) >= v.Min && uint(len(files)) <= v.Max
}

// Name returns the string name of the validator.
func (v *FileCountBetweenValidator) Name() string { return "file_count_between" }

// FileCountBetween the field under validation must be a multi-files
// with a number of files between the specified min and max.
func FileCountBetween(min, max uint) *FileCountBetweenValidator {
	return &FileCountBetweenValidator{Min: min, Max: max}
}
