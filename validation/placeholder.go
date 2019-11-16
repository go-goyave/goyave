package validation

import (
	"strings"
	"time"

	"github.com/System-Glitch/goyave/lang"
)

// Placeholder function defining a placeholder in a validation message.
// This function should return the value to replace the placeholder with.
type Placeholder func(string, string, []string, string) string

var placeholders map[string]Placeholder = map[string]Placeholder{}

// SetPlaceholder sets the replacer function for the given placeholder.
// If a placeholder with this name already exists, the latter will be overridden.
//  validation.SetPlaceholder("min", func(field string, rule string, parameters []string, language string) string {
//  	return parameters[0] // Replace ":min" by the first parameter in the rule definition
//  })
func SetPlaceholder(placeholderName string, replacer Placeholder) {
	placeholders[":"+placeholderName] = replacer
}

func processPlaceholders(field string, rule string, params []string, message string, language string) string {
	for placeholder, replacer := range placeholders {
		if strings.Contains(message, placeholder) {
			message = strings.ReplaceAll(message, placeholder, replacer(field, rule, params, language))
		}
	}
	return message
}

func replaceField(field, language string) string {
	entry := "validation.fields." + field
	attr := lang.Get(language, entry)
	if attr == entry {
		return field
	}
	return attr
}

func simpleParameterPlaceholder(field string, rule string, parameters []string, language string) string {
	return parameters[0]
}

func datePlaceholder(index int, parameters []string, language string) string {
	_, err := time.Parse("2006-01-02T15:04:05", parameters[index])
	if err != nil {
		// Not a date, may be a field
		return replaceField(parameters[index], language)
	}
	return parameters[index]
}

func init() {
	SetPlaceholder("field", func(field string, rule string, parameters []string, language string) string {
		return replaceField(field, language)
	})
	SetPlaceholder("value", simpleParameterPlaceholder)
	SetPlaceholder("min", simpleParameterPlaceholder)
	SetPlaceholder("max", func(field string, rule string, parameters []string, language string) string {
		index := 0
		if strings.Contains(rule, "between") {
			index = 1
		}
		return parameters[index]
	})
	SetPlaceholder("other", func(field string, rule string, parameters []string, language string) string {
		return replaceField(parameters[0], language)
	})
	SetPlaceholder("values", func(field string, rule string, parameters []string, language string) string {
		return strings.Join(parameters, ", ")
	})
	SetPlaceholder("version", func(field string, rule string, parameters []string, language string) string {
		if len(parameters) > 0 {
			return "v" + parameters[0]
		}
		return ""
	})
	SetPlaceholder("date", func(field string, rule string, parameters []string, language string) string {
		return datePlaceholder(0, parameters, language)
	})
	SetPlaceholder("max_date", func(field string, rule string, parameters []string, language string) string {
		return datePlaceholder(1, parameters, language)
	})
}
