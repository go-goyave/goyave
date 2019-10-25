package validation

// Rule function defining a validation rule.
// Passing rules should return true, false otherwise.
type Rule func(string, interface{}, []string, map[string]interface{}) bool

func validateString(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	_, ok := value.(string)
	return ok
}
