package validation

import (
	"strings"

	"github.com/System-Glitch/goyave/lang"
)

// Placeholder function defining a placeholder in a validation message.
// This function should return the value to replace the placeholder with.
type Placeholder func(string, string, []string, string) string

var placeholders map[string]Placeholder = map[string]Placeholder{}

// SetPlaceholder sets the replacer function for the given placeholder.
// If a placeholder with this name already exists, the latter will be overridden.
//  validation.SetPlaceholder("min", func(field string, rule string, parameters []string, language string) string {
//  	return parameters[1] // Replace ":min" by the second parameter in the rule definition
//  })
func SetPlaceholder(placeholderName string, replacer Placeholder) {
	placeholders[":"+placeholderName] = replacer
}

func processPlaceholders(field string, rule string, params []string, message string, language string) string {
	for placeholder, replacer := range placeholders {
		message = strings.ReplaceAll(message, placeholder, replacer(field, rule, params, language))
	}
	return message
}

func simpleParameterPlaceholder(field string, rule string, parameters []string, language string) string {
	return parameters[0]
}

func init() {
	SetPlaceholder("field", func(field string, rule string, parameters []string, language string) string {
		entry := "validation.fields." + field
		attr := lang.Get(language, entry)
		if attr == entry {
			return field
		}
		return attr
	})
	SetPlaceholder("min", simpleParameterPlaceholder)
	SetPlaceholder("max", simpleParameterPlaceholder)
	// TODO set more placeholders
}
