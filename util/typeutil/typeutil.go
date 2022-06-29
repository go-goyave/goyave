package typeutil

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/mapstructure"
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

// Convert anything into the desired type.
func Convert[T any](data any) (T, error) {
	var result T
	if err := mapstructure.Decode(data, &result); err != nil {
		return result, err
	}
	return result, nil
}
