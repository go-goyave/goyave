package validation

import (
	"sort"
	"strings"

	"golang.org/x/exp/slices"
	"goyave.dev/goyave/v4/util/walk"
)

// Ruler adapter interface to make allow both RuleSet and Rules to
// be used when calling `Validate()`.
type Ruler interface {
	AsRules() Rules
}

// Validator is a Component validating a field value.
// A validator should not be re-usable or usable concurrently. They are meant to be
// scoped to a single field validation in a single request.
type Validator interface {
	Composable
	init(*Options)

	// Validate checks the field under validation satisfies this validator's criteria.
	// If necessary, replaces the `Context.Value` with a converted value (see `IsType()`).
	Validate(*Context) bool

	// Name returns the string name of the validator.
	// This is used to generate the language entry for the
	// validation error message.
	Name() string

	// IsTypeDependent returns true if the validator is type-dependent.
	// Type-dependent validators can be used with different field types
	// (numeric, string, arrays and files) and have a different validation messages
	// depending on the type.
	// The language entry used will be "validation.rules.rulename.type"
	IsTypeDependent() bool

	// IsType returns true if the validator if a type validator.
	// A type validator checks if a field has a certain type
	// and can convert the raw value to a value fitting. For example, the UUID
	// validator is a type validator because it takes a string as input, checks if it's a
	// valid UUID and converts it to a `uuid.UUID`.
	IsType() bool

	// MessagePlaceholders returns an associative slice of placeholders and their replacement.
	// This is use to generate the validation error message. An empty slice can be returned.
	// See `lang.Language.Get()` for more details.
	MessagePlaceholders(ctx *Context) []string
}

// BaseValidator composable structure that implements the basic functions required to
// satisfy the `Validator` interface.
type BaseValidator struct {
	component
}

func (v *BaseValidator) init(options *Options) {
	v.component = component{
		db:        options.DB,
		config:    options.Config,
		lang:      options.Language,
		logger:    options.Logger,
		errLogger: options.ErrLogger,
	}
}

// IsTypeDependent returns false.
func (v *BaseValidator) IsTypeDependent() bool { return false }

// IsType returns false.
func (v *BaseValidator) IsType() bool { return false }

// MessagePlaceholders returns an empty slice (no placeholders)
func (v *BaseValidator) MessagePlaceholders(_ *Context) []string { return []string{} }

// FieldRulesApplier types implementing this interface define their behavior
// when applying a `FieldRules` to `Rules`. This enables rule sets composition.
type FieldRulesApplier interface {
	apply(rules Rules, path string, field *FieldRules, pDepth uint, pendingArrays map[string]any) Rules
}

// List of validators which will be applied on the field. The validators are executed in the
// order of the slice.
type List []Validator

func (l List) apply(rules Rules, path string, field *FieldRules, pDepth uint, pendingArrays map[string]any) Rules {
	isArrayElement := strings.HasSuffix(field.Path, "[]")
	if isArrayElement {
		path = "[]"
	}
	f := newField(path, field.Rules.(List), pDepth)
	if !isArrayElement {
		rules = append(rules, f)
	} else {
		pendingArrays[field.Path[:len(field.Path)-2]] = f
	}

	if arrayElements, ok := pendingArrays[field.Path]; ok {
		switch applier := arrayElements.(type) {
		case *Field:
			f.Elements = applier
		case Rules:
			i := 0
			if len(applier) > 0 && applier[1].Path.Type == walk.PathTypeElement && applier[1].Path.Name != nil && *applier[1].Path.Name == CurrentElement {
				i = 1
				f.Elements = applier[1]
			}
			rules = append(rules, applier[i:]...)
		}
		delete(pendingArrays, field.Path)
	}
	return rules
}

// FieldRules structure associating a path (see `walk.Path`) identifying a field
// with a `FieldRulesApplier` (a `List` of rules or another `RuleSet` via composition).
type FieldRules struct {
	// TODO what behavior if there are duplicates? If it ever becomes a problem, can probably merge the Lists. But it's unnecessary for now.
	Rules FieldRulesApplier
	Path  string
}

// RuleSet definition of the validation rules applied on each field in the request.
// RuleSets are not meant to be re-used across multiple requests nor used concurrently.
type RuleSet []*FieldRules

func (r RuleSet) apply(rules Rules, path string, field *FieldRules, _ uint, pendingArrays map[string]any) Rules {
	if strings.HasSuffix(field.Path, "[]") {
		pendingArrays[field.Path[:len(field.Path)-2]] = field.Rules.(RuleSet).asRulesWithPrefix(path)
		return rules
	}
	return append(rules, field.Rules.(RuleSet).asRulesWithPrefix(path)...)
}

// AsRules converts this RuleSet to a Rules structure.
func (r RuleSet) AsRules() Rules {
	return r.asRulesWithPrefix("")
}

func (r RuleSet) asRulesWithPrefix(prefix string) Rules {
	pDepth := uint(0)
	if prefix != "" {
		prefixPath, err := walk.Parse(prefix)
		if err != nil {
			panic(err)
		}
		pDepth = prefixPath.Depth()
	}

	pendingArrays := map[string]any{}
	sortedRuleSet := slices.Clone(r)
	sortedRuleSet = sortedRuleSet.injectArrayParents()
	sortedRuleSet.sort()

	rules := make(Rules, 0, len(r))
	for _, field := range sortedRuleSet {
		p := prefix
		if field.Path != CurrentElement && !strings.HasPrefix("[]", field.Path) { // TODO test the use of composition on CurrentElement and arrays
			if p != "" {
				p += "." + field.Path
			} else {
				p = field.Path
			}
		}

		rules = field.Rules.apply(rules, p, field, pDepth, pendingArrays)
	}
	return rules
}

// injectArrayParents makes sure all array elements in the RuleSet have a parent field.
func (r RuleSet) injectArrayParents() RuleSet {
	keys := make(map[string]struct{}, len(r))
	for _, f := range r {
		keys[f.Path] = struct{}{}
	}
	for i := 0; i < len(r); i++ {
		f := r[i]
		if strings.HasSuffix(f.Path, "[]") {
			parentPath := f.Path[:len(f.Path)-2]
			if _, ok := keys[parentPath]; !ok {
				// No parent array found, inject it
				parent := &FieldRules{
					Path:  parentPath,
					Rules: List{Array()},
				}
				r = append(r[:i+1], append(RuleSet{parent}, r[i+1:]...)...)
			}
		}
	}

	return r
}

func (r RuleSet) sort() {
	// Make sure the array elements are before their parent in the list
	sort.SliceStable(r, func(i, j int) bool {
		field1 := r[i]
		field2 := r[j]
		if strings.HasSuffix(field1.Path, "[]") && !strings.HasSuffix(field2.Path, "[]") {
			return true
		}
		if !strings.HasSuffix(field1.Path, "[]") && !strings.HasSuffix(field2.Path, "[]") {
			return false
		}
		count1 := strings.Count(field1.Path, "[]")
		count2 := strings.Count(field2.Path, "[]")
		return count1 > count2
	})
	sort.SliceStable(r, func(i, j int) bool {
		// CurrentElement must always be first
		return r[i].Path == CurrentElement
	})
}

// Rules is the result of the transformation of RuleSet using `AsRules()`.
// It is a format that is more easily machine-readable than RuleSet.
type Rules []*Field

// AsRules returns itself.
func (r Rules) AsRules() Rules {
	return r
}
