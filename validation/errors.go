package validation

import (
	"goyave.dev/goyave/v3/helper/walk"
)

type Errors map[string]*FieldErrors

type ArrayErrors map[int]*FieldErrors

type FieldErrors struct {
	Errors   []string    `json:"errors,omitempty"`
	Fields   Errors      `json:"fields,omitempty"`
	Elements ArrayErrors `json:"elements,omitempty"`
}

func (e Errors) Add(path *walk.Path, message string) {
	errs, ok := e[path.Name]
	if !ok {
		errs = &FieldErrors{}
		e[path.Name] = errs
	}
	errs.Add(path, message)
}

func (e ArrayErrors) Add(path *walk.Path, index int, message string) {
	errs, ok := e[index]
	if !ok {
		errs = &FieldErrors{}
		e[index] = errs
	}
	errs.Add(path, message)
}

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
