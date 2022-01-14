package validation

import (
	"goyave.dev/goyave/v4/util/walk"
)

// Errors structure representing errors associated with an element of the validated data.
// The element may represent the root object or the fields of a nested object.
// The key is the name of the field.
type Errors map[string]*FieldErrors

// ArrayErrors structure representing the errors associated with an element
// of an array. The key is the index of the element in the array, or -1 if the element
// doesn't exist.
type ArrayErrors map[int]*FieldErrors

// FieldErrors structure representing the errors associated with an element.
// If the element is an object (`map[string]interface{}`), `Fields` represents the
// errors associated with this object's fields. The key is the name of the field. `Fields` may be `nil`.
// If the element is a slice, `Elements` represents the errors associated with each element
// of the array. See `ArrayErrors` for more details.
type FieldErrors struct {
	Fields   Errors      `json:"fields,omitempty"`
	Elements ArrayErrors `json:"elements,omitempty"`
	Errors   []string    `json:"errors,omitempty"`
}

// Add an error message to the element identified by the given path.
// Creates all missing elements in the path.
func (e Errors) Add(path *walk.Path, message string) {
	errs, ok := e[path.Name]
	if !ok {
		errs = &FieldErrors{}
		e[path.Name] = errs
	}
	errs.Add(path, message)
}

// Add an error message to the element identified by the given path in the array,
// at the given index. "-1" index is accepted to identify non-existing elements.
// Creates all missing elements in the path.
func (e ArrayErrors) Add(path *walk.Path, index int, message string) {
	errs, ok := e[index]
	if !ok {
		errs = &FieldErrors{}
		e[index] = errs
	}
	errs.Add(path, message)
}

// Add an error message to the element identified by the given path.
// If a step in the path of type `PathTypeArray` doesn't provide an index,
// -1 will be used to indicate that the element doesn't exist.
// Creates all missing elements in the path.
func (e *FieldErrors) Add(path *walk.Path, message string) {
	switch path.Type {
	case walk.PathTypeElement:
		e.Errors = append(e.Errors, message)
	case walk.PathTypeArray:
		if e.Elements == nil {
			e.Elements = make(map[int]*FieldErrors)
		}

		index := -1
		if path.Index != nil {
			index = *path.Index
		}
		e.Elements.Add(path.Next, index, message)
	case walk.PathTypeObject:
		if e.Fields == nil {
			e.Fields = make(Errors)
		}
		e.Fields.Add(path.Next, message)
	}
}
