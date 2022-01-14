package sqlutil

import "strings"

// EscapeLike escape "%" and "_" characters in the given string
// for use in SQL "LIKE" clauses.
func EscapeLike(str string) string {
	escapeChars := []string{"%", "_"}
	for _, v := range escapeChars {
		str = strings.ReplaceAll(str, v, "\\"+v)
	}
	return str
}
