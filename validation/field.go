package validation

import (
	"goyave.dev/goyave/v5/util/walk"
)

// Field representation of a single field in the data being validated.
// Provides useful information based on its validators (if required, nullable, etc).
type Field struct {
	isRequired func(*Context) bool

	Path       *walk.Path
	Elements   *Field
	Validators []Validator

	// prefixDepth When using composition, `prefixDepth` allows to truncate the path to the
	// validated element in order to retrieve the root object or array relative to
	// the composed RuleSet.
	prefixDepth uint

	isArray    bool
	isObject   bool
	isNullable bool
}

func alwaysRequired(_ *Context) bool { return true }

func newField(path string, validators []Validator, prefixDepth uint) *Field {
	p := walk.MustParse(path)
	f := &Field{
		Path:        p,
		Validators:  validators,
		prefixDepth: prefixDepth,
	}

	for _, v := range validators {
		switch v := v.(type) {
		case *RequiredValidator:
			f.isRequired = alwaysRequired
		case *RequiredIfValidator:
			f.isRequired = v.Condition
		case *NullableValidator:
			f.isNullable = true
		case *ArrayValidator:
			f.isArray = true
		case *ObjectValidator:
			f.isObject = true
		}
	}

	return f
}

// getErrorPath returns the path to use when appending the error message to the
// final validation errors.
//
// The given `parentPath` corresponds to the path to the parent array
// if the parent is an array, otherwise `nil`. If `nil`, returns the unmodified
// path from the `walk.Context`.
func (f *Field) getErrorPath(parentPath *walk.Path, c *walk.Context) *walk.Path {
	if parentPath != nil {
		clone := parentPath.Clone()
		tail := clone.Tail()
		tail.Type = walk.PathTypeArray
		tail.Index = &c.Index
		tail.Next = &walk.Path{Type: walk.PathTypeElement}
		return clone
	}

	return c.Path
}

// IsRequired check if a field has the "required" rule
func (f *Field) IsRequired(ctx *Context) bool {
	return f.isRequired != nil && f.isRequired(ctx)
}

// IsNullable check if a field has the "nullable" rule
func (f *Field) IsNullable() bool {
	return f.isNullable
}

// IsArray check if a field has the "array" rule
func (f *Field) IsArray() bool {
	return f.isArray
}

// IsObject check if a field has the "object" rule
func (f *Field) IsObject() bool {
	return f.isObject
}

// PrefixDepth When using composition, `prefixDepth` allows to truncate the path to the
// validated element in order to retrieve the root object or array relative to
// the composed RuleSet.
func (f *Field) PrefixDepth() uint {
	return f.prefixDepth
}
