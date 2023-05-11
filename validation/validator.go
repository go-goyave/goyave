package validation

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"goyave.dev/goyave/v4/lang"
	"goyave.dev/goyave/v4/util/sliceutil"
	"goyave.dev/goyave/v4/util/walk"
)

const (
	// CurrentElement special key for field name in composite rule sets.
	// Use it if you want to apply rules to the current object element.
	// You cannot apply rules on the root element, these rules will only
	// apply if the rule set is used with composition.
	CurrentElement = ""
)

// Ruler adapter interface for method dispatching between RuleSet and Rules
// at route registration time. Allows to input both of these types as parameters
// of the Route.Validate method.
type Ruler interface {
	AsRules() *Rules
}

// Context validation context for RuleFunc.
// Contains all the information needed for validation rules.
type Context struct {
	Data   map[string]interface{}
	Extra  map[string]interface{}
	Value  interface{}
	Parent interface{}
	Field  *Field
	Rule   *Rule
	Now    time.Time

	// The name of the field under validation
	Name  string
	valid bool // Set to false if there was at least one validation error on the field
}

// NewContext returns a new valid context. Used for testing.
func NewContext() *Context {
	return &Context{
		valid: true,
	}
}

// Valid returns false if at least one validator prior to the current one didn't pass
// on the field under validation.
func (c *Context) Valid() bool {
	return c.valid
}

// RuleFunc function defining a validation rule.
// Passing rules should return true, false otherwise.
//
// Rules can modifiy the validated value if needed.
// For example, the "numeric" rule converts the data to float64 if it's a string.
type RuleFunc func(*Context) bool

// RuleDefinition is the definition of a rule, containing the information
// related to the behavior executed on validation-time.
type RuleDefinition struct {

	// The Function field is the function that will be executed
	Function RuleFunc

	// The minimum amount of parameters
	RequiredParameters int

	// A type rule is a rule that checks if a field has a certain type
	// and can convert the raw value to a value fitting. For example, the UUID
	// rule is a type rule because it takes a string as input, checks if it's a
	// valid UUID and converts it to a "uuid.UUID".
	// The "array" rule is an exception. It does convert the value to a new slice of
	// the correct type if provided, but is not considered a type rule to avoid being
	// able to be used as parameter for itself ("array:array").
	IsType bool

	// Type-dependent rules are rules that can be used with different field types
	// (numeric, string, arrays and files) and have a different validation messages
	// depending on the type.
	// The language entry used will be "validation.rules.rulename.type"
	IsTypeDependent bool

	// ComparesFields is true when the rule compares the value of the field under
	// validation with another field. A field containing at least one rule with
	// ComparesFields = true will be executed later in the validation process to
	// ensure conversions are properly executed prior.
	ComparesFields bool
}

// RuleSetApplier types implementing this interface define their behavior
// when they're applied to a RuleSet. This enables rule sets composition.
type RuleSetApplier interface {
	apply(set RuleSet, name string)
}

// FieldMapApplier types implementing this interface define their behavior
// when they're applied to a FieldMap. This enables verbose syntax
// rule sets composition.
type FieldMapApplier interface {
	apply(fieldMap FieldMap, name string)
}

// List of rules string representation.
// e.g.: `validation.List{"required", "min:10"}`
type List []string

func (l List) apply(set RuleSet, name string) {
	set[name] = l
}

// StructList of rules struct representation.
// e.g.:
//
//	validation.StructList{
//		{Name: "required"},
//		{Name: "min", Params: []string{"3"}},
//	}
type StructList []*Rule

func (l StructList) apply(set RuleSet, name string) {
	set[name] = l
}

// RuleSet is a request rules definition. Each entry is a field in the request.
type RuleSet map[string]RuleSetApplier

var _ Ruler = (RuleSet)(nil) // implements Ruler

// AsRules parses and checks this RuleSet and returns it as Rules.
func (r RuleSet) AsRules() *Rules {
	return r.parse()
}

// Parse converts the more convenient RuleSet validation rules syntax to
// a Rules map.
func (r RuleSet) parse() *Rules {
	r.processComposition()
	rules := &Rules{
		Fields: make(FieldMap, len(r)),
	}
	for k, fieldRules := range r {
		var parsedRules []*Rule
		switch fr := fieldRules.(type) {
		case List:
			parsedRules = make([]*Rule, 0, len(fr))
			for _, v := range fr {
				parsedRules = append(parsedRules, parseRule(v))
			}
		case StructList:
			parsedRules = make([]*Rule, 0, len(fr))
			for _, v := range fr {
				cpy := &Rule{Name: v.Name, Params: []string{}}
				cpy.Params = append(cpy.Params, v.Params...)
				parsedRules = append(parsedRules, cpy)
			}
		}
		rules.Fields[k] = &Field{Rules: parsedRules}
	}
	rules.Check()
	return rules
}

