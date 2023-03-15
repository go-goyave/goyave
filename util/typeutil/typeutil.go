package typeutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// Map is an alias to map[string]any
// Useful and a cleaner way to create a JSON response object
// TODO remove this
type Map map[string]any

// ToFloat64 convert a numeric value to float64.
func ToFloat64(value any) (float64, error) { // TODO remove this
	return strconv.ParseFloat(ToString(value), 64)
}

// ToString convert a value to string.
func ToString(value any) string { // TODO remove this
	return fmt.Sprintf("%v", value)
}

// Convert anything into the desired type using JSON marshaling and unmarshaling.
func Convert[T any](data any) (T, error) {
	if v, ok := data.(T); ok {
		return v, nil
	}

	var result T
	buffer := &bytes.Buffer{}
	decoder := json.NewDecoder(buffer)
	writer := json.NewEncoder(buffer)

	if err := writer.Encode(data); err != nil {
		return result, err
	}
	err := decoder.Decode(&result)
	return result, err
}

// MustConvert anything into the desired type using JSON marshaling and unmarshaling.
// Panics if it fails.
func MustConvert[T any](data any) T {
	res, err := Convert[T](data)
	if err != nil {
		panic(err)
	}
	return res
}

// Must takes any return values of a function that can also return an error.
// Panics if the error is not nil, otherwise returns the value.
//
//	typeutil.Must(time.Parse(time.RFC3339, "2023-03-15 11:07:42"))
func Must[V any](value V, err error) V {
	if err != nil {
		panic(err)
	}
	return value
}

// Ptr returns a pointer to the given value.
func Ptr[V any](value V) *V {
	return &value
}
