package validation

import (
	"fmt"
	"net/http"
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
	"string":   validateString,
}

// AddRule register a validation rule.
// The rule will be usable in request validation by using the
// given rule name.
func AddRule(name string, rule Rule) {
	if _, exists := validationRules[name]; exists {
		panic(fmt.Errorf("Rule %s already exists", name))
	}
	validationRules[name] = rule
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
				message := processPlaceholders(fieldName, ruleName, params, lang.Get(language, "validation.rules."+ruleName), language)
				errors[fieldName] = append(errors[fieldName], message)
			}
		}
	}
	return errors
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
