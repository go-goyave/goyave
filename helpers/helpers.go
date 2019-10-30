package helpers

import (
	"fmt"
	"reflect"
	"strconv"
)

// Contains check if a slice contains a value
func Contains(slice interface{}, value interface{}) bool {
	list := reflect.ValueOf(slice)
	for i := 0; i < list.Len(); i++ {
		if list.Index(i).Interface() == value {
			return true
		}
	}
	return false
}

// ToFloat64 convert a numeric value to float64
func ToFloat64(value interface{}) (float64, error) {
	return strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
}