func (r RuleSet) processComposition() {
	for name, field := range r {
		field.apply(r, name)
	}
	delete(r, CurrentElement)
}

func (r RuleSet) apply(set RuleSet, name string) {
	for k, rules := range r {
		if k != CurrentElement {
			rules.apply(set, name+"."+k)
		}
	}
	rules, ok := r[CurrentElement]
	if ok {
		rules.apply(set, name)
	}
}

// Rule is a component of rule sets for route validation. Each validated fields
// has one or multiple validation rules. The goal of this struct is to
// gather information about how to use a rule definition for this field.
// This inludes the rule name (referring to a RuleDefinition), the parameters
// and the array dimension for array validation.
type Rule struct {
	Name   string
	Params []string
}

// IsType returns true if the rule definition is a type rule.
// See RuleDefinition.IsType
func (r *Rule) IsType() bool {
	if r.Name == "nullable" {
		return false
	}
	def, exists := validationRules[r.Name]
	if !exists {
		panic(fmt.Sprintf("Rule \"%s\" doesn't exist", r.Name))
	}
	return def.IsType
}

// IsTypeDependent returns true if the rule definition is a type-dependent rule.
// See RuleDefinition.IsTypeDependent
func (r *Rule) IsTypeDependent() bool {
	if r.Name == "nullable" {
		return false
	}
	def, exists := validationRules[r.Name]
	if !exists {
		panic(fmt.Sprintf("Rule \"%s\" doesn't exist", r.Name))
	}
	return def.IsTypeDependent
}

// PostValidationHook executed after the whole validation process is over, no matter
// if the validation passes or not. The errors parameter is never nil (but can be empty).
// A post validation hook can process additional checks on the data and alter the resulting "validation.Errors"
// as needed.
// These hooks always return a "validation.Errors", typically the same as the one they received as
// a paramater, but it is possible to return an entirely different instance of "validation.Errors".
type PostValidationHook func(data map[string]interface{}, errors Errors, now time.Time) Errors

// FieldMap is an alias to shorten verbose validation rules declaration.
// Maps a field name (key) with a Field struct (value).
type FieldMap map[string]FieldMapApplier

// Rules is a component of route validation and maps a
// field name (key) with a Field struct (value).
type Rules struct {
	Fields              FieldMap
	PostValidationHooks []PostValidationHook
	sortedKeys          []string
	checked             bool
}

var _ Ruler = (*Rules)(nil) // implements Ruler

// AsRules performs the checking and returns the same Rules instance.
func (r *Rules) AsRules() *Rules {
	r.Check()
	return r
}

