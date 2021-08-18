package helper

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// IndexOf get the index of the given value in the given slice,
// or -1 if not found.
func IndexOf(slice interface{}, value interface{}) int {
	list := reflect.ValueOf(slice)
	length := list.Len()
	for i := 0; i < length; i++ {
		if list.Index(i).Interface() == value {
			return i
		}
	}
	return -1
}

// Contains check if a slice contains a value.
func Contains(slice interface{}, value interface{}) bool {
	return IndexOf(slice, value) != -1
}

// IndexOfStr get the index of the given value in the given string slice,
// or -1 if not found.
// Prefer using this helper instead of IndexOf for better performance.
func IndexOfStr(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

// ContainsStr check if a string slice contains a value.
// Prefer using this helper instead of Contains for better performance.
func ContainsStr(slice []string, value string) bool {
	return IndexOfStr(slice, value) != -1
}

// SliceEqual check if two generic slices are the same.
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

// ToFloat64 convert a numeric value to float64.
func ToFloat64(value interface{}) (float64, error) {
	return strconv.ParseFloat(ToString(value), 64)
}

// ToString convert a value to string.
func ToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}

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
//  "text/html,text/*;q=0.5,*/*;q=0.7"
// then
//  [{text/html 1} {*/* 0.7} {text/* 0.5}]
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

// RemoveHiddenFields if the given model is a struct pointer.
// All fields marked with the tag `model:"hide"` will be
// set to their zero value.
func RemoveHiddenFields(model interface{}) {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		if t.Kind() == reflect.Struct {
			value := reflect.ValueOf(model).Elem()
			for i := 0; i < t.NumField(); i++ {
				field := value.Field(i)
				fieldType := t.Field(i)

				if !field.CanSet() {
					continue
				}

				if field.Kind() == reflect.Struct && fieldType.Anonymous {
					// Check promoted fields recursively
					RemoveHiddenFields(field.Addr().Interface())
					continue
				}

				tag := strings.Split(fieldType.Tag.Get("model"), ";")
				if ContainsStr(tag, "hide") {
					field.Set(reflect.Zero(fieldType.Type))
				}
			}
		}
	}
}

// Only extracts the requested field from the given map[string] or structure and
// returns a map[string]interface{} containing only those values.
//
// For example:
//  type Model struct {
//    Field string
//    Num   int
//    Slice []float64
//  }
//  model := Model{
// 	  Field: "value",
// 	  Num:   42,
// 	  Slice: []float64{3, 6, 9},
//  }
//  res := Only(model, "Field", "Slice")
//
// Result:
//  map[string]interface{}{
// 	  "Field": "value",
// 	  "Slice": []float64{3, 6, 9},
//  }
//
// In case of conflicting fields (if a promoted field has the same name as a parent's
// struct field), the higher level field is kept.
func Only(data interface{}, fields ...string) map[string]interface{} {
	result := make(map[string]interface{}, len(fields))
	t := reflect.TypeOf(data)
	value := reflect.ValueOf(data)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		value = value.Elem()
	}

	if !value.IsValid() {
		return result
	}

	switch t.Kind() {
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			panic(fmt.Errorf("helper.Only only supports map[string] and structures, %s given", t.String()))
		}
		for _, k := range value.MapKeys() {
			name := k.String()
			if ContainsStr(fields, name) {
				result[name] = value.MapIndex(k).Interface()
			}
		}
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			field := value.Field(i)
			strctType := t.Field(i)
			fieldType := strctType.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			name := strctType.Name
			if fieldType.Kind() == reflect.Struct && strctType.Anonymous {
				for k, v := range Only(field.Interface(), fields...) {
					// Check if fields are conflicting
					// Highest level fields have priority
					if _, ok := result[k]; !ok {
						result[k] = v
					}
				}
			} else if ContainsStr(fields, name) {
				result[name] = value.Field(i).Interface()
			}
		}
	default:
		panic(fmt.Errorf("helper.Only only supports map[string] and structures, %s given", t.Kind()))
	}

	return result
}

// EscapeLike escape "%" and "_" characters in the given string
// for use in SQL "LIKE" clauses.
func EscapeLike(str string) string {
	escapeChars := []string{"%", "_"}
	for _, v := range escapeChars {
		str = strings.ReplaceAll(str, v, "\\"+v)
	}
	return str
}
