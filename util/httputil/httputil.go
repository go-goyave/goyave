package httputil

import (
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

// ParseMultiValuesHeader parses multi-values HTTP headers, taking the
// quality values into account. The result is a slice of values sorted
// according to the order of priority.
// If the input is empty, returns an empty slice.
//
// The priority of a value is the value of its "q" parameter, the quality
// value defined by RFC 9110 section 12.4.2: a number between 0 and 1 with
// at most three digits after the decimal point. Values without a "q"
// parameter have a priority of 1, the default given by the RFC. If the
// quality value cannot be parsed, the priority is 0.
//
// Parameters other than "q" are supported, for example the "charset" of a
// media type. They are ignored: they don't change the priority, and they are
// not part of the returned value, which only contains what is located before
// the first ";".
//
// The input and each segment are trimmed. This includes the header value, the
// parameter name and the value of the "q" parameter.
// "text/html ; q = 0.5" is therefore parsed the same way
// as "text/html;q=0.5".
//
// The "q" parameter name is case-insensitive. "Q=0.5" is accepted.
//
// See:
//
//   - https://datatracker.ietf.org/doc/html/rfc9110#section-12.4.2
//   - https://developer.mozilla.org/en-US/docs/Glossary/Quality_values
//
// For the following header value:
//
//	"text/html,text/*;q=0.5,*/*;q=0.7"
//
// returns
//
//	[{text/html 1} {*/* 0.7} {text/* 0.5}]
func ParseMultiValuesHeader(header string) []HeaderValue {
	count := strings.Count(header, ",")
	values := make([]HeaderValue, 0, count+1)

	h := strings.TrimSpace(header)
	if h == "" {
		return values
	}
	for {
		comma := strings.Index(h, ",")
		if comma == -1 {
			comma = len(h)
		}

		value, params, _ := strings.Cut(h[:comma], ";")
		values = append(values, HeaderValue{
			Value:    strings.TrimSpace(value),
			Priority: parseQuality(params),
		})

		if comma == len(h) {
			break
		}
		h = h[comma+1:]
	}

	sort.Sort(byPriority(values))

	return values
}

// parseQuality returns the quality value defined by the "q" parameter in the
// given parameter list. Returns 1 if the list doesn't contain a "q" parameter, as specified
// by RFC 9110, and 0 if the quality value cannot be parsed.
func parseQuality(params string) float64 {
	for param := range strings.SplitSeq(params, ";") {
		name, value, ok := strings.Cut(param, "=")
		if !ok || !strings.EqualFold(strings.TrimSpace(name), "q") {
			continue
		}
		q := strings.TrimSpace(value)
		priority, err := strconv.ParseFloat(q, 64)
		if err != nil || priority < 0 || priority > 1 || !isQualityValue(q) {
			return 0
		}
		return priority
	}
	return 1
}

// isQualityValue returns true if the given string has the shape of a quality
// value as defined by RFC 9110 section 12.4.2: digits, at most one decimal
// point, and at most three digits after that point. strconv.ParseFloat accepts
// numbers that a quality value cannot contain, such as "1e-2", "+0.5", "0x1p-1"
// and "NaN", so the characters are checked here as well.
func isQualityValue(q string) bool {
	dot := strings.IndexByte(q, '.')
	if dot >= 0 && len(q)-dot-1 > 3 {
		return false
	}
	for i, c := range q {
		if i != dot && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}
