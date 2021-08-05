package validation

import (
	"bufio"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

// PathType type of the element being explored.
type PathType int

const (
	// PathTypeElement the explored element is used as a final element (leaf).
	PathTypeElement PathType = iota

	// PathTypeArray the explored element is used as an array and not a final element.
	// All elements in the array will be explored using the next PathItem.
	PathTypeArray

	// PathTypeObject the explored element is used as an object (`map[string]interface{}`)
	// and not a final element.
	PathTypeObject
)

// PathItem step in exploration.
// Items NOT having `PathTypeElement` as a `Type` are expected to have a non-nil `Next`.
type PathItem struct {
	Next *PathItem
	Name string
	Type PathType
}

type WalkContext struct {
	Value  interface{}
	Parent interface{} // Either map[string]interface{} or a slice
	Name   string      // Name of the current element
	Index  int         // If parent is a slice, the index of the current element in the slice, else -1
}

// Walk this path and execute the given behavior for each matching element.
// TODO document behavior function.
func (p *PathItem) Walk(currentElement interface{}, f func(WalkContext)) {
	p.walk(currentElement, nil, -1, f)
}

func (p *PathItem) walk(currentElement interface{}, parent interface{}, index int, f func(WalkContext)) {
	element := currentElement
	if p.Name != "" {
		var ok bool
		element, ok = currentElement.(map[string]interface{})[p.Name]
		if !ok {
			// TODO check this doesn't result in skip behavior for certain rules
			return
		}
		parent = currentElement
	}

	switch p.Type {
	case PathTypeElement:
		f(WalkContext{
			Value:  element,
			Parent: parent,
			Name:   p.Name,
			Index:  index,
		})
	case PathTypeArray:
		list := reflect.ValueOf(element)
		length := list.Len()
		for i := 0; i < length; i++ {
			v := list.Index(i)
			value := v.Interface()
			p.Next.walk(value, element, i, f)
		}
	case PathTypeObject:
		p.Next.walk(element, parent, -1, f)
		// TODO better safety checks
	}
}

// ComputePath transform given path string representation into usable PathItem.
//
// Example paths:
//   name
//   object.field
//   object.subobject.field
//   object.array[]
//   object.arrayOfObjects[].field
func ComputePath(p string) (*PathItem, error) {
	rootPath := &PathItem{}
	path := rootPath

	scanner := createPathScanner(p)
	for scanner.Scan() {
		t := scanner.Text()
		switch t {
		case "[]":
			if path.Type == PathTypeArray {
				path.Next = &PathItem{
					Type: PathTypeArray,
				}
				path = path.Next
			} else {
				path.Type = PathTypeArray
			}
		case ".":
			if path.Type == PathTypeArray {
				path.Next = &PathItem{
					Type: PathTypeObject,
					Next: &PathItem{
						Type: PathTypeElement,
					},
				}
				path = path.Next.Next
			} else {
				path.Type = PathTypeObject
				path.Next = &PathItem{
					Type: PathTypeElement,
				}
				path = path.Next
			}
		default:
			path.Name = t
		}
	}

	if path.Type != PathTypeElement {
		path.Next = &PathItem{
			Type: PathTypeElement,
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rootPath, nil
}

func createPathScanner(path string) *bufio.Scanner {
	scanner := bufio.NewScanner(strings.NewReader(path))
	lastSeparator := rune(0)
	split := func(data []byte, atEOF bool) (int, []byte, error) {
		for width, i := 0, 0; i < len(data); i += width {
			var r rune
			r, width = utf8.DecodeRune(data[i:])
			if r == '.' {
				if lastSeparator == '.' {
					return 0, nil, fmt.Errorf("Consecutive dot")
				}
				lastSeparator = r
				return i + width, data[:i+width], nil
			} else if r == ']' && lastSeparator == '[' {
				lastSeparator = r
				return i + width, data[:i+width], nil
			} else if r == '[' && lastSeparator != '[' {
				lastSeparator = r
				if i != 0 {
					return i, data[:i], nil
				}
			}
		}
		if atEOF && len(data) > 0 {
			var err error
			if lastSeparator == '[' {
				err = fmt.Errorf("Unclosed bracket")
			}
			return len(data), data[:], err
		}
		return 0, nil, nil
	}
	scanner.Split(split)
	return scanner
}
