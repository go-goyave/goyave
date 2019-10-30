package validation

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/System-Glitch/goyave/helpers"

	"github.com/System-Glitch/goyave/lang"
)

// RuleSet is a request rules definition. Each entry is a field in the request.
type RuleSet map[string][]string

// Errors is a map of validation errors with the field name as a key.
type Errors map[string][]string

var validationRules map[string]Rule = map[string]Rule{
	"required": validateRequired,
	"numeric":  validateNumeric,
	"integer":  validateInteger,
	"min":      validateMin,
	"max":      validateMax,
	"between":  validateBetween,
	"string":   validateString,
}

var typeDependentMessageRules []string = []string{"min", "max"}

// AddRule register a validation rule.
// The rule will be usable in request validation by using the
// given rule name.
//
// Type-dependent messages let you define a different message for
// numeric, string, arrays and files.
// The language entry used will be "validation.rules.rulename.type"
func AddRule(name string, typeDependentMessage bool, rule Rule) {
	if _, exists := validationRules[name]; exists {
		panic(fmt.Errorf("Rule %s already exists", name))
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
		return map[string][]string{"_error": {malformedMessage}}
	}

	return validate(data, rules, language)
}

func validate(data map[string]interface{}, rules RuleSet, language string) Errors {
	errors := Errors{}
	for fieldName, field := range rules {
		// TODO document that if field is not required and is missing, don't check rules
		if !isRequired(field) && !validateRequired(fieldName, data[fieldName], []string{}, data) {
			continue
		}
		for _, rule := range field {
			ruleName, params := parseRule(rule)
			if !validationRules[ruleName](fieldName, data[fieldName], params, data) {
				message := processPlaceholders(fieldName, ruleName, params, getMessage(ruleName, data[fieldName], language), language)
				errors[fieldName] = append(errors[fieldName], message)
			}
		}
	}
	return errors
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
		return "array"
	// TODO file field type
	default:
		return "unsupported"
	}
}

func isTypeDependent(rule string) bool {
	return helpers.Contains(typeDependentMessageRules, rule)
}

func isRequired(field []string) bool {
	return helpers.Contains(field, "required")
}

func parseRule(rule string) (string, []string) {
	indexName := strings.Index(rule, ":")
	params := []string{}
	var ruleName string
	if indexName == -1 {
		if strings.Count(rule, ",") > 0 {
			panic(fmt.Errorf("Invalid rule: \"%s\"", rule))
		}
		ruleName = rule
	} else {
		ruleName = rule[:indexName]
		params = strings.Split(rule[indexName+1:], ",")
	}

	if _, exists := validationRules[ruleName]; !exists {
		panic(fmt.Errorf("Rule \"%s\" doesn't exist", ruleName))
	}

	return ruleName, params
}
