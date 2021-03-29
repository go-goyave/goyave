package validation

import (
	"fmt"
	"reflect"
	"strings"

	"goyave.dev/goyave/v3/lang"
)

// Ruler adapter interface for method dispatching between RuleSet and Rules
// at route registration time. Allows to input both of these types as parameters
// of the Route.Validate method.
type Ruler interface {
	AsRules() *Rules
}

// RuleFunc function defining a validation rule.
// Passing rules should return true, false otherwise.
//
// Rules can modifiy the validated value if needed.
// For example, the "numeric" rule converts the data to float64 if it's a string.
type RuleFunc func(string, interface{}, []string, map[string]interface{}) bool

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
}

// RuleSet is a request rules definition. Each entry is a field in the request.
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
	rules.check()
	return rules
}

// Rule is a component of rule sets for route validation. Each validated fields
// has one or multiple validation rules. The goal of this struct is to
// gather information about how to use a rule definition for this field.
// This inludes the rule name (referring to a RuleDefinition), the parameters
// and the array dimension for array validation.
type Rule struct {
	Name           string
	Params         []string
	ArrayDimension uint8
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
	Rules      []*Rule
	isArray    bool
	isRequired bool
	isNullable bool
}

// IsRequired check if a field has the "required" rule
func (v *Field) IsRequired() bool {
	return v.isRequired
}

// IsNullable check if a field has the "nullable" rule
func (v *Field) IsNullable() bool {
	return v.isNullable
}

// IsArray check if a field has the "array" rule
func (v *Field) IsArray() bool {
	return v.isArray
}

