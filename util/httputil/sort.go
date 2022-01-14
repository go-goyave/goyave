package httputil

import "strings"

type byPriority []HeaderValue

func (s byPriority) Len() int {
	return len(s)
}
func (s byPriority) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byPriority) Less(i, j int) bool {
	if s[j].Priority == s[i].Priority {
		return specificity(s[j]) < specificity(s[i])
	}

	return s[j].Priority < s[i].Priority
}

func specificity(value HeaderValue) int {
	return strings.Count(value.Value, "-") + strings.Count(value.Value, "/") - strings.Count(value.Value, "*")
}
