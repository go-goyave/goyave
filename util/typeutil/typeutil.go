package typeutil

import (
	"fmt"
	"strconv"
)

// Map is an alias to map[string]interface{}
// Useful and a cleaner way to create a JSON response object
type Map map[string]interface{}

// ToFloat64 convert a numeric value to float64.
func ToFloat64(value interface{}) (float64, error) {
	return strconv.ParseFloat(ToString(value), 64)
}

// ToString convert a value to string.
func ToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}
