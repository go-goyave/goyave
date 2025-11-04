package walk

import (
	"bufio"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5/util/errors"
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

	// PathTypeObject the explored element is used as an object (`map[string]any`)
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

var wildcard = lo.ToPtr("*")

var (
	// EscapeChars the list of characters that can be escaped using a backslash `\` when
	// parsing a Path. This map is read-only.
	EscapeChars = map[rune]struct{}{
		'*':  {},
		'[':  {},
		']':  {},
		'.':  {},
		'\\': {},
	}

	escapeRemover = strings.NewReplacer(
		`\*`, `*`,
		`\[`, `[`,
		`\]`, `]`,
		`\.`, `.`,
		`\\`, `\`,
	)

	escapeReplacer = strings.NewReplacer(
		`*`, `\*`,
		`[`, `\[`,
		`]`, `\]`,
		`.`, `\.`,
		`\`, `\\`,
	)
)

// Path allows for complex untyped data structure exploration.
// An instance of this structure represents a step in exploration.
// Items NOT having `PathTypeElement` as a `Type` are expected to have a non-nil `Next`.
type Path struct {
	Next  *Path
	Index *int
	Name  *string
	Type  PathType
}

// Context information sent to walk function.
type Context struct {
	Value  any
	Parent any       // Either map[string]any or a slice
	Path   *Path     // Exact Path to the current element
	Name   string    // Name of the current element
	Index  int       // If parent is a slice, the index of the current element in the slice, else -1
	Found  FoundType // True if the path could not be completely explored

	stop bool
}

// Break when called, indicates the path walker to stop.
// This means the current call of the callback function will be the last.
func (c *Context) Break() {
	c.stop = true
}

// Walk this path and execute the given callback for each matching element. Elements are final,
// meaning they are the deepest explorable element using this path.
// Only `map[string]any` and n-dimensional slices parents are supported.
// The given "f" function is executed for each final element matched. If the path
// cannot be completed because the step's name doesn't exist in the currently explored map,
// the function will be executed as well, with a the `Context`'s `NotFound` field set to `true`.
func (p *Path) Walk(currentElement any, f func(*Context)) {
	trackPath := &Path{
		Name: p.Name,
		Type: p.Type,
	}
	p.walk(currentElement, nil, -1, trackPath, trackPath, f)
}

func (p *Path) walk(currentElement any, parent any, index int, trackPath *Path, lastPathElement *Path, f func(*Context)) bool {
	element := currentElement
	if p.Name != nil {
		ce, ok := currentElement.(map[string]any)
		notFoundType := ParentNotFound
		if ok {
			if len(ce) != 0 && p.Name == wildcard {
				for k := range ce {
					key := k
					trackClone := trackPath.Clone()
					tail := trackClone.Tail()
					tail.Name = &key

					clone := p.Clone()
					clone.Name = &key
					if !clone.walk(element, parent, -1, trackClone, tail, f) {
						return false
					}
				}
				return true
			}

			element, ok = ce[*p.Name]
			if !ok && p.Type == PathTypeElement {
				notFoundType = ElementNotFound
			}
			index = -1
		}
		if !ok {
			p.completePath(lastPathElement)
			ctx := newNotFoundContext(currentElement, trackPath, p.Name, index, notFoundType)
			f(ctx)
			return !ctx.stop
		}
		parent = currentElement
	}

	stop := false
	switch p.Type {
	case PathTypeElement:
		c := &Context{
			Value:  element,
			Parent: parent,
			Path:   trackPath.Clone(),
			Index:  index,
		}
		if p.Name != nil {
			c.Name = *p.Name
		}
		f(c)
		stop = c.stop
	case PathTypeArray:
		stop = !p.walkArray(element, parent, index, trackPath, lastPathElement, f)
	case PathTypeObject:
		lastPathElement.Next = &Path{Name: p.Next.Name, Type: p.Next.Type}
		stop = !p.Next.walk(element, parent, index, trackPath, lastPathElement.Next, f)
	}
	return !stop
}

func (p *Path) walkArray(element any, parent any, index int, trackPath *Path, lastPathElement *Path, f func(*Context)) bool {
	list := reflect.ValueOf(element)
	if list.Kind() != reflect.Slice {
		lastPathElement.Type = PathTypeElement
		ctx := newNotFoundContext(parent, trackPath, p.Name, index, ParentNotFound)
		f(ctx)
		return !ctx.stop
	}
	length := list.Len()
	if p.Index != nil {
		lastPathElement.Index = p.Index
		lastPathElement.Next = &Path{Name: p.Next.Name, Type: p.Next.Type}
		if p.outOfBounds(length) {
			ctx := newNotFoundContext(element, trackPath, nil, *p.Index, ElementNotFound)
			f(ctx)
			return !ctx.stop
		}
		v := list.Index(*p.Index)
		value := v.Interface()
		return p.Next.walk(value, element, *p.Index, trackPath, lastPathElement.Next, f)
	}
	if length == 0 {
		lastPathElement.Next = &Path{Name: p.Next.Name, Type: PathTypeElement}
		notFoundType := ElementNotFound
		if p.Next.Type != PathTypeElement {
			notFoundType = ParentNotFound
		}
		ctx := newNotFoundContext(element, trackPath, nil, -1, notFoundType)
		f(ctx)
		return !ctx.stop
	}
	for i := range length {
		j := i
		trackClone := trackPath.Clone()
		tail := trackClone.Tail()
		tail.Index = &j
		tail.Next = &Path{Name: p.Next.Name, Type: p.Next.Type}
		v := list.Index(i)
		value := v.Interface()
		if !p.Next.walk(value, element, i, trackClone, tail.Next, f) {
			return false
		}
	}
	return true
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

func newNotFoundContext(parent any, path *Path, name *string, index int, found FoundType) *Context {
	c := &Context{
		Value:  nil,
		Parent: parent,
		Path:   path.Clone(),
		Index:  index,
		Found:  found,
		stop:   false,
	}
	if name != nil {
		c.Name = *name
	}
	return c
}

// First returns the first final element matched by the Path.
// Note that the returned Context may indicate that the value could
// not be found, so you should always check `Context.Found` before using
// `Context.Value`.
//
// Bear in mind that map iteration order is not guaranteed. Using paths containing
// wildcards `*` will not always yield the same result.
func (p *Path) First(currentElement any) *Context {
	var result *Context
	p.Walk(currentElement, func(ctx *Context) {
		result = ctx
		ctx.Break()
	})
	return result
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

// Depth returns the depth of the path. For each step in the path, increments the depth by one.
func (p *Path) Depth() uint {
	depth := uint(1)
	step := p
	for step.Next != nil {
		step = step.Next
		depth++
	}
	return depth
}

// Truncate returns a clone of the n first steps of the path so the returned path's depth
// equals the given depth.
func (p *Path) Truncate(depth uint) *Path {
	if depth == 0 {
		return nil
	}
	if depth == 1 {
		return &Path{
			Name:  p.Name,
			Type:  PathTypeElement,
			Index: p.Index,
		}
	}
	clone := &Path{
		Name:  p.Name,
		Type:  p.Type,
		Index: p.Index,
	}
	if p.Next != nil {
		clone.Next = p.Next.Truncate(depth - 1)
	}

	return clone
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

// IsWildcard returns true if the path has unescaped "*" as Name.
func (p *Path) IsWildcard() bool {
	return p.Name == wildcard
}

// String returns a string representation of the Path.
// The result contains the escape characters, if any.
func (p *Path) String() string {
	return p.toString(true)
}

// UnescapedString returns a string representation of the Path
// without the escape characters.
func (p *Path) UnescapedString() string {
	return p.toString(false)
}

func (p *Path) toString(showEscapeChars bool) string {
	path := ""
	if p.Name != nil {
		if p.Name == wildcard {
			path += "*"
		} else if showEscapeChars {
			path += escapeReplacer.Replace(*p.Name)
		} else {
			path += *p.Name
		}
	}
	switch p.Type {
	case PathTypeElement:
	case PathTypeArray:
		if p.Index != nil {
			path += fmt.Sprintf("[%d]", *p.Index)
		} else {
			path += "[]"
		}
	case PathTypeObject:
		path += "."
	}

	if p.Next != nil {
		path += p.Next.toString(showEscapeChars)
	}
	return path
}

// Unescape remove escape characters from a path without parsing it.
func Unescape(path string) string {
	return removeEscapeChars(path)
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
// The wildcard `*` can be used to match all keys of an object. It is only effective
// if it is an entire path segment. For example, the `*` in `field*name` won't be
// considered a wildcard and only the literal field will match.
//
// Special characters defined in the `Escape` map (by default `*`, `[`, `]`, `.` and `\`)
// can be escaped using a backslack `\`.
//
// Example paths:
//
//	name
//	object.field
//	object.subobject.field
//	object.*
//	object.\*
//	object.array[]
//	object.arrayOfObjects[].field
//	[]
//	[].field
//	field*name
//	object.field\[]
//	object.field\[text\]
//	abc\[\]def
//	path\\to\\element
func Parse(p string) (*Path, error) {
	rootPath := &Path{}
	path := rootPath

	if p == "" {
		rootPath.Name = &p
		return rootPath, nil
	}
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
			if t == "*" {
				path.Name = wildcard
			} else {
				t = removeEscapeChars(t)
				path.Name = &t
			}
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

// MustParse is the same as `Parse` but panics if there is an error.
func MustParse(p string) *Path {
	path, err := Parse(p)
	if err != nil {
		panic(err)
	}
	return path
}

// Depth calculate the path's depth without parsing it.
func Depth(p string) uint {
	return uint(strings.Count(p, ".")-strings.Count(p, `\.`)+strings.Count(p, "[]")-strings.Count(p, `\[]`)) + 1
}

func createPathScanner(path string) *bufio.Scanner {
	scanner := bufio.NewScanner(strings.NewReader(path))
	split := func(data []byte, atEOF bool) (int, []byte, error) {
		if len(path) == 0 {
			return len(data), data[:], errors.Errorf("illegal syntax: \"%s\" (path is empty)", path)
		}
		if path[0] == '.' {
			return len(data), data[:], errors.Errorf("illegal syntax: \"%s\" (path cannot start with a dot)", path)
		}
		for width, i := 0, 0; i < len(data); i += width {
			var r rune
			r, width = utf8.DecodeRune(data[i:])

			if i+width < len(data) {
				next, nextWidth := utf8.DecodeRune(data[i+width:])
				if syntaxErr := checkSyntax(r, next); syntaxErr != nil {
					return len(data), data[:], errors.Errorf("illegal syntax: \"%s\" (%w)", path, syntaxErr)
				}

				if _, escapeNext := EscapeChars[next]; r == '\\' && escapeNext {
					// Skip the next character
					width += nextWidth
					r = next
					next, _ = utf8.DecodeRune(data[i+width:])
				}

				if r == '.' && i == 0 {
					return i + width, data[:i+width], nil
				} else if next == '.' || next == '[' {
					return i + width, data[:i+width], nil
				}
			} else if r == '.' || r == '[' || r == '\\' {
				return len(data), data[:], errors.Errorf("illegal syntax: \"%s\" (path cannot end with a dot, an open bracket or a backslash)", path)
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

func checkSyntax(r rune, next rune) error {
	if r == '.' && next == '.' {
		return fmt.Errorf("a dot cannot be followed by another dot")
	}
	if r == '[' && next != ']' {
		return fmt.Errorf("an open bracket must be followed by a closed bracket")
	}
	if r == '.' && (next == ']' || next == '[') {
		return fmt.Errorf("a dot cannot be followed by brackets")
	}
	if r != '.' && r != '[' && r != '\\' && next == ']' {
		return fmt.Errorf("a closed bracket must be preceded by an open bracket")
	}
	if r == ']' && next != '[' && next != '.' {
		return fmt.Errorf("a closed bracket can only be followed by an open bracket or a dot")
	}
	_, escapeNext := EscapeChars[next]
	if r == '\\' && !escapeNext {
		return fmt.Errorf("cannot escape character \"%s\"", string(next))
	}
	return nil
}

func removeEscapeChars(t string) string {
	return escapeRemover.Replace(t)
}
