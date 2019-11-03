package helpers

import (
	"fmt"
	"reflect"
	"strconv"
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