// Check all rules in this set. This function will panic if
// any of the rules doesn't refer to an existing RuleDefinition, doesn't
// meet the parameters requirement, or if the rule cannot be used in array validation
// while ArrayDimension is not equal to 0.
// Also processes composition. After calling this function, you can safely assume all
// `Fields` elements are of type `*Field`.
func (r *Rules) Check() {
	if !r.checked {
		r.processComposition()
		r.sortKeys()
		for _, path := range r.sortedKeys {
			field := r.Fields[path].(*Field)
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
					parent.(*Field).Elements = field
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

func (r *Rules) processComposition() {
	for name, field := range r.Fields {
		field.apply(r.Fields, name)
	}
	delete(r.Fields, CurrentElement)
}

func (r *Rules) apply(fieldMap FieldMap, name string) {
	for k, f := range r.Fields {
		if k != CurrentElement {
			f.apply(fieldMap, name+"."+k)
		}
	}
	fields, ok := r.Fields[CurrentElement]
	if ok {
		fields.apply(fieldMap, name)
	}
}

func (r *Rules) sortKeys() {
	r.sortedKeys = make([]string, 0, len(r.Fields))

	for k := range r.Fields {
		if k != CurrentElement {
			r.sortedKeys = append(r.sortedKeys, k)
		}
	}

	sort.SliceStable(r.sortedKeys, func(i, j int) bool {
		fieldName1 := r.sortedKeys[i]
		field2 := r.Fields[r.sortedKeys[j]].(*Field)
		for _, r := range field2.Rules {
			def, ok := validationRules[r.Name]
			if ok && def.ComparesFields && sliceutil.ContainsStr(r.Params, fieldName1) {
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
}

var validationRules map[string]*RuleDefinition

func init() {
	validationRules = map[string]*RuleDefinition{
		"required":           {validateRequired, 0, false, false, false},
		"numeric":            {validateNumeric, 0, true, false, false},
		"integer":            {validateInteger, 0, true, false, false},
		"min":                {validateMin, 1, false, true, false},
		"max":                {validateMax, 1, false, true, false},
		"between":            {validateBetween, 2, false, true, false},
		"greater_than":       {validateGreaterThan, 1, false, true, true},
		"greater_than_equal": {validateGreaterThanEqual, 1, false, true, true},
		"lower_than":         {validateLowerThan, 1, false, true, true},
		"lower_than_equal":   {validateLowerThanEqual, 1, false, true, true},
		"string":             {validateString, 0, true, false, false},
		"array":              {validateArray, 0, false, false, false},
		"distinct":           {validateDistinct, 0, false, false, false},
		"digits":             {validateDigits, 0, false, false, false},
		"regex":              {validateRegex, 1, false, false, false},
		"email":              {validateEmail, 0, false, false, false},
		"size":               {validateSize, 1, false, true, false},
		"alpha":              {validateAlpha, 0, false, false, false},
		"alpha_dash":         {validateAlphaDash, 0, false, false, false},
		"alpha_num":          {validateAlphaNumeric, 0, false, false, false},
		"starts_with":        {validateStartsWith, 1, false, false, false},
		"ends_with":          {validateEndsWith, 1, false, false, false},
		"in":                 {validateIn, 1, false, false, false},
		"not_in":             {validateNotIn, 1, false, false, false},
		"in_array":           {validateInArray, 1, false, false, true},
		"not_in_array":       {validateNotInArray, 1, false, false, true},
		"timezone":           {validateTimezone, 0, true, false, false},
		"ip":                 {validateIP, 0, true, false, false},
		"ipv4":               {validateIPv4, 0, true, false, false},
		"ipv6":               {validateIPv6, 0, true, false, false},
		"json":               {validateJSON, 0, true, false, false},
		"url":                {validateURL, 0, true, false, false},
		"uuid":               {validateUUID, 0, true, false, false},
		"bool":               {validateBool, 0, true, false, false},
		"same":               {validateSame, 1, false, false, true},
		"different":          {validateDifferent, 1, false, false, true},
		"file":               {validateFile, 0, false, false, false},
		"mime":               {validateMIME, 1, false, false, false},
		"image":              {validateImage, 0, false, false, false},
		"extension":          {validateExtension, 1, false, false, false},
		"count":              {validateCount, 1, false, false, false},
		"count_min":          {validateCountMin, 1, false, false, false},
		"count_max":          {validateCountMax, 1, false, false, false},
		"count_between":      {validateCountBetween, 2, false, false, false},
		"date":               {validateDate, 0, true, false, false},
		"before":             {validateBefore, 1, false, false, true},
		"before_equal":       {validateBeforeEqual, 1, false, false, true},
		"after":              {validateAfter, 1, false, false, true},
		"after_equal":        {validateAfterEqual, 1, false, false, true},
		"date_equals":        {validateDateEquals, 1, false, false, true},
		"date_between":       {validateDateBetween, 2, false, false, true},
		"before_now":         {validateDateBeforeNow, 0, false, false, false},
		"after_now":          {validateDateAfterNow, 0, false, false, false},
		"object":             {validateObject, 0, true, false, false},
	}
}

// AddRule register a validation rule.
// The rule will be usable in request validation by using the
// given rule name.
//
// Type-dependent messages let you define a different message for
// numeric, string, arrays and files.
// The language entry used will be "validation.rules.rulename.type"
func AddRule(name string, rule *RuleDefinition) {
	if _, exists := validationRules[name]; exists {
		panic(fmt.Sprintf("Rule %s already exists", name))
	}
	validationRules[name] = rule
}

// Validate the given data with the given rule set.
// If all validation rules pass, returns nil.
// Third parameter tells the function if the data comes from a JSON request.
// Last parameter sets the language of the validation error messages.
func Validate(data map[string]interface{}, rules Ruler, isJSON bool, language string) Errors {
	return ValidateWithExtra(data, rules, isJSON, language, map[string]interface{}{})
}

// ValidateWithExtra the given data with the given rule set.
// If all validation rules pass, returns nil.
// Third parameter tells the function if the data comes from a JSON request.
// The fourth parameter sets the language of the validation error messages.
// The last parameter is a map of extra information given to validation rules via `validation.Context.Extra`.
// This map is copied for each validation rule. One copy of the extra map is therefore scoped to a single
// validation rule function call and the resulting validation error message / placeholder.
func ValidateWithExtra(data map[string]interface{}, rules Ruler, isJSON bool, language string, extra map[string]interface{}) Errors {
	if data == nil {
		var malformedMessage string
		if isJSON {
			malformedMessage = lang.Get(language, "malformed-json")
		} else {
			malformedMessage = lang.Get(language, "malformed-request")
		}
		return Errors{"[data]": &FieldErrors{Errors: []string{malformedMessage}}}
	}

	errsBag := validate(data, isJSON, rules.AsRules(), language, extra)
	if len(errsBag) == 0 {
		return nil
	}
	return errsBag
}

func validate(data map[string]interface{}, isJSON bool, rules *Rules, language string, extra map[string]interface{}) Errors {
	errors := Errors{}
	now := time.Now()

	for _, fieldName := range rules.sortedKeys {
		field := rules.Fields[fieldName].(*Field)
		validateField(fieldName, field, isJSON, data, data, nil, now, language, extra, errors)
	}
	for _, hook := range rules.PostValidationHooks {
		errors = hook(data, errors, now)
	}
	return errors
}

func validateField(fieldName string, field *Field, isJSON bool, data map[string]interface{}, walkData interface{}, parentPath *walk.Path, now time.Time, language string, extra map[string]interface{}, errors Errors) {
	field.Path.Walk(walkData, func(c walk.Context) {
		parentObject, parentIsObject := c.Parent.(map[string]interface{})
		if c.Found == walk.Found {
			if parentIsObject && !field.IsNullable() && c.Value == nil {
				delete(parentObject, c.Name)
			}

			if shouldConvertSingleValueArray(fieldName, isJSON) {
				c.Value = convertSingleValueArray(field, c.Value) // Convert single value arrays in url-encoded requests
				parentObject[c.Name] = c.Value
			}
		}

		if isAbsent(field, c, data) {
			return
		}

		if field.Elements != nil {
			// This is an array, recursively validate it so it can be converted to correct type
			if _, ok := c.Value.([]interface{}); !ok {
				if newValue, ok := makeGenericSlice(c.Value); ok {
					replaceValue(c.Value, c)
					c.Value = newValue
				}
			}

			path := c.Path
			if parentPath != nil {
				clone := parentPath.Clone()
				tail := clone.Tail()
				tail.Type = walk.PathTypeArray
				tail.Index = &c.Index
				tail.Next = path.Next
				path = clone
			}
			validateField(fieldName+"[]", field.Elements, isJSON, data, c.Value, path, now, language, extra, errors)
		}

		value := c.Value
		valid := true
		for _, rule := range field.Rules {
			if rule.Name == "nullable" {
				if value == nil {
					break
				}
				continue
			}

			ctx := &Context{
				Data:   data,
				Extra:  cloneMap(extra),
				Value:  value,
				Parent: c.Parent,
				Field:  field,
				Rule:   rule,
				Now:    now,
				Name:   c.Name,
				valid:  valid,
			}
			if !validationRules[rule.Name].Function(ctx) {
				valid = false
				path := field.getErrorPath(parentPath, c)
				message := processPlaceholders(fieldName, getMessage(field, rule, reflect.ValueOf(value), language), language, ctx)
				errors.Add(path, message)
				continue
			}

			value = ctx.Value
		}
		// Value may be modified (converting rule), replace it in the parent element
		replaceValue(value, c)
	})
}

func cloneMap(m map[string]interface{}) map[string]interface{} {
	clone := make(map[string]interface{}, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

func isAbsent(field *Field, c walk.Context, data map[string]interface{}) bool {
	if c.Found == walk.ParentNotFound {
		return true
	}
	requiredCtx := &Context{
		Data:   data,
		Value:  c.Value,
		Parent: c.Parent,
		Field:  field,
		Rule:   &Rule{Name: "required"},
		Name:   c.Name,
	}
	return !field.IsRequired() && !validateRequired(requiredCtx)
}

func shouldConvertSingleValueArray(fieldName string, isJSON bool) bool {
	return !isJSON && !strings.Contains(fieldName, ".") && !strings.Contains(fieldName, "[]")
}

func replaceValue(value interface{}, c walk.Context) {
	if c.Found != walk.Found {
		return
	}

	if parentObject, ok := c.Parent.(map[string]interface{}); ok {
		parentObject[c.Name] = value
	} else {
		// Parent is slice
		reflect.ValueOf(c.Parent).Index(c.Index).Set(reflect.ValueOf(value))
	}
}

func makeGenericSlice(original interface{}) ([]interface{}, bool) {
	list := reflect.ValueOf(original)
	if list.Kind() != reflect.Slice {
		return nil, false
	}
	length := list.Len()
	newSlice := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		newSlice = append(newSlice, list.Index(i).Interface())
	}
	return newSlice, true
}

func convertSingleValueArray(field *Field, value interface{}) interface{} {
	rv := reflect.ValueOf(value)
	kind := rv.Kind().String()
	if field.IsArray() && kind != "slice" {
		rt := reflect.TypeOf(value)
		slice := reflect.MakeSlice(reflect.SliceOf(rt), 0, 1)
		slice = reflect.Append(slice, rv)
		return slice.Interface()
	}
	return value
}

func getMessage(field *Field, rule *Rule, value reflect.Value, language string) string {
	langEntry := "validation.rules." + rule.Name
	if rule.IsTypeDependent() {
		expectedType := findTypeRule(field.Rules)
		if expectedType == "unsupported" {
			langEntry += "." + getFieldType(value)
		} else {
			if expectedType == "integer" {
				expectedType = "numeric"
			}
			langEntry += "." + expectedType
		}
	}

	lastParent := field.Path.LastParent()
	if lastParent != nil && lastParent.Type == walk.PathTypeArray {
		langEntry += ".array"
	}

	return lang.Get(language, langEntry)
}

// findTypeRule find the expected type of a field for a given array dimension.
func findTypeRule(rules []*Rule) string {
	for _, rule := range rules {
		if rule.IsType() {
			return rule.Name
		}
	}
	return "unsupported"
}

// GetFieldType returns the non-technical type of the given "value" interface.
// This is used by validation rules to know if the input data is a candidate
// for validation or not and is especially useful for type-dependent rules.
//   - "numeric" if the value is an int, uint or a float
//   - "string" if the value is a string
//   - "array" if the value is a slice
//   - "file" if the value is a slice of "fsutil.File"
//   - "unsupported" otherwise
func GetFieldType(value interface{}) string {
	return getFieldType(reflect.ValueOf(value))
}

func getFieldType(value reflect.Value) string {
	kind := value.Kind().String()
	switch {
	case strings.HasPrefix(kind, "int"), strings.HasPrefix(kind, "uint") && kind != "uintptr", strings.HasPrefix(kind, "float"):
		return "numeric"
	case kind == "string":
		return "string"
	case kind == "slice":
		if value.Type().String() == "[]fsutil.File" {
			return "file"
		}
		return "array"
	default:
		if value.IsValid() {
			if _, ok := value.Interface().(map[string]interface{}); ok {
				return "object"
			}
		}
		return "unsupported"
	}
}

// GetFieldFromName find potentially nested field by it's dot-separated path
// in the given object.
// Returns the name without its prefix, the value, its parent object and a bool indicating if it has been found or not.
func GetFieldFromName(name string, data map[string]interface{}) (string, interface{}, map[string]interface{}, bool) {
	key := name
	i := strings.Index(name, ".")
	if i != -1 {
		key = name[:i]
	}
	val, ok := data[key]
	if !ok {
		return "", nil, nil, false
	}

	if i != -1 {
		if obj, ok := val.(map[string]interface{}); ok {
			return GetFieldFromName(name[len(key)+1:], obj)
		}
	}

	return name, val, data, ok
}

func parseRule(rule string) *Rule {
	indexName := strings.Index(rule, ":")
	params := []string{}
	var ruleName string
	if indexName == -1 {
		if strings.Count(rule, ",") > 0 {
			panic(fmt.Sprintf("Invalid rule: \"%s\"", rule))
		}
		ruleName = rule
	} else {
		ruleName = rule[:indexName]
		params = strings.Split(rule[indexName+1:], ",")
	}

	return &Rule{ruleName, params}
}
