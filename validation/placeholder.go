package validation

import (
	"sort"
	"strings"
	"time"

	"goyave.dev/goyave/v4/lang"
)

// Placeholder function defining a placeholder in a validation message.
// This function should return the value to replace the placeholder with.
type Placeholder func(fieldName string, language string, ctx *Context) string

var placeholders = map[string]Placeholder{}
var sortedKeys = []string{}

// SetPlaceholder sets the replacer function for the given placeholder.
// If a placeholder with this name already exists, the latter will be overridden.
//
//	validation.SetPlaceholder("min", func(field string, rule string, parameters []string, language string) string {
//		return parameters[0] // Replace ":min" by the first parameter in the rule definition
//	})
func SetPlaceholder(placeholderName string, replacer Placeholder) {
	key := ":" + placeholderName
	placeholders[key] = replacer

	// Sort keys to process placeholders in order.
	// Needed to avoid conflict between "values" and "value" for example.
	sortedKeys = append(sortedKeys, key)
	sort.Sort(sort.Reverse(sort.StringSlice(sortedKeys)))
}

func processPlaceholders(fieldName string, originalMessage string, language string, ctx *Context) string {
	if i := strings.LastIndex(fieldName, "."); i != -1 {
		fieldName = fieldName[i+1:]
	}
	fieldName = strings.TrimSuffix(fieldName, "[]")
	message := originalMessage
	for _, placeholder := range sortedKeys {
		if strings.Contains(originalMessage, placeholder) {
			replacer := placeholders[placeholder]
			message = strings.ReplaceAll(message, placeholder, replacer(fieldName, language, ctx))
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

func simpleParameterPlaceholder(_ string, _ string, ctx *Context) string {
	return ctx.Rule.Params[0]
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
	SetPlaceholder("field", func(field string, language string, ctx *Context) string {
		return replaceField(field, language)
	})
	SetPlaceholder("value", simpleParameterPlaceholder)
	SetPlaceholder("min", simpleParameterPlaceholder)
	SetPlaceholder("max", func(field string, language string, ctx *Context) string {
		index := 0
		if strings.Contains(ctx.Rule.Name, "between") {
			index = 1
		}
		return ctx.Rule.Params[index]
	})
	SetPlaceholder("other", func(field string, language string, ctx *Context) string {
		return replaceField(ctx.Rule.Params[0], language)
	})
	SetPlaceholder("values", func(field string, language string, ctx *Context) string {
		return strings.Join(ctx.Rule.Params, ", ")
	})
	SetPlaceholder("version", func(field string, language string, ctx *Context) string {
		if len(ctx.Rule.Params) > 0 {
			return "v" + ctx.Rule.Params[0]
		}
		return ""
	})
	SetPlaceholder("date", func(field string, language string, ctx *Context) string {
		return datePlaceholder(0, ctx.Rule.Params, language)
	})
	SetPlaceholder("max_date", func(field string, language string, ctx *Context) string {
		return datePlaceholder(1, ctx.Rule.Params, language)
	})
}
