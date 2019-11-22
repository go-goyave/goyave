package validation

import (
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/System-Glitch/goyave/helper"

	"github.com/System-Glitch/goyave/lang"
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

var validationRules map[string]Rule = map[string]Rule{
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

var typeDependentMessageRules []string = []string{
	"min", "max", "size",
	"greater_than", "greater_than_equal",
	"lower_than", "lower_than_equal",
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

// Validate the given request with the given rule set
// If all validation rules pass, returns nil
func Validate(request *http.Request, data map[string]interface{}, rules RuleSet, language string) Errors {
	var malformedMessage string
	if request.Header.Get("Content-Type") == "application/json" {
		malformedMessage = "Malformed JSON"
	} else {
		malformedMessage = "Malformed request"
	}
	if data == nil {
		return map[string][]string{"error": {malformedMessage}}
	}

	return validate(request, data, rules, language)
}

func validate(request *http.Request, data map[string]interface{}, rules RuleSet, language string) Errors {
	errors := Errors{}
	isJSON := request.Header.Get("Content-Type") == "application/json"

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
			ruleName, params := parseRule(rule)
			if !validationRules[ruleName](fieldName, data[fieldName], params, data) {
				message := processPlaceholders(fieldName, ruleName, params, getMessage(ruleName, data[fieldName], language), language)
				errors[fieldName] = append(errors[fieldName], message)
			}
		}
	}
	return errors
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

func getMessage(rule string, value interface{}, language string) string {
	langEntry := "validation.rules." + rule
	if isTypeDependent(rule) {
		langEntry = langEntry + "." + getFieldType(value)
	}
	return lang.Get(language, langEntry)
}

func getFieldType(value interface{}) string {
	rv := reflect.ValueOf(value)
	kind := rv.Kind().String()
	switch {
	case strings.HasPrefix(kind, "int"), strings.HasPrefix(kind, "uint") && kind != "uintptr", strings.HasPrefix(kind, "float"):
		return "numeric"
	case kind == "string":
		return "string"
	case kind == "slice":
		if rv.Type().String() == "[]filesystem.File" {
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

func isRequired(field []string) bool {
	return helper.Contains(field, "required")
}

func isNullable(field []string) bool {
	return helper.Contains(field, "nullable")
}

func isArray(field []string) bool {
	return helper.Contains(field, "array")
}

func parseRule(rule string) (string, []string) {
	indexName := strings.Index(rule, ":")
	params := []string{}
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

	if _, exists := validationRules[ruleName]; !exists {
		log.Panicf("Rule \"%s\" doesn't exist", ruleName)
	}

	return ruleName, params
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
