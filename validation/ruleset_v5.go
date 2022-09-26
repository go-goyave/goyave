package validation

import (
	"sort"
	"strings"

	"goyave.dev/goyave/v4/util/walk"
)

// Ruler adapter interface for method dispatching between RuleSet and Rules
// at route registration time. Allows to input both of these types as parameters
// of the Route.Validate method.
type RulerV5 interface {
	AsRules() *RulesV5
}

type Validator interface { // TODO rename to Rule?
	Validate(*ContextV5) bool
	Name() string
	IsTypeDependent() bool
	IsType() bool
}

type ComparatorValidator interface {
	ComparesWith() string // Returns a path to the compared element
}

type BaseValidator struct{}

func (v *BaseValidator) IsTypeDependent() bool { return false }
func (v *BaseValidator) IsType() bool          { return false }
func (v *BaseValidator) ComparesFields() bool  { return false }

// Applier interface allowing RuleSet composition.
// Types implementing this interface define how they should
// be applied to the `Fields` map of the `Rules` struct when being converted
// using `AsRules()`.
type Applier interface {
	apply(fields map[string]*FieldV5, path string, prefixDepth uint)
}

type ListV5 []Validator

type RuleSetV5 map[string]Applier

func (r RuleSetV5) apply(fields map[string]*FieldV5, path string, prefixDepth uint) {
	prefix, err := walk.Parse(path)
	if err != nil {
		panic(err)
	}
	pDepth := prefix.Depth()

	for k, rules := range r {
		p := path
		if k != CurrentElement && !strings.HasPrefix("[]", k) { // TODO test the use of composition on CurrentElement and arrays
			p += "." + k
		}
		rules.apply(fields, p, pDepth)
	}
}

func (r ListV5) apply(fields map[string]*FieldV5, path string, prefixDepth uint) {
	fields[path] = &FieldV5{Rules: r, prefixDepth: prefixDepth}
}

// TODO test composition with special cases (arrays, recursive composition, nested composition, etc)

func (r RuleSetV5) AsRules() *RulesV5 {
	rules := &RulesV5{
		Fields: make(map[string]*FieldV5, len(r)),
	}

	for k, v := range r {
		v.apply(rules.Fields, k, 0)
	}

	rules.Check()
	return rules
}

// Rules is the result of the transformation of RuleSet using `AsRules()`.
// It is a format that is more easily machine-readable than RuleSet.
// Before use, parses field paths and creates a sorted map keys slice
// to ensure validation order.
type RulesV5 struct {
	Fields     map[string]*FieldV5
	sortedKeys []string
	checked    bool
}

func (r *RulesV5) Check() {
	if !r.checked {
		for path, field := range r.Fields {
			p, err := walk.Parse(path)
			if err != nil {
				panic(err)
			}
			field.Path = p
			field.Check()
			if strings.HasSuffix(path, "[]") {
				// This field is an element of an array, find it and assign it to f.Elements
				parent, ok := r.Fields[path[:len(path)-2]]
				if ok {
					parent.Elements = field
					field.Path = &walk.Path{
						Type: walk.PathTypeArray,
						Next: &walk.Path{
							Type: walk.PathTypeElement,
						},
					}
					delete(r.Fields, path)
				}
			}
		}
		r.sortKeys()
		r.checked = true
	}
}

func (r *RulesV5) sortKeys() {
	r.sortedKeys = make([]string, 0, len(r.Fields))

	for k := range r.Fields {
		r.sortedKeys = append(r.sortedKeys, k)
	}

	sort.SliceStable(r.sortedKeys, func(i, j int) bool {
		fieldName1 := r.sortedKeys[i]
		field2 := r.Fields[r.sortedKeys[j]]
		for _, r := range field2.Rules {
			c, ok := r.(ComparatorValidator)
			if ok && strings.HasPrefix(c.ComparesWith(), fieldName1) {
				return true
			}
		}
		return false
	})
	sort.SliceStable(r.sortedKeys, func(i, j int) bool {
		count1 := strings.Count(r.sortedKeys[i], "[]")
		count2 := strings.Count(r.sortedKeys[j], "[]")
		if count1 == count2 {
			return false
		}
		return count1 > count2
	})
	sort.SliceStable(r.sortedKeys, func(i, j int) bool {
		// CurrentElement must always be first
		return r.sortedKeys[i] == CurrentElement
	})
}

func (r *RulesV5) AsRules() *RulesV5 {
	r.Check()
	return r
}