// check if rules meet the minimum parameters requirement and update
// the isRequired, isNullable and isArray fields.
func (v *Field) check() {
	for _, rule := range v.Rules {
		switch rule.Name {
		case "confirmed", "file", "mime", "image", "extension", "count",
			"count_min", "count_max", "count_between":
			if rule.ArrayDimension != 0 {
				panic(fmt.Sprintf("Cannot use rule \"%s\" in array validation", rule.Name))
			}
		case "required":
			v.isRequired = true
		case "nullable":
			v.isNullable = true
			continue
		case "array":
			v.isArray = true
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
	Fields  FieldMap
	checked bool
}

var _ Ruler = (*Rules)(nil) // implements Ruler

// AsRules performs the checking and returns the same Rules instance.
func (r *Rules) AsRules() *Rules {
	r.check()
	return r
}

// check all rules in this set. This function will panic if
// any of the rules doesn't refer to an existing RuleDefinition, doesn't
// meet the parameters requirement, or if the rule cannot be used in array validation
// while ArrayDimension is not equal to 0.
func (r *Rules) check() {
	if !r.checked {
		for _, field := range r.Fields {
			field.check()
		}
		r.checked = true
	}
}

// Errors is a map of validation errors with the field name as a key.
type Errors map[string][]string

var validationRules map[string]*RuleDefinition

func init() {
	validationRules = map[string]*RuleDefinition{
		"required":           {validateRequired, 0, false, false},
		"numeric":            {validateNumeric, 0, true, false},
		"integer":            {validateInteger, 0, true, false},
		"min":                {validateMin, 1, false, true},
		"max":                {validateMax, 1, false, true},
		"between":            {validateBetween, 2, false, true},
		"greater_than":       {validateGreaterThan, 1, false, true},
		"greater_than_equal": {validateGreaterThanEqual, 1, false, true},
		"lower_than":         {validateLowerThan, 1, false, true},
		"lower_than_equal":   {validateLowerThanEqual, 1, false, true},
		"string":             {validateString, 0, true, false},
		"array":              {validateArray, 0, false, false},
		"distinct":           {validateDistinct, 0, false, false},
		"digits":             {validateDigits, 0, false, false},
		"regex":              {validateRegex, 1, false, false},
		"email":              {validateEmail, 0, false, false},
		"size":               {validateSize, 1, false, true},
		"alpha":              {validateAlpha, 0, false, false},
		"alpha_dash":         {validateAlphaDash, 0, false, false},
		"alpha_num":          {validateAlphaNumeric, 0, false, false},
		"starts_with":        {validateStartsWith, 1, false, false},
		"ends_with":          {validateEndsWith, 1, false, false},
		"in":                 {validateIn, 1, false, false},
		"not_in":             {validateNotIn, 1, false, false},
		"in_array":           {validateInArray, 1, false, false},
		"not_in_array":       {validateNotInArray, 1, false, false},
		"timezone":           {validateTimezone, 0, true, false},
		"ip":                 {validateIP, 0, true, false},
		"ipv4":               {validateIPv4, 0, true, false},
		"ipv6":               {validateIPv6, 0, true, false},
		"json":               {validateJSON, 0, true, false},
		"url":                {validateURL, 0, true, false},
		"uuid":               {validateUUID, 0, true, false},
		"bool":               {validateBool, 0, true, false},
		"same":               {validateSame, 1, false, false},
		"different":          {validateDifferent, 1, false, false},
		"confirmed":          {validateConfirmed, 0, false, false},
		"file":               {validateFile, 0, false, false},
		"mime":               {validateMIME, 1, false, false},
		"image":              {validateImage, 0, false, false},
		"extension":          {validateExtension, 1, false, false},
		"count":              {validateCount, 1, false, false},
		"count_min":          {validateCountMin, 1, false, false},
		"count_max":          {validateCountMax, 1, false, false},
		"count_between":      {validateCountBetween, 2, false, false},
		"date":               {validateDate, 0, true, false},
		"before":             {validateBefore, 1, false, false},
		"before_equal":       {validateBeforeEqual, 1, false, false},
		"after":              {validateAfter, 1, false, false},
		"after_equal":        {validateAfterEqual, 1, false, false},
		"date_equals":        {validateDateEquals, 1, false, false},
		"date_between":       {validateDateBetween, 2, false, false},
		"object":             {validateObject, 0, true, false},
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

	for fieldName, field := range rules.Fields {
		name, fieldVal, parent, _ := GetFieldFromName(fieldName, data)
		if !field.IsNullable() && fieldVal == nil {
			delete(parent, fieldName)
		}

		if !field.IsRequired() && !validateRequired(fieldName, fieldVal, nil, data) {
			continue
		}

		convertArray(isJSON, name, field, parent) // Convert single value arrays in url-encoded requests

		for _, rule := range field.Rules {
			fieldVal = parent[name]
			if rule.Name == "nullable" {
				if fieldVal == nil {
					break
				}
				continue
			}

			if rule.ArrayDimension > 0 {
				if ok, errorValue := validateRuleInArray(rule, fieldName, rule.ArrayDimension, data); !ok {
					errors[fieldName] = append(
						errors[fieldName],
						processPlaceholders(fieldName, rule.Name, rule.Params, getMessage(field.Rules, rule, errorValue, language), language),
					)
				}
			} else if !validationRules[rule.Name].Function(fieldName, fieldVal, rule.Params, data) {
				errors[fieldName] = append(
					errors[fieldName],
					processPlaceholders(fieldName, rule.Name, rule.Params, getMessage(field.Rules, rule, reflect.ValueOf(fieldVal), language), language),
				)
			}
		}
	}
	return errors
}

func validateRuleInArray(rule *Rule, fieldName string, arrayDimension uint8, data map[string]interface{}) (bool, reflect.Value) {
	if t := GetFieldType(data[fieldName]); t != "array" {
		return false, reflect.ValueOf(data[fieldName])
	}

	converted := false
	var convertedArr reflect.Value
	list := reflect.ValueOf(data[fieldName])
	length := list.Len()
	for i := 0; i < length; i++ {
		v := list.Index(i)
		value := v.Interface()
		tmpData := map[string]interface{}{fieldName: value}
		if arrayDimension > 1 {
			ok, errorValue := validateRuleInArray(rule, fieldName, arrayDimension-1, tmpData)
			if !ok {
				return false, errorValue
			}
		} else if !validationRules[rule.Name].Function(fieldName, value, rule.Params, tmpData) {
			return false, v
		}

		// Update original array if value has been modified.
		if rule.Name == "array" {
			if !converted { // Ensure field is a two dimensional array of the correct type
				convertedArr = reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(tmpData[fieldName])), 0, length)
				converted = true
			}
			convertedArr = reflect.Append(convertedArr, reflect.ValueOf(tmpData[fieldName]))
		} else {
			v.Set(reflect.ValueOf(tmpData[fieldName]))
		}
	}

	if converted {
		data[fieldName] = convertedArr.Interface()
	}
	return true, reflect.Value{}
}

func convertArray(isJSON bool, fieldName string, field *Field, data map[string]interface{}) {
	if !isJSON {
		val := data[fieldName]
		rv := reflect.ValueOf(val)
		kind := rv.Kind().String()
		if field.IsArray() && kind != "slice" {
			rt := reflect.TypeOf(val)
			slice := reflect.MakeSlice(reflect.SliceOf(rt), 0, 1)
			slice = reflect.Append(slice, rv)
			data[fieldName] = slice.Interface()
		}
	}
}

func getMessage(rules []*Rule, rule *Rule, value reflect.Value, language string) string {
	langEntry := "validation.rules." + rule.Name
	if validationRules[rule.Name].IsTypeDependent {
		expectedType := findTypeRule(rules, rule.ArrayDimension)
		if expectedType == "unsupported" {
			langEntry += "." + getFieldType(value)
		} else {
			if expectedType == "integer" {
				expectedType = "numeric"
			}
			langEntry += "." + expectedType
		}
	}

	if rule.ArrayDimension > 0 {
		langEntry += ".array"
	}

	return lang.Get(language, langEntry)
}

// findTypeRule find the expected type of a field for a given array dimension.
func findTypeRule(rules []*Rule, arrayDimension uint8) string {
	for _, rule := range rules {
		if rule.ArrayDimension == arrayDimension-1 && rule.Name == "array" && len(rule.Params) > 0 {
			return rule.Params[0]
		} else if rule.ArrayDimension == arrayDimension && validationRules[rule.Name].IsType {
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
	arrayDimensions := uint8(0)
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

	if ruleName[0] == '>' {
		for ruleName[0] == '>' {
			ruleName = ruleName[1:]
			arrayDimensions++
		}
	}

	return &Rule{ruleName, params, arrayDimensions}
}
