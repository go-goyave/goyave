package validation

import (
	"log"
	"reflect"
	"strings"

	"github.com/System-Glitch/goyave/v2/helper"

	"github.com/System-Glitch/goyave/v2/lang"
)

// Rule function defining a validation rule.
// Passing rules should return true, false otherwise.
//
// Rules can modifiy the validated value if needed.
// For example, the "numeric" rule converts the data to float64 if it's a string.
type Rule func(string, interface{}, []string, map[string]interface{}) bool

// RuleSet is a request rules definition. Each entry is a field in the request.
type RuleSet map[string][]string

// Errors is a map of validation errors with the field name as a key.
type Errors map[string][]string

var (
	validationRules           map[string]Rule
	typeDependentMessageRules []string

	// Rules that check the data type and can be used to validate arrays.
	typeRules []string
)

func init() {
	validationRules = map[string]Rule{
		"required":           validateRequired,
		"numeric":            validateNumeric,
		"integer":            validateInteger,
		"min":                validateMin,
		"max":                validateMax,
		"between":            validateBetween,
		"greater_than":       validateGreaterThan,
		"greater_than_equal": validateGreaterThanEqual,
		"lower_than":         validateLowerThan,
		"lower_than_equal":   validateLowerThanEqual,
		"string":             validateString,
		"array":              validateArray,
		"distinct":           validateDistinct,
		"digits":             validateDigits,
		"regex":              validateRegex,
		"email":              validateEmail,
		"size":               validateSize,
		"alpha":              validateAlpha,
		"alpha_dash":         validateAlphaDash,
		"alpha_num":          validateAlphaNumeric,
		"starts_with":        validateStartsWith,
		"ends_with":          validateEndsWith,
		"in":                 validateIn,
		"not_in":             validateNotIn,
		"in_array":           validateInArray,
		"not_in_array":       validateNotInArray,
		"timezone":           validateTimezone,
		"ip":                 validateIP,
		"ipv4":               validateIPv4,
		"ipv6":               validateIPv6,
		"json":               validateJSON,
		"url":                validateURL,
		"uuid":               validateUUID,
		"bool":               validateBool,
		"same":               validateSame,
		"different":          validateDifferent,
		"confirmed":          validateConfirmed,
		"file":               validateFile,
		"mime":               validateMIME,
		"image":              validateImage,
		"extension":          validateExtension,
		"count":              validateCount,
		"count_min":          validateCountMin,
		"count_max":          validateCountMax,
		"count_between":      validateCountBetween,
		"date":               validateDate,
		"before":             validateBefore,
		"before_equal":       validateBeforeEqual,
		"after":              validateAfter,
		"after_equal":        validateAfterEqual,
		"date_equals":        validateDateEquals,
		"date_between":       validateDateBetween,
	}

	typeDependentMessageRules = []string{
		"min", "max", "between", "size",
		"greater_than", "greater_than_equal",
		"lower_than", "lower_than_equal",
	}

	typeRules = []string{
		"numeric", "integer", "timezone", "ip",
		"ipv4", "ipv6", "json", "url", "uuid",
		"bool", "date", "string",
	}
}

// AddRule register a validation rule.
// The rule will be usable in request validation by using the
// given rule name.
//
// Type-dependent messages let you define a different message for
// numeric, string, arrays and files.
// The language entry used will be "validation.rules.rulename.type"
func AddRule(name string, typeDependentMessage bool, rule Rule) {
	if _, exists := validationRules[name]; exists {
		log.Panicf("Rule %s already exists", name)
	}
	validationRules[name] = rule

	if typeDependentMessage {
		typeDependentMessageRules = append(typeDependentMessageRules, name)
	}
}

