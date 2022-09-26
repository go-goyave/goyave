package validation

import (
	"fmt"

	"goyave.dev/goyave/v4/util/walk"
)

// Field is a component of route validation. A Field is a value in
// a Rules map, the key being the name of the field.
type FieldV5 struct {
	Path     *walk.Path
	Elements *FieldV5 // If the field is an array, the field representing its elements, or nil
	// Maybe use the same concept for objects too?
	Rules       []Validator
	prefixDepth uint

	isArray    bool
	isObject   bool
	isRequired bool
	isNullable bool
}

// IsRequired check if a field has the "required" rule
func (f *FieldV5) IsRequired() bool {
	return f.isRequired
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

// Check if rules meet the minimum parameters requirement and update
// the isRequired, isNullable and isArray fields.
func (f *FieldV5) Check() {
	for _, rule := range f.Rules {
		switch rule.Name() { // TODO use type-assert instead (switch type)
		case "file", "mime", "image", "extension", "count",
			"count_min", "count_max", "count_between":
			if f.Path.HasArray() {
				panic(fmt.Sprintf("Cannot use rule \"%s\" in array validation", rule.Name))
			}
		case "required":
			f.isRequired = true
		case "nullable":
			f.isNullable = true
			continue
		case "array":
			f.isArray = true
		case "object":
			f.isObject = true
		}
	}
}

func (f *FieldV5) getErrorPath(parentPath *walk.Path, c walk.Context) *walk.Path {
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
