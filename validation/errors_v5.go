package validation

import "goyave.dev/goyave/v4/util/walk"

// Errors structure representing the errors associated with an element.
// If the element is an object (`map[string]any`), `Fields` represents the
// errors associated with this object's fields. The key is the name of the field. `Fields` may be `nil`.
// If the element is a slice, `Elements` represents the errors associated with each element
// of the array. See `ArrayErrors` for more details.
type ErrorsV5 struct {
	Fields   FieldsErrors  `json:"fields,omitempty"`
	Elements ArrayErrorsV5 `json:"elements,omitempty"`
	Errors   []string      `json:"errors,omitempty"`
}

// FieldsErrors representing the errors associated with the fields of an object,
// the key being the name of the field.
type FieldsErrors map[string]*ErrorsV5

// ArrayErrors representing the errors associated with an element
// of an array. The key is the index of the element in the array, or -1 if the element
// doesn't exist.
type ArrayErrorsV5 map[int]*ErrorsV5

// Add an error message to the element identified by the given path.
// If a step in the path of type `PathTypeArray` doesn't provide an index,
// -1 will be used to indicate that the element doesn't exist.
// Creates all missing elements in the path.
func (e *ErrorsV5) Add(path *walk.Path, message string) {
	switch path.Type {
	case walk.PathTypeElement:
		e.Errors = append(e.Errors, message)
	case walk.PathTypeArray:
		if e.Elements == nil {
			e.Elements = make(map[int]*ErrorsV5)
		}

		index := -1
		if path.Index != nil {
			index = *path.Index
		}
		e.Elements.Add(path.Next, index, message)
	case walk.PathTypeObject:
		if e.Fields == nil {
			e.Fields = make(map[string]*ErrorsV5)
		}
		e.Fields.Add(path.Next, message)
	}
}

// Add an error message to the element identified by the given path.
// Creates all missing elements in the path.
func (e FieldsErrors) Add(path *walk.Path, message string) {
	errs, ok := e[*path.Name]
	if !ok {
		errs = &ErrorsV5{}
		e[*path.Name] = errs
	}
	errs.Add(path, message)
}

// Add an error message to the element identified by the given path in the array,
// at the given index. "-1" index is accepted to identify non-existing elements.
// Creates all missing elements in the path.
func (e ArrayErrorsV5) Add(path *walk.Path, index int, message string) {
	errs, ok := e[index]
	if !ok {
		errs = &ErrorsV5{}
		e[index] = errs
	}
	errs.Add(path, message)
}
