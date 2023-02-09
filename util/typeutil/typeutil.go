package typeutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// Map is an alias to map[string]any
// Useful and a cleaner way to create a JSON response object
type Map map[string]any

// ToFloat64 convert a numeric value to float64.
func ToFloat64(value any) (float64, error) {
	return strconv.ParseFloat(ToString(value), 64)
}

// ToString convert a value to string.
func ToString(value any) string {
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

	// TODO it doesn't work well with null values (null.String)
	// for example: *null.String (want: if field is absent, have: if field is absent or if value is null)
	// but if using null.String, can't differentiate between absent and null

	if err := writer.Encode(data); err != nil {
		return result, err
	}
	err := decoder.Decode(&result)
	return result, err
}
