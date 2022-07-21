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
	ComparesWith() string
}

type BaseValidator struct{}

func (v *BaseValidator) IsTypeDependent() bool { return false }
func (v *BaseValidator) IsType() bool          { return false }
func (v *BaseValidator) ComparesFields() bool  { return false }

type RuleSetV5 map[string][]Validator

func (r RuleSetV5) AsRules() *RulesV5 { // TODO composition
	rules := &RulesV5{
		Fields: make(map[string]*FieldV5, len(r)),
	}

	for k, v := range r {
		rules.Fields[k] = &FieldV5{Rules: v}
	}

	rules.Check()
	return rules
}

// Rules is a component of route validation and maps a
// field name (key) with a Field struct (value).
type RulesV5 struct {
	Fields map[string]*FieldV5
	// PostValidationHooks []PostValidationHook
	sortedKeys []string
	checked    bool
}

func (r *RulesV5) Check() {
	if !r.checked {
		r.sortKeys()
		for _, path := range r.sortedKeys {
			field := r.Fields[path]
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
		if k != CurrentElement {
			r.sortedKeys = append(r.sortedKeys, k)
		}
	}

	sort.SliceStable(r.sortedKeys, func(i, j int) bool {
		fieldName1 := r.sortedKeys[i]
		field2 := r.Fields[r.sortedKeys[j]]
		for _, r := range field2.Rules {
			c, ok := r.(ComparatorValidator)
			if ok && c.ComparesWith() == fieldName1 {
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
