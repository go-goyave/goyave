package walk

import (
	"bufio"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

// PathType type of the element being explored.
type PathType int

// FoundType adds extra information about not found elements whether
// what's not found is their parent or themselves.
type FoundType int

const (
	// PathTypeElement the explored element is used as a final element (leaf).
	PathTypeElement PathType = iota

	// PathTypeArray the explored element is used as an array and not a final element.
	// All elements in the array will be explored using the next Path.
	PathTypeArray

	// PathTypeObject the explored element is used as an object (`map[string]interface{}`)
	// and not a final element.
	PathTypeObject
)

const (
	// Found indicates the element could be found.
	Found FoundType = iota
	// ParentNotFound indicates one of the parents of the element could no be found.
	ParentNotFound
	// ElementNotFound indicates all parents of the element were found but the element
	// itself could not.
	ElementNotFound
)

// Path allows for complex untyped data structure exploration.
// An instance of this structure represents a step in exploration.
// Items NOT having `PathTypeElement` as a `Type` are expected to have a non-nil `Next`.
type Path struct {
	Next  *Path
	Index *int
	Name  string
	Type  PathType
}

// Context information sent to walk function.
type Context struct {
	Value  interface{}
	Parent interface{} // Either map[string]interface{} or a slice
	Path   *Path       // Exact Path to the current element
	Name   string      // Name of the current element
	Index  int         // If parent is a slice, the index of the current element in the slice, else -1
	Found  FoundType   // True if the path could not be completely explored
}

// Walk this path and execute the given behavior for each matching element. Elements are final,
// meaning they are the deepest explorable element using this path.
// Only `map[string]interface{}` and n-dimensional slices parents are supported.
// The given "f" function is executed for each final element matched. If the path
// cannot be completed because the step's name doesn't exist in the currently explored map,
// the function will be executed as well, with a the `Context`'s `NotFound` field set to `true`.
func (p *Path) Walk(currentElement interface{}, f func(Context)) {
	path := &Path{
		Name: p.Name,
		Type: p.Type,
	}
	p.walk(currentElement, nil, -1, path, path, f)
}

func (p *Path) walk(currentElement interface{}, parent interface{}, index int, path *Path, lastPathElement *Path, f func(Context)) {
	element := currentElement
	if p.Name != "" {
		ce, ok := currentElement.(map[string]interface{})
		found := ParentNotFound
		if ok {
			element, ok = ce[p.Name]
			if !ok && p.Type == PathTypeElement {
				found = ElementNotFound
			}
			index = -1
		}
		if !ok {
			p.completePath(lastPathElement)
			f(newNotFoundContext(currentElement, path, p.Name, index, found))
			return
		}
		parent = currentElement
	}

	switch p.Type {
	case PathTypeElement:
		f(Context{
			Value:  element,
			Parent: parent,
			Path:   path,
			Name:   p.Name,
			Index:  index,
		})
	case PathTypeArray:
		list := reflect.ValueOf(element)
		if list.Kind() != reflect.Slice {
			lastPathElement.Type = PathTypeElement
			f(newNotFoundContext(parent, path, p.Name, index, ParentNotFound))
			return
		}
		length := list.Len()
		if p.Index != nil {
			lastPathElement.Index = p.Index
			lastPathElement.Next = &Path{Name: p.Next.Name, Type: p.Next.Type}
			if p.outOfBounds(length) {
				f(newNotFoundContext(element, path, "", *p.Index, ElementNotFound))
				return
			}
			v := list.Index(*p.Index)
			value := v.Interface()
			p.Next.walk(value, element, *p.Index, path, lastPathElement.Next, f)
			return
		}
		if length == 0 {
			lastPathElement.Next = &Path{Name: p.Next.Name, Type: PathTypeElement}
			found := ElementNotFound
			if p.Next.Type != PathTypeElement {
				found = ParentNotFound
			}
			f(newNotFoundContext(element, path, "", -1, found))
			return
		}
		for i := 0; i < length; i++ {
			j := i
			clone := path.Clone()
			tail := clone.Tail()
			tail.Index = &j
			tail.Next = &Path{Name: p.Next.Name, Type: p.Next.Type}
			v := list.Index(i)
			value := v.Interface()
			p.Next.walk(value, element, i, clone, tail.Next, f)
		}
	case PathTypeObject:
		lastPathElement.Next = &Path{Name: p.Next.Name, Type: p.Next.Type}
		p.Next.walk(element, parent, index, path, lastPathElement.Next, f)
	}
}

func (p *Path) outOfBounds(length int) bool {
	return *p.Index >= length || *p.Index < 0
}

func (p *Path) completePath(lastPathElement *Path) {
	completedPath := lastPathElement
	if p.Type == PathTypeArray {
		i := -1
		completedPath.Index = &i
	}
	if p.Type != PathTypeElement {
		completedPath.Next = p.Next.Clone()
		completedPath.Next.setAllMissingIndexes()
	}
}

func newNotFoundContext(parent interface{}, path *Path, name string, index int, found FoundType) Context {
	return Context{
		Value:  nil,
		Parent: parent,
		Path:   path,
		Name:   name,
		Index:  index,
		Found:  found,
	}
}

// HasArray returns true if a least one step in the path involves an array.
func (p *Path) HasArray() bool {
	step := p
	for step != nil {
		if step.Type == PathTypeArray {
			return true
		}
		step = step.Next
	}
	return false
}

// LastParent returns the last step in the path that is not a PathTypeElement, excluding
// the first step in the path, or nil.
func (p *Path) LastParent() *Path {
	step := p
	for step != nil {
		if step.Next != nil && step.Next.Type == PathTypeElement {
			return step
		}
		step = step.Next
	}
	return nil
}

// Tail returns the last step in the path.
func (p *Path) Tail() *Path {
	step := p
	for step.Next != nil {
		step = step.Next
	}
	return step
}

// Clone returns a deep clone of this Path.
func (p *Path) Clone() *Path {
	clone := &Path{
		Name:  p.Name,
		Type:  p.Type,
		Index: p.Index,
	}
	if p.Next != nil {
		clone.Next = p.Next.Clone()
	}

	return clone
}

// setAllMissingIndexes set Index to -1 for all `PathTypeArray` steps in this path.
func (p *Path) setAllMissingIndexes() {
	i := -1
	for step := p; step != nil; step = step.Next {
		if step.Type == PathTypeArray {
			step.Index = &i
		}
	}
}

// Parse transform given path string representation into usable Path.
//
// Example paths:
//
//	name
//	object.field
//	object.subobject.field
//	object.array[]
//	object.arrayOfObjects[].field
func Parse(p string) (*Path, error) {
	rootPath := &Path{}
	path := rootPath

	scanner := createPathScanner(p)
	for scanner.Scan() {
		t := scanner.Text()
		switch t {
		case "[]":
			if path.Type == PathTypeArray {
				path.Next = &Path{
					Type: PathTypeArray,
				}
				path = path.Next
			} else {
				path.Type = PathTypeArray
			}
		case ".":
			if path.Type == PathTypeArray {
				path.Next = &Path{
					Type: PathTypeObject,
					Next: &Path{
						Type: PathTypeElement,
					},
				}
				path = path.Next.Next
			} else {
				path.Type = PathTypeObject
				path.Next = &Path{
					Type: PathTypeElement,
				}
				path = path.Next
			}
		default:
			path.Name = t
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if path.Type != PathTypeElement {
		path.Next = &Path{
			Type: PathTypeElement,
		}
	}

	return rootPath, nil
}

func createPathScanner(path string) *bufio.Scanner {
	scanner := bufio.NewScanner(strings.NewReader(path))
	split := func(data []byte, atEOF bool) (int, []byte, error) {
		if len(path) == 0 || path[0] == '.' {
			return len(data), data[:], fmt.Errorf("Illegal syntax: %q", path)
		}
		for width, i := 0, 0; i < len(data); i += width {
			var r rune
			r, width = utf8.DecodeRune(data[i:])

			if i+width < len(data) {
				next, _ := utf8.DecodeRune(data[i+width:])
				if isValidSyntax(r, next) {
					return len(data), data[:], fmt.Errorf("Illegal syntax: %q", path)
				}

				if r == '.' && i == 0 {
					return i + width, data[:i+width], nil
				} else if next == '.' || next == '[' {
					return i + width, data[:i+width], nil
				}
			} else if r == '.' || r == '[' {
				return len(data), data[:], fmt.Errorf("Illegal syntax: %q", path)
			}
		}
		if atEOF && len(data) > 0 {
			return len(data), data[:], nil
		}
		return 0, nil, nil
	}
	scanner.Split(split)
	return scanner
}

func isValidSyntax(r rune, next rune) bool {
	return (r == '.' && next == '.') ||
		(r == '[' && next != ']') ||
		(r == '.' && (next == ']' || next == '[')) ||
		(r != '.' && r != '[' && next == ']') ||
		(r == ']' && next != '[' && next != '.')
}
