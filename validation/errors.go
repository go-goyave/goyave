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
	switch path.Type {
	case walk.PathTypeElement:
		errs, ok := e[path.Name]
		if !ok {
			errs = &FieldErrors{}
			e[path.Name] = errs
		}
		// TODO check duplicate messages
		errs.Errors = append(errs.Errors, message)
	case walk.PathTypeArray:
		errs, ok := e[path.Name]
		if !ok {
			errs = &FieldErrors{}
			e[path.Name] = errs
		}
		if errs.Elements == nil {
			errs.Elements = make(map[int]*FieldErrors)
		}
		errs.Elements.Add(path.Next, *path.Index, message)
	case walk.PathTypeObject:
		// TODO factorize this, if ok or not ok is repeated
		errs, ok := e[path.Name]
		if !ok {
			errs = &FieldErrors{}
			e[path.Name] = errs
		}
		if errs.Fields == nil {
			errs.Fields = make(Errors)
		}
		errs.Fields.Add(path.Next, message)
	}
}

func (e ArrayErrors) Add(path *walk.Path, index int, message string) {
	switch path.Type {
	case walk.PathTypeElement:
		elem, ok := e[index]
		if !ok {
			elem = &FieldErrors{}
			e[index] = elem
		}
		// TODO check duplicate messages
		elem.Errors = append(elem.Errors, message)
	case walk.PathTypeArray:
		errs, ok := e[index]
		if !ok {
			errs = &FieldErrors{}
			e[index] = errs
		}
		if errs.Elements == nil {
			errs.Elements = make(map[int]*FieldErrors)
		}
		errs.Elements.Add(path.Next, *path.Index, message)
	case walk.PathTypeObject:
		errs, ok := e[index]
		if !ok {
			errs = &FieldErrors{}
			e[index] = errs
		}
		if errs.Fields == nil {
			errs.Fields = make(Errors)
		}
		errs.Fields.Add(path.Next, message)
	}
}
