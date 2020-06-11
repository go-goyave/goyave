package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/System-Glitch/goyave/v2/lang"
)

// RuleFunc function defining a validation rule.
// Passing rules should return true, false otherwise.
//
// Rules can modifiy the validated value if needed.
// For example, the "numeric" rule converts the data to float64 if it's a string.
type RuleFunc func(string, interface{}, []string, map[string]interface{}) bool

// RuleDefinition TODO document this
type RuleDefinition struct {
	Function        RuleFunc
	IsType          bool
	IsTypeDependent bool
}

// RuleSet is a request rules definition. Each entry is a field in the request.
type RuleSet map[string][]string

// ValidatedField TODO document this
type ValidatedField struct {
	Rules      []*Rule
	isArray    bool // TODO these fields are never set when using verbose declaration
	isRequired bool
	isNullable bool
}

// TODO temporary solution for verbose declaration
func (v *ValidatedField) is(rule string) bool {
	for _, r := range v.Rules {
		if r.Name == rule {
			return true
		}
	}
	return false
}

// IsRequired check if a field has the "required" rule
func (v *ValidatedField) IsRequired() bool {
	// return v.isRequired
	return v.is("required")
}

// IsNullable check if a field has the "nullable" rule
func (v *ValidatedField) IsNullable() bool {
	// return v.isNullable
	return v.is("nullable")
}

// IsArray check if a field has the "array" rule
func (v *ValidatedField) IsArray() bool {
	// return v.isArray
	return v.is("array")
}

// Rule TODO document this
type Rule struct {
	Name           string
	Params         []string
	ArrayDimension uint8
}

// Rules TODO document this
type Rules map[string]*ValidatedField

// Errors is a map of validation errors with the field name as a key.
type Errors map[string][]string

var validationRules map[string]*RuleDefinition

func init() {
	validationRules = map[string]*RuleDefinition{
		"required":           {validateRequired, false, false},
		"numeric":            {validateNumeric, true, false},
		"integer":            {validateInteger, true, false},
		"min":                {validateMin, false, true},
		"max":                {validateMax, false, true},
		"between":            {validateBetween, false, true},
		"greater_than":       {validateGreaterThan, false, true},
		"greater_than_equal": {validateGreaterThanEqual, false, true},
		"lower_than":         {validateLowerThan, false, true},
		"lower_than_equal":   {validateLowerThanEqual, false, true},
		"string":             {validateString, true, false},
		"array":              {validateArray, false, false},
		"distinct":           {validateDistinct, false, false},
		"digits":             {validateDigits, false, false},
		"regex":              {validateRegex, false, false},
		"email":              {validateEmail, false, false},
		"size":               {validateSize, false, false},
		"alpha":              {validateAlpha, false, false},
		"alpha_dash":         {validateAlphaDash, false, false},
		"alpha_num":          {validateAlphaNumeric, false, false},
		"starts_with":        {validateStartsWith, false, false},
		"ends_with":          {validateEndsWith, false, false},
		"in":                 {validateIn, false, false},
		"not_in":             {validateNotIn, false, false},
		"in_array":           {validateInArray, false, false},
		"not_in_array":       {validateNotInArray, false, false},
		"timezone":           {validateTimezone, true, false},
		"ip":                 {validateIP, true, false},
		"ipv4":               {validateIPv4, true, false},
		"ipv6":               {validateIPv6, true, false},
		"json":               {validateJSON, true, false},
		"url":                {validateURL, true, false},
		"uuid":               {validateUUID, true, false},
		"bool":               {validateBool, true, false},
		"same":               {validateSame, false, false},
		"different":          {validateDifferent, false, false},
		"confirmed":          {validateConfirmed, false, false},
		"file":               {validateFile, false, false},
		"mime":               {validateMIME, false, false},
		"image":              {validateImage, false, false},
		"extension":          {validateExtension, false, false},
		"count":              {validateCount, false, false},
		"count_min":          {validateCountMin, false, false},
		"count_max":          {validateCountMax, false, false},
		"count_between":      {validateCountBetween, false, false},
		"date":               {validateDate, true, false},
		"before":             {validateBefore, false, false},
		"before_equal":       {validateBeforeEqual, false, false},
		"after":              {validateAfter, false, false},
		"after_equal":        {validateAfterEqual, false, false},
		"date_equals":        {validateDateEquals, false, false},
		"date_between":       {validateDateBetween, false, false},
	}
}

// AddRule register a validation rule.
// The rule will be usable in request validation by using the
// given rule name.
//
// Type-dependent messages let you define a different message for
// numeric, string, arrays and files.
// The language entry used will be "validation.rules.rulename.type"
func AddRule(name string, rule *RuleDefinition) { // TODO update documentation
	if _, exists := validationRules[name]; exists {
		panic(fmt.Sprintf("Rule %s already exists", name))
	}
	validationRules[name] = rule
}

// Validate the given data with the given rule set.
// If all validation rules pass, returns an empty "validation.Errors".
// Third parameter tells the function if the data comes from a JSON request.
// Last parameter sets the language of the validation error messages.
func Validate(data map[string]interface{}, rules Rules, isJSON bool, language string) Errors {
	var malformedMessage string
	if isJSON {
		malformedMessage = lang.Get(language, "malformed-json")
	} else {
		malformedMessage = lang.Get(language, "malformed-request")
	}
	if data == nil {
		return map[string][]string{"error": {malformedMessage}}
	}

	return validate(data, isJSON, rules, language)
}

