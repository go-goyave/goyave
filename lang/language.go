package lang

import "strings"

type validationLines struct {
	// Default messages for rules
	rules map[string]string

	// Field names translations
	fields map[string]string
}

// Language represents a full Language.
type Language struct {
	lines      map[string]string
	validation validationLines
	name       string
}

// Name returns the name of the language. For example "en-US".
func (l *Language) Name() string {
	return l.name
}

func (l *Language) clone() *Language {
	cpy := &Language{
		name:  l.name,
		lines: make(map[string]string, len(l.lines)),
		validation: validationLines{
			rules:  make(map[string]string, len(l.validation.rules)),
			fields: make(map[string]string, len(l.validation.fields)),
		},
	}

	mergeMap(cpy.lines, l.lines)
	mergeMap(cpy.validation.rules, l.validation.rules)
	mergeMap(cpy.validation.fields, l.validation.fields)

	return cpy
}

// Get a language line.
//
// For validation rules messages and field names, use a dot-separated path:
//   - "validation.rules.<rule_name>"
//   - "validation.fields.<field_name>"
//
// For normal lines, just use the name of the line. Note that if you have
// a line called "validation", it won't conflict with the dot-separated paths.
//
// If not found, returns the exact "line" argument.
//
// The placeholders parameter is a variadic associative slice of placeholders and their
// replacement. In the following example, the placeholder ":username" will be replaced
// with the Name field in the user struct.
//
//	lang.Get("greetings", ":username", user.Name)
func (l *Language) Get(line string, placeholders ...string) string {
	if strings.HasPrefix(line, "validation.rules.") {
		return convertEmptyLine(line, l.validation.rules[line[17:]], placeholders)
	} else if strings.HasPrefix(line, "validation.fields.") {
		return convertEmptyLine(line, l.validation.fields[line[18:]], placeholders)
	}

	return convertEmptyLine(line, l.lines[line], placeholders)
}

func convertEmptyLine(entry, line string, placeholders []string) string {
	if line == "" {
		return entry
	}
	return processPlaceholders(line, placeholders)
}

func processPlaceholders(message string, values []string) string {
	length := len(values) - 1
	result := message
	for i := 0; i < length; i += 2 {
		if strings.Contains(message, values[i]) {
			result = strings.ReplaceAll(result, values[i], values[i+1])
		}
	}
	return result
}
