package validation

import (
	"strings"

	"slices"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/walk"
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

	// init unexported method to force compositing with `BaseValidator`.
	init(opts *Options)

	// Init the validator with the resources required by the `Composable` interface.
	Init(opts *Options)

	// Validate checks the field under validation satisfies this validator's criteria.
	// If necessary, replaces the `Context.Value` with a converted value (see `IsType()`).
	Validate(ctx *Context) bool

	// Name returns the string name of the validator.
	// This is used to generate the language entry for the
	// validation error message.
	Name() string

	// IsTypeDependent returns true if the validator is type-dependent.
	// Type-dependent validators can be used with different field types
	// (numeric, string, arrays, objects and files) and have a different validation messages
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

	overrideMessage(langEntry string)
	getMessageOverride() string
}

// BaseValidator composable structure that implements the basic functions required to
// satisfy the `Validator` interface.
type BaseValidator struct {
	component
	messageOverride string
}

func (v *BaseValidator) init(options *Options) {
	v.component = component{
		db:     options.DB,
		config: options.Config,
		lang:   options.Language,
		logger: options.Logger,
	}
}

// Init the validator with the resources required by the `Composable` interface.
func (v *BaseValidator) Init(options *Options) {
	v.init(options)
}

// IsTypeDependent returns false.
func (v *BaseValidator) IsTypeDependent() bool { return false }

// IsType returns false.
func (v *BaseValidator) IsType() bool { return false }

// MessagePlaceholders returns an empty slice (no placeholders)
func (v *BaseValidator) MessagePlaceholders(_ *Context) []string { return []string{} }

func (v *BaseValidator) overrideMessage(langEntry string) {
	v.messageOverride = langEntry
}

func (v *BaseValidator) getMessageOverride() string {
	return v.messageOverride
}

// WithMessage set a custom language entry for the error message of a Validator.
// Original placeholders returned by the validator are still used to render the message.
// Type-dependent and "element" suffixes are not added when the message is overridden.
func WithMessage[V Validator](v V, langEntry string) V {
	v.overrideMessage(langEntry)
	return v
}

// FieldRulesConverter types implementing this interface define their behavior
// when converting a `FieldRules` to `Rules`. This enables rule sets composition.
type FieldRulesConverter interface {
	convert(path string, field *FieldRules, prefixDepth uint) Rules
}

// List of validators which will be applied on the field. The validators are executed in the
// order of the slice.
type List []Validator

func (l List) convert(path string, field *FieldRules, prefixDepth uint) Rules {
	f := newField(path, field.Rules.(List), prefixDepth)
	return Rules{f}
}

// FieldRules structure associating a path (see `walk.Path`) identifying a field
// with a `FieldRulesApplier` (a `List` of rules or another `RuleSet` via composition).
type FieldRules struct {
	Rules FieldRulesConverter
	Path  string
}

// RuleSet definition of the validation rules applied on each field in the request.
// RuleSets are not meant to be re-used across multiple requests nor used concurrently.
type RuleSet []*FieldRules

func (r RuleSet) convert(path string, _ *FieldRules, _ uint) Rules {
	return r.asRulesWithPrefix(path)
}

// AsRules converts this RuleSet to a Rules structure.
func (r RuleSet) AsRules() Rules {
	return r.asRulesWithPrefix("")
}

func (r RuleSet) asRulesWithPrefix(prefix string) Rules {
	pDepth := uint(0)
	if prefix != "" {
		pDepth = walk.Depth(prefix)
	}

	r = r.injectArrayParents()

	rules := make(Rules, 0, len(r))
	// Keep a map for array fields to easily assign their element field later
	arrays := make(map[string]*Field, len(r))
	for _, field := range r {
		path := prefix
		if field.Path != CurrentElement {
			if strings.HasPrefix(field.Path, "[]") || path == "" {
				path += field.Path
			} else {
				path += "." + field.Path
			}
		}

		fields := field.Rules.convert(path, field, pDepth)

		rules = append(rules, fields...)
		for _, f := range fields {
			if f.isArray {
				arrays[f.Path.String()] = f
			}
		}
	}

	rules.checkDuplicates()

	for {
		arrayElement, index, ok := lo.FindIndexOf(rules, func(f *Field) bool {
			p := f.Path
			for range int(pDepth) - 1 {
				p = lo.Ternary(p.Next == nil, p, p.Next)
			}
			relativePath := p.String()
			return strings.HasSuffix(relativePath, "[]")
		})
		if !ok {
			break
		}

		parentArrayPath := arrayElement.Path.Clone()
		lastParent := parentArrayPath.LastParent()
		lastParent.Type = walk.PathTypeElement
		lastParent.Next = nil

		parentArrayPathStr := parentArrayPath.String()
		parentArrayElement, parentFound := arrays[parentArrayPathStr]

		rules = slices.Delete(rules, index, index+1)
		if parentFound { // Should never be false because we injected array parents and there are no duplicates.
			arrayElement.Path = &walk.Path{Type: walk.PathTypeArray, Next: &walk.Path{}}
			parentArrayElement.Elements = arrayElement
		}
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
		// len(r) MUST be re-evaluated each loop, using "range r" would break it
		// because the length is only evaluated once at the start of the loop.
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

func (r Rules) checkDuplicates() {
	paths := make(map[string]struct{}, len(r))
	for _, f := range r {
		path := f.Path.String()
		includeElementsKeys(paths, path, f.Elements)
		if _, exists := paths[path]; exists {
			panic(errors.Errorf("validation.RuleSet: duplicate path %q in rule set", path))
		}
		paths[path] = struct{}{}
	}
}

func includeElementsKeys(paths map[string]struct{}, path string, elementField *Field) {
	if elementField == nil {
		return
	}
	elementPath := path + "[]"
	paths[elementPath] = struct{}{}
	includeElementsKeys(paths, elementPath, elementField.Elements)
}

// Rules is the result of the transformation of RuleSet using `AsRules()`.
// It is a format that is more easily machine-readable than RuleSet.
type Rules []*Field

// AsRules returns itself.
func (r Rules) AsRules() Rules {
	return r
}