func validate(data map[string]interface{}, isJSON bool, rules Rules, language string) Errors {
	errors := Errors{}

	for fieldName, field := range rules {
		fmt.Println(fieldName)
		if !field.IsNullable() && data[fieldName] == nil {
			delete(data, fieldName)
		}

		if !field.IsRequired() && !validateRequired(fieldName, data[fieldName], []string{}, data) {
			continue
		}

		convertArray(isJSON, fieldName, field, data) // Convert single value arrays in url-encoded requests

		for _, rule := range field.Rules {
			// TODO better nullable tests
			if rule.Name == "nullable" {
				if data[fieldName] == nil {
					break
				}
				continue
			}

			if rule.ArrayDimension > 0 {
				if ok, errorValue := validateRuleInArray(rule, fieldName, rule.ArrayDimension, data); !ok {
					errors[fieldName] = append(
						errors[fieldName],
						processPlaceholders(fieldName, rule.Name, rule.Params, getMessage(rule, *errorValue, language), language),
					)
				}
			} else if !validationRules[rule.Name].Function(fieldName, data[fieldName], rule.Params, data) {
				errors[fieldName] = append(
					errors[fieldName],
					processPlaceholders(fieldName, rule.Name, rule.Params, getMessage(rule, reflect.ValueOf(data[fieldName]), language), language),
				)
			}
		}
	}
	return errors
}

func validateRuleInArray(rule *Rule, fieldName string, arrayDimension uint8, data map[string]interface{}) (bool, *reflect.Value) {
	if t := GetFieldType(data[fieldName]); t != "array" {
		panic(fmt.Sprintf("Cannot validate array values on non-array field %s of type %s", fieldName, t))
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
			return false, &v
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
	return true, nil
}

func convertArray(isJSON bool, fieldName string, field *ValidatedField, data map[string]interface{}) {
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

func getMessage(rule *Rule, value reflect.Value, language string) string {
	langEntry := "validation.rules." + rule.Name
	if validationRules[rule.Name].IsTypeDependent {
		langEntry += "." + getFieldType(value)
	}

	if rule.ArrayDimension > 0 {
		langEntry += ".array"
	}

	return lang.Get(language, langEntry)
}

// GetFieldType returns the non-technical type of the given "value" interface.
// This is used by validation rules to know if the input data is a candidate
// for validation or not and is especially useful for type-dependent rules.
// - "numeric" if the value is an int, uint or a float
// - "string" if the value is a string
// - "array" if the value is a slice
// - "file" if the value is a slice of "filesystem.File"
// - "unsupported" otherwise
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
		return "unsupported"
	}
}

// ParseRuleSet TODO document this
func ParseRuleSet(set RuleSet) Rules { // TODO test this
	rules := make(Rules, len(set))
	for k, r := range set {
		field := &ValidatedField{
			Rules: make([]*Rule, 0, len(r)),
		}
		for _, v := range r {
			rule := parseInternalRule(v)
			switch rule.Name {
			case "array":
				field.isArray = true
			case "required":
				field.isRequired = true
			case "nullable":
				field.isNullable = true
			}
			field.Rules = append(field.Rules, rule)
		}
		rules[k] = field
	}
	return rules
}

func parseInternalRule(rule string) *Rule { // TODO refactor parseRule to directly return *Rule
	name, arrDims, params := parseRule(rule)
	return &Rule{name, params, arrDims}
}

func parseRule(rule string) (string, uint8, []string) {
	indexName := strings.Index(rule, ":")
	params := []string{}
	validatesArray := uint8(0)
	var ruleName string
	if indexName == -1 {
		if strings.Count(rule, ",") > 0 {
			panic(fmt.Sprintf("Invalid rule: \"%s\"", rule))
		}
		ruleName = rule
	} else {
		ruleName = rule[:indexName]
		params = strings.Split(rule[indexName+1:], ",") // TODO how to escape comma? -> with verbose syntax
	}

	if ruleName[0] == '>' {
		for ruleName[0] == '>' {
			ruleName = ruleName[1:]
			validatesArray++
		}

		switch ruleName {
		case "confirmed", "file", "mime", "image", "extension", "count",
			"count_min", "count_max", "count_between":
			panic(fmt.Sprintf("Cannot use rule \"%s\" in array validation", ruleName))
		}

	}

	if ruleName != "nullable" {
		if _, exists := validationRules[ruleName]; !exists {
			panic(fmt.Sprintf("Rule \"%s\" doesn't exist", ruleName))
		}
	}

	return ruleName, validatesArray, params
}

// RequireParametersCount checks if the given parameters slice has at least "count" elements.
// If this is not the case, panics.
//
// Use this to make sure your validation rules are correctly used.
func RequireParametersCount(rule string, params []string, count int) { // TODO check params count at startup
	if len(params) < count {
		panic(fmt.Sprintf("Rule \"%s\" requires %d parameter(s)", rule, count))
	}
}
