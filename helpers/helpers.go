package helpers

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Contains check if a slice contains a value
func Contains(slice interface{}, value interface{}) bool {
	list := reflect.ValueOf(slice)
	length := list.Len()
	for i := 0; i < length; i++ {
		if list.Index(i).Interface() == value {
			return true
		}
	}
	return false
}

// SliceEqual check if a slice is the same as another one
func SliceEqual(first interface{}, second interface{}) bool {
	l1 := reflect.ValueOf(first)
	l2 := reflect.ValueOf(second)
	length := l1.Len()
	if length != l2.Len() {
		return false
	}

	for i := 0; i < length; i++ {
		if l1.Index(i).Interface() != l2.Index(i).Interface() {
			return false
		}
	}
	return true
}

// ToFloat64 convert a numeric value to float64
func ToFloat64(value interface{}) (float64, error) {
	return strconv.ParseFloat(ToString(value), 64)
}

// ToString convert a value to string
func ToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

// ParseMultiValuesHeader parses multi-values HTTP headers, taking the
// quality values into account. The result is a slice of values sorted
// according to the order of priority.
//
// See: https://developer.mozilla.org/en-US/docs/Glossary/Quality_values
//
// For the following header:
//  "text/html;q=0.8,text/*;q=0.8,*/*;q=0.8"
// then
//  ["text/html" "text/*" "*/*"]
func ParseMultiValuesHeader(header string) []HeaderValue {
	values := []HeaderValue{}
	regex := regexp.MustCompile("^q=([01]\\.[0-9]{1,3})$")

	for _, v := range strings.Split(header, ",") {
		val := HeaderValue{}
		if i := strings.Index(v, ";"); i != -1 {
			// Parse priority
			q := v[i+1:]

			sub := regex.FindStringSubmatch(q)
			priority := 0.0
			if len(sub) > 1 {
				if p, err := strconv.ParseFloat(sub[1], 64); err == nil {
					priority = p
				}
			}
			// Priority set to 0 if the quality value cannot be parsed
			val.Priority = priority

			val.Value = v[:i]
		} else {
			val.Value = v
			val.Priority = 1
		}

		values = append(values, val)
	}

	sort.Sort(byPriority(values))

	return values
}

// HeaderValue represent a value and its quality value (priority)
// in a multi-values HTTP header.
type HeaderValue struct {
	Value    string
	Priority float64
}
