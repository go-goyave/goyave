package validation

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"goyave.dev/goyave/v3/helper"
	"goyave.dev/goyave/v3/lang"
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
	Value  interface{}
	Parent interface{}
	Field  *Field
	Rule   *Rule
	Name   string
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

// RuleSet is a request rules definition. Each entry is a field in the request.
// TODO map of some interface to be able to use composition
type RuleSet map[string][]string

var _ Ruler = (RuleSet)(nil) // implements Ruler

// AsRules parses and checks this RuleSet and returns it as Rules.
func (r RuleSet) AsRules() *Rules {
	return r.parse()
}

// Parse converts the more convenient RuleSet validation rules syntax to
// a Rules map.
func (r RuleSet) parse() *Rules {
	rules := &Rules{
		Fields: make(FieldMap, len(r)),
	}
	for k, r := range r {
		field := &Field{
			Rules: make([]*Rule, 0, len(r)),
		}
		for _, v := range r {
			field.Rules = append(field.Rules, parseRule(v))
		}
		rules.Fields[k] = field
	}
	rules.Check()
	return rules
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

// Field is a component of route validation. A Field is a value in
// a Rules map, the key being the name of the field.
type Field struct {
	Path     *PathItem
	Elements *Field // If the field is an array, the field representing its elements, or nil
	// Maybe use the same concept for objects too?
	Rules      []*Rule
	isArray    bool
	isRequired bool
	isNullable bool
}

// IsRequired check if a field has the "required" rule
func (f *Field) IsRequired() bool {
	return f.isRequired
}

// IsNullable check if a field has the "nullable" rule
func (f *Field) IsNullable() bool {
	return f.isNullable
}

// IsArray check if a field has the "array" rule
func (f *Field) IsArray() bool {
	return f.isArray
}

// Check if rules meet the minimum parameters requirement and update
// the isRequired, isNullable and isArray fields.
func (f *Field) Check() {
	for _, rule := range f.Rules {
		switch rule.Name {
		case "confirmed", "file", "mime", "image", "extension", "count",
			"count_min", "count_max", "count_between":
			if f.Path.HasArray() {
				panic(fmt.Sprintf("Cannot use rule \"%s\" in array validation", rule.Name))
			}
		case "required":
			f.isRequired = true
		case "nullable":
			f.isNullable = true
			continue
		case "array":
			f.isArray = true
		}

		def, exists := validationRules[rule.Name]
		if !exists {
			panic(fmt.Sprintf("Rule \"%s\" doesn't exist", rule.Name))
		}
		if len(rule.Params) < def.RequiredParameters {
			panic(fmt.Sprintf("Rule \"%s\" requires %d parameter(s)", rule.Name, def.RequiredParameters))
		}
	}
}

// FieldMap is an alias to shorten verbose validation rules declaration.
// Maps a field name (key) with a Field struct (value).
type FieldMap map[string]*Field

// Rules is a component of route validation and maps a
// field name (key) with a Field struct (value).
type Rules struct {
	Fields     FieldMap
	sortedKeys []string
	checked    bool
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
func (r *Rules) Check() {
	if !r.checked {
		r.sortKeys()
		for _, path := range r.sortedKeys {
			field := r.Fields[path]
			p, err := ComputePath(path)
			if err != nil {
				// TODO test this
				panic(err)
			}
			field.Path = p
			field.Check()
			if strings.HasSuffix(path, "[]") { // This field is an element of an array, find it and assign it to f.Elements
				parent, ok := r.Fields[path[:len(path)-2]]
				if ok {
					parent.Elements = field
					field.Path = &PathItem{
						Type: PathTypeArray,
						Next: &PathItem{
							Type: PathTypeElement,
						},
					}
					delete(r.Fields, path)
				}
			}
		}
		r.checked = true
		r.sortKeys()
	}
}

func (r *Rules) sortKeys() {
	r.sortedKeys = make([]string, 0, len(r.Fields))

	for k := range r.Fields {
		r.sortedKeys = append(r.sortedKeys, k)
	}

	sort.SliceStable(r.sortedKeys, func(i, j int) bool {
		fieldName1 := r.sortedKeys[i]
		field2 := r.Fields[r.sortedKeys[j]]
		for _, r := range field2.Rules {
			def, ok := validationRules[r.Name]
			if ok && def.ComparesFields && helper.ContainsStr(r.Params, fieldName1) {
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

// Errors is a map of validation errors with the field name as a key.
// TODO Errors should be map[string]interface{} so it's easier to read errors related to object fields
type Errors map[string][]string

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
		"confirmed":          {validateConfirmed, 0, false, false, false},
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
// If all validation rules pass, returns an empty "validation.Errors".
// Third parameter tells the function if the data comes from a JSON request.
// Last parameter sets the language of the validation error messages.
func Validate(data map[string]interface{}, rules Ruler, isJSON bool, language string) Errors {
	if data == nil {
		var malformedMessage string
		if isJSON {
			malformedMessage = lang.Get(language, "malformed-json")
		} else {
			malformedMessage = lang.Get(language, "malformed-request")
		}
		return map[string][]string{"error": {malformedMessage}}
	}

	return validate(data, isJSON, rules.AsRules(), language)
}

func validate(data map[string]interface{}, isJSON bool, rules *Rules, language string) Errors {
	errors := Errors{}

	for _, fieldName := range rules.sortedKeys {
		field := rules.Fields[fieldName]
		validateField(fieldName, field, isJSON, data, data, language, errors)
	}
	return errors
}

func validateField(fieldName string, field *Field, isJSON bool, data map[string]interface{}, walkData interface{}, language string, errors Errors) {
	field.Path.Walk(walkData, func(c WalkContext) {
		parentObject, parentIsObject := c.Parent.(map[string]interface{})
		if parentIsObject {
			if !field.IsNullable() && c.Value == nil {
				delete(parentObject, fieldName)
			}
		}

		if !isJSON && !strings.Contains(fieldName, ".") && !strings.Contains(fieldName, "[]") {
			c.Value = convertSingleValueArray(field, c.Value, parentObject) // Convert single value arrays in url-encoded requests
			parentObject[c.Name] = c.Value
		}

		requiredCtx := &Context{
			Data:   data,
			Value:  c.Value,
			Parent: c.Parent,
			Field:  field,
			Rule:   &Rule{Name: "required"},
			Name:   c.Name,
		}
		if !field.IsRequired() && !validateRequired(requiredCtx) {
			return
		}

		if field.Elements != nil {
			// This is an array, recursively validate it so it can be converted to correct type
			if _, ok := c.Value.([]interface{}); !ok {
				c.Value = makeGenericSlice(c.Value)
				replaceValue(c.Value, c)
			}

			validateField(fieldName+"[]", field.Elements, isJSON, data, c.Value, language, errors)
		}

		value := c.Value
		for _, rule := range field.Rules {
			if rule.Name == "nullable" {
				if value == nil {
					break
				}
				continue
			}

			ctx := &Context{
				Data:   data,
				Value:  value,
				Parent: c.Parent,
				Field:  field,
				Rule:   rule,
				Name:   c.Name,
			}
			if !validationRules[rule.Name].Function(ctx) {
				// TODO test possible duplicate error messages
				errors[fieldName] = append(
					errors[fieldName],
					processPlaceholders(fieldName, rule.Name, rule.Params, getMessage(field, rule, reflect.ValueOf(c.Value), language), language),
				)
				continue
			}

			value = ctx.Value
		}
		// Value may be modified (converting rule), replace it in the parent element
		replaceValue(value, c)
	})
}

func replaceValue(value interface{}, c WalkContext) {
	if parentObject, ok := c.Parent.(map[string]interface{}); ok {
		parentObject[c.Name] = value
	} else {
		// Parent is slice
		reflect.ValueOf(c.Parent).Index(c.Index).Set(reflect.ValueOf(value))
	}
}

func makeGenericSlice(original interface{}) []interface{} {
	list := reflect.ValueOf(original)
	length := list.Len()
	newSlice := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		newSlice = append(newSlice, list.Index(i).Interface())
	}
	return newSlice
}

func convertSingleValueArray(field *Field, value interface{}, data map[string]interface{}) interface{} {
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
	if validationRules[rule.Name].IsTypeDependent {
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
	if lastParent != nil && lastParent.Type == PathTypeArray {
		langEntry += ".array"
	}

	return lang.Get(language, langEntry)
}

// findTypeRule find the expected type of a field for a given array dimension.
func findTypeRule(rules []*Rule) string {
	for _, rule := range rules {
		if validationRules[rule.Name].IsType {
			return rule.Name
		}
	}
	return "unsupported"
}

// GetFieldType returns the non-technical type of the given "value" interface.
// This is used by validation rules to know if the input data is a candidate
// for validation or not and is especially useful for type-dependent rules.
//  - "numeric" if the value is an int, uint or a float
//  - "string" if the value is a string
//  - "array" if the value is a slice
//  - "file" if the value is a slice of "filesystem.File"
//  - "unsupported" otherwise
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
		if value.Type().String() == "[]filesystem.File" {
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
