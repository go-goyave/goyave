package validation

import (
	"goyave.dev/goyave/v5/util/walk"
)

// Errors structure representing the errors associated with an element.
// If the element is an object (`map[string]any`), `Fields` represents the
// errors associated with this object's fields. The key is the name of the field. `Fields` may be `nil`.
// If the element is a slice, `Elements` represents the errors associated with each element
// of the array. See `ArrayErrors` for more details.
type Errors struct {
	Fields   FieldsErrors `json:"fields,omitempty"`
	Elements ArrayErrors  `json:"elements,omitempty"`
	Errors   []string     `json:"errors,omitempty"`
}

// FieldsErrors representing the errors associated with the fields of an object,
// the key being the name of the field.
type FieldsErrors map[string]*Errors

// ArrayErrors representing the errors associated with an element
// of an array. The key is the index of the element in the array, or -1 if the element
// doesn't exist.
type ArrayErrors map[int]*Errors

// Add an error message to the element identified by the given path.
// If a step in the path of type `PathTypeArray` doesn't provide an index,
// -1 will be used to indicate that the element doesn't exist.
// Creates all missing elements in the path.
//
// Note that the walking behavior is slightly different from `walk.Path.Walk`:
// if the first step in the path is of type `walk.PathTypeObject`, it will be
// considered as the root element and skipped. This allows this implementation
// to know the root element is an object and create the `FieldsErrors` accordingly.
func (e *Errors) Add(path *walk.Path, message string) {
	switch path.Type {
	case walk.PathTypeElement:
		e.Errors = append(e.Errors, message)
	case walk.PathTypeArray:
		if e.Elements == nil {
			e.Elements = make(map[int]*Errors)
		}

		index := -1
		if path.Index != nil {
			index = *path.Index
		}
		e.Elements.Add(path.Next, index, message)
	case walk.PathTypeObject:
		if e.Fields == nil {
			e.Fields = make(FieldsErrors)
		}
		e.Fields.Add(path.Next, message)
	}
}

// Merge the given errors into this bag of errors at the given path.
// This can be used when a validator uses nested validation and wants
// to add the results in the higher-level validation errors.
//
// Missing path segments will be added automatically.
// Elements from the given errors are NOT cloned. Therefore there can
// be side-effects if you modify them after the call of `Merge`.
func (e *Errors) Merge(path *walk.Path, errors *Errors) {
	switch path.Type {
	case walk.PathTypeElement:
		if len(errors.Fields) > 0 && e.Fields == nil {
			e.Fields = make(FieldsErrors, len(errors.Fields))
		}
		for k, v := range errors.Fields {
			if fields, ok := e.Fields[k]; ok {
				fields.Merge(path, v)
			} else {
				e.Fields[k] = v
			}
		}
		if len(errors.Elements) > 0 && e.Elements == nil {
			e.Elements = make(ArrayErrors, len(errors.Elements))
		}
		for i, v := range errors.Elements {
			if elements, ok := e.Elements[i]; ok {
				elements.Merge(path, v)
			} else {
				e.Elements[i] = v
			}
		}
		e.Errors = append(e.Errors, errors.Errors...)
	case walk.PathTypeArray:
		if e.Elements == nil {
			e.Elements = make(ArrayErrors)
		}

		index := -1
		if path.Index != nil {
			index = *path.Index
		}
		e.Elements.Merge(path.Next, index, errors)
	case walk.PathTypeObject:
		if e.Fields == nil {
			e.Fields = make(FieldsErrors)
		}
		e.Fields.Merge(path.Next, errors)
	}
}

// Add an error message to the element identified by the given path.
// Creates all missing elements in the path.
func (e FieldsErrors) Add(path *walk.Path, message string) {
	errs, ok := e[*path.Name]
	if !ok {
		errs = &Errors{}
		e[*path.Name] = errs
	}
	errs.Add(path, message)
}

// Merge the given errors into this bag of errors at the given path.
func (e FieldsErrors) Merge(path *walk.Path, errors *Errors) {
	errs, ok := e[*path.Name]
	if !ok {
		errs = &Errors{}
		e[*path.Name] = errs
	}
	errs.Merge(path, errors)
}

// Add an error message to the element identified by the given path in the array,
// at the given index. "-1" index is accepted to identify non-existing elements.
// Creates all missing elements in the path.
func (e ArrayErrors) Add(path *walk.Path, index int, message string) {
	errs, ok := e[index]
	if !ok {
		errs = &Errors{}
		e[index] = errs
	}
	errs.Add(path, message)
}

// Merge the given errors into this bag of errors at the given path.
func (e ArrayErrors) Merge(path *walk.Path, index int, errors *Errors) {
	errs, ok := e[index]
	if !ok {
		errs = &Errors{}
		e[index] = errs
	}
	errs.Merge(path, errors)
}
