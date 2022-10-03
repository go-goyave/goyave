package httputil

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// HeaderValue represent a value and its quality value (priority)
// in a multi-values HTTP header.
type HeaderValue struct {
	Value    string
	Priority float64
}

var multiValuesHeaderRegex *regexp.Regexp

// ParseMultiValuesHeader parses multi-values HTTP headers, taking the
// quality values into account. The result is a slice of values sorted
// according to the order of priority.
//
// See: https://developer.mozilla.org/en-US/docs/Glossary/Quality_values
//
// For the following header:
//
//	"text/html,text/*;q=0.5,*/*;q=0.7"
//
// returns
//
//	[{text/html 1} {*/* 0.7} {text/* 0.5}]
func ParseMultiValuesHeader(header string) []HeaderValue {
	if multiValuesHeaderRegex == nil {
		multiValuesHeaderRegex = regexp.MustCompile(`^q=([01]\.[0-9]{1,3})$`)
	}
	split := strings.Split(header, ",")
	values := make([]HeaderValue, 0, len(split))

	for _, v := range split {
		val := HeaderValue{}
		if i := strings.Index(v, ";"); i != -1 {
			// Parse priority
			q := v[i+1:]

			sub := multiValuesHeaderRegex.FindStringSubmatch(q)
			priority := 0.0
			if len(sub) > 1 {
				if p, err := strconv.ParseFloat(sub[1], 64); err == nil {
					priority = p
				}
			}
			// Priority set to 0 if the quality value cannot be parsed
			val.Priority = priority

			val.Value = strings.Trim(v[:i], " ")
		} else {
			val.Value = strings.Trim(v, " ")
			val.Priority = 1
		}

		values = append(values, val)
	}

	sort.Sort(byPriority(values))

	return values
}
