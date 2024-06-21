package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5/util/errors"
)

// Entry is the internal reprensentation of a config entry.
// It contains the entry value, its expected type (for validation)
// and a slice of authorized values (for validation too). If this slice
// is empty, it means any value can be used, provided it is of the correct type.
type Entry struct {
	Value            any
	AuthorizedValues []any // Leave empty for "any"
	Type             reflect.Kind
	IsSlice          bool
	Required         bool
}

func makeEntryFromValue(value any) *Entry {
	isSlice := false
	t := reflect.TypeOf(value)
	kind := t.Kind()
	if kind == reflect.Slice {
		kind = t.Elem().Kind()
		isSlice = true
	}
	return &Entry{value, []any{}, kind, isSlice, false}
}

func (e *Entry) validate(key string) error {
	if err := e.tryEnvVarConversion(key); err != nil {
		return err
	}

	v := reflect.ValueOf(e.Value)
	if e.Required && (!v.IsValid() || e.Value == nil || (v.Kind() == reflect.Pointer && v.IsNil())) {
		return errors.Errorf("%q is required", key)
	}

	t := reflect.TypeOf(e.Value)
	if t == nil {
		return nil // Can't determine type, is 'zero' value.
	}
	kind := t.Kind()
	if e.IsSlice && kind == reflect.Slice {
		kind = t.Elem().Kind()
	}
	if kind != e.Type && !e.tryConversion(kind) {
		var message string
		if e.IsSlice {
			message = "%q must be a slice of %s"
		} else {
			message = "%q type must be %s"
		}

		return errors.Errorf(message, key, e.Type)
	}

	if len(e.AuthorizedValues) > 0 {
		if e.IsSlice {
			// Accepted values for slices define the values that can be used inside the slice
			// It doesn't represent the value of the slice itself (content and order)
			length := v.Len()
			for i := 0; i < length; i++ {
				if !lo.Contains(e.AuthorizedValues, v.Index(i).Interface()) {
					return errors.Errorf("%q elements must have one of the following values: %v", key, e.AuthorizedValues)
				}
			}
		} else if !lo.Contains(e.AuthorizedValues, e.Value) {
			return errors.Errorf("%q must have one of the following values: %v", key, e.AuthorizedValues)
		}
	}

	return nil
}

func (e *Entry) tryConversion(kind reflect.Kind) bool {
	if !e.IsSlice && kind == reflect.Float64 && e.Type == reflect.Int {
		intVal, ok := convertInt(e.Value.(float64))
		if ok {
			e.Value = intVal
			return true
		}
	} else if e.IsSlice && kind == reflect.Interface {
		original := e.Value.([]any)
		var newValue any
		var ok bool
		switch e.Type {
		case reflect.Int:
			newValue, ok = convertIntSlice(original)
		case reflect.Float64:
			newValue, ok = convertSlice[float64](original)
		case reflect.String:
			newValue, ok = convertSlice[string](original)
		case reflect.Bool:
			newValue, ok = convertSlice[bool](original)
		}
		if ok {
			e.Value = newValue
			return true
		}
	}

	return false
}

func convertSlice[T any](slice []any) ([]T, bool) {
	result := make([]T, len(slice))
	for k, v := range slice {
		value, ok := v.(T)
		if !ok {
			return nil, false
		}
		result[k] = value
	}
	return result, true
}

func convertInt(value any) (int, bool) {
	switch val := value.(type) {
	case int:
		return val, true
	case float64:
		intVal := int(val)
		if val == float64(intVal) {
			return intVal, true
		}
	}
	return 0, false
}

func convertIntSlice(original []any) ([]int, bool) {
	slice := make([]int, len(original))
	for k, v := range original {
		intVal, ok := convertInt(v)
		if !ok {
			return nil, false
		}
		slice[k] = intVal
	}
	return slice, true
}

func (e *Entry) tryEnvVarConversion(key string) error {
	str, ok := e.Value.(string)
	if ok {
		val, err := e.convertEnvVar(str, key)
		if err == nil && val != nil {

			if e.IsSlice {
				return errors.Errorf("%q is a slice entry, it cannot be loaded from env", key)
			}

			e.Value = val
		}
		return err
	}

	return nil
}

func (e *Entry) convertEnvVar(str, key string) (any, error) {
	if strings.HasPrefix(str, "${") && strings.HasSuffix(str, "}") {
		varName := str[2 : len(str)-1]
		value, set := os.LookupEnv(varName)
		if !set {
			return nil, errors.Errorf("%q: %q environment variable is not set", key, varName)
		}

		switch e.Type {
		case reflect.Int:
			if i, err := strconv.Atoi(value); err == nil {
				return i, nil
			}
			return nil, errors.Errorf("%q could not be converted to int from environment variable %q of value %q", key, varName, value)
		case reflect.Float64:
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				return f, nil
			}
			return nil, errors.Errorf("%q could not be converted to float64 from environment variable %q of value %q", key, varName, value)
		case reflect.Bool:
			if b, err := strconv.ParseBool(value); err == nil {
				return b, nil
			}
			return nil, errors.Errorf("%q could not be converted to bool from environment variable %q of value %q", key, varName, value)
		default:
			// Keep value as string if type is not supported and let validation do its job
			return value, nil
		}
	}

	return nil, nil
}
