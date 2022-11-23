package validation

import (
	"goyave.dev/goyave/v4/util/walk"
)

type FieldV5 struct {
	Path       *walk.Path
	Elements   *FieldV5
	Validators []Validator

	// prefixDepth When using composition, `prefixDepth` allows to truncate the path to the
	// validated element in order to retrieve the root object or array relative to
	// the composed RuleSet.
	prefixDepth uint

	isArray    bool
	isObject   bool
	isRequired func(*ContextV5) bool
	isNullable bool
}

func newField(path string, validators []Validator, prefixDepth uint) *FieldV5 {
	p, err := walk.Parse(path)
	if err != nil {
		panic(err)
	}

	f := &FieldV5{
		Path:        p,
		Validators:  validators,
		prefixDepth: prefixDepth,
	}

	for _, v := range validators {
		switch v.(type) {
		case *RequiredValidator:
			f.isRequired = func(ctx *ContextV5) bool { return true }
		case *RequiredIfValidator:
			f.isRequired = v.(*RequiredIfValidator).Condition
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

func (f *FieldV5) getErrorPath(parentPath *walk.Path, c *walk.Context) *walk.Path {
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
func (f *FieldV5) IsRequired(ctx *ContextV5) bool {
	return f.isRequired != nil && f.isRequired(ctx)
}

// IsNullable check if a field has the "nullable" rule
func (f *FieldV5) IsNullable() bool {
	return f.isNullable
}

// IsArray check if a field has the "array" rule
func (f *FieldV5) IsArray() bool {
	return f.isArray
}

// IsObject check if a field has the "object" rule
func (f *FieldV5) IsObject() bool {
	return f.isObject
}

// PrefixDepth When using composition, `prefixDepth` allows to truncate the path to the
// validated element in order to retrieve the root object or array relative to
// the composed RuleSet.
func (f *FieldV5) PrefixDepth() uint {
	return f.prefixDepth
}
