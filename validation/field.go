package validation

import (
	"fmt"

	"goyave.dev/goyave/v4/util/walk"
)

// Field is a component of route validation. A Field is a value in
// a Rules map, the key being the name of the field.
type Field struct {
	Path     *walk.Path
	Elements *Field // If the field is an array, the field representing its elements, or nil
	// Maybe use the same concept for objects too?
	Rules      []*Rule
	isArray    bool
	isObject   bool
	isRequired bool
	isNullable bool
}

// IsRequired check if a field has the "required" rule
func (f *Field) IsRequired() bool {
	return f.isRequired
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

// Check if rules meet the minimum parameters requirement and update
// the isRequired, isNullable and isArray fields.
func (f *Field) Check() {
	for _, rule := range f.Rules {
		switch rule.Name {
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

		def, exists := validationRules[rule.Name]
		if !exists {
			panic(fmt.Sprintf("Rule \"%s\" doesn't exist", rule.Name))
		}
		if len(rule.Params) < def.RequiredParameters {
			panic(fmt.Sprintf("Rule \"%s\" requires %d parameter(s)", rule.Name, def.RequiredParameters))
		}
	}
}

func (f *Field) getErrorPath(parentPath *walk.Path, c walk.Context) *walk.Path {
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

func (f *Field) apply(fieldMap FieldMap, name string) {
	fieldMap[name] = f
}