// Validate the given data with the given rule set.
// If all validation rules pass, returns an empty "validation.Errors".
// Third parameter tells the function if the data comes from a JSON request.
// Last parameter sets the language of the validation error messages.
func Validate(data map[string]interface{}, rules RuleSet, isJSON bool, language string) Errors {
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

func validate(data map[string]interface{}, isJSON bool, rules RuleSet, language string) Errors {
	errors := Errors{}

	for fieldName, field := range rules {
		if !isNullable(field) && data[fieldName] == nil {
			delete(data, fieldName)
		}

		if !isRequired(field) && !validateRequired(fieldName, data[fieldName], []string{}, data) {
			continue
		}

		convertArray(isJSON, fieldName, field, data) // Convert single value arrays in url-encoded requests

		for _, rule := range field {
			if rule == "nullable" {
				if data[fieldName] == nil {
					break
				}
				continue
			}
			ruleName, arrayDimensions, params := parseRule(rule)

			if arrayDimensions > 0 {
				if ok, errorValue := validateRuleInArray(ruleName, fieldName, arrayDimensions, data, params); !ok {
					errors[fieldName] = append(
						errors[fieldName],
						processPlaceholders(fieldName, ruleName, params, getMessage(ruleName, *errorValue, language, arrayDimensions), language),
					)
				}
			} else if !validationRules[ruleName](fieldName, data[fieldName], params, data) {
				errors[fieldName] = append(
					errors[fieldName],
					processPlaceholders(fieldName, ruleName, params, getMessage(ruleName, reflect.ValueOf(data[fieldName]), language, arrayDimensions), language),
				)
			}
		}
	}
	return errors
}

func validateRuleInArray(ruleName, fieldName string, arrayDimensions uint8, data map[string]interface{}, params []string) (bool, *reflect.Value) {
	if t := GetFieldType(data[fieldName]); t != "array" {
		log.Panicf("Cannot validate array values on non-array field %s of type %s", fieldName, t)
	}

	converted := false
	var convertedArr reflect.Value
	list := reflect.ValueOf(data[fieldName])
	length := list.Len()
	for i := 0; i < length; i++ {
		v := list.Index(i)
		value := v.Interface()
		tmpData := map[string]interface{}{fieldName: value}
		if arrayDimensions > 1 {
			ok, errorValue := validateRuleInArray(ruleName, fieldName, arrayDimensions-1, tmpData, params)
			if !ok {
				return false, errorValue
			}
		} else if !validationRules[ruleName](fieldName, value, params, tmpData) {
			return false, &v
		}

		// Update original array if value has been modified.
		if ruleName == "array" {
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

func convertArray(isJSON bool, fieldName string, field []string, data map[string]interface{}) {
	if !isJSON {
		val := data[fieldName]
		rv := reflect.ValueOf(val)
		kind := rv.Kind().String()
		if isArray(field) && kind != "slice" {
			rt := reflect.TypeOf(val)
			slice := reflect.MakeSlice(reflect.SliceOf(rt), 0, 1)
			slice = reflect.Append(slice, rv)
			data[fieldName] = slice.Interface()
		}
	}
}

func getMessage(rule string, value reflect.Value, language string, arrayDimensions uint8) string {
	langEntry := "validation.rules." + rule
	if isTypeDependent(rule) {
		langEntry += "." + getFieldType(value)
	}

	if arrayDimensions > 0 {
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

func isTypeDependent(rule string) bool {
	return helper.Contains(typeDependentMessageRules, rule)
}

func isArrayType(rule string) bool {
	return helper.Contains(typeRules, rule)
}

func isRequired(field []string) bool {
	return helper.Contains(field, "required")
}

func isNullable(field []string) bool {
	return helper.Contains(field, "nullable")
}

func isArray(field []string) bool {
	return helper.Contains(field, "array")
}

func parseRule(rule string) (string, uint8, []string) {
	indexName := strings.Index(rule, ":")
	params := []string{}
	validatesArray := uint8(0)
	var ruleName string
	if indexName == -1 {
		if strings.Count(rule, ",") > 0 {
			log.Panicf("Invalid rule: \"%s\"", rule)
		}
		ruleName = rule
	} else {
		ruleName = rule[:indexName]
		params = strings.Split(rule[indexName+1:], ",") // TODO how to escape comma?
	}

	if ruleName[0] == '>' {
		for ruleName[0] == '>' {
			ruleName = ruleName[1:]
			validatesArray++
		}

		switch ruleName {
		case "confirmed", "file", "mime", "image", "extension", "count",
			"count_min", "count_max", "count_between":
			log.Panicf("Cannot use rule \"%s\" in array validation", ruleName)
		}

	}

	if _, exists := validationRules[ruleName]; !exists {
		log.Panicf("Rule \"%s\" doesn't exist", ruleName)
	}

	return ruleName, validatesArray, params
}

// RequireParametersCount checks if the given parameters slice has at least "count" elements.
// If this is not the case, panics.
//
// Use this to make sure your validation rules are correctly used.
func RequireParametersCount(rule string, params []string, count int) {
	if len(params) < count {
		log.Panicf("Rule \"%s\" requires %d parameter(s)", rule, count)
	}
}
