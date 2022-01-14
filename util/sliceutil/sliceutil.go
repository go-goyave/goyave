package sliceutil

import "reflect"

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
// Prefer using this function instead of IndexOf for better performance.
func IndexOfStr(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

// ContainsStr check if a string slice contains a value.
// Prefer using this function instead of Contains for better performance.
func ContainsStr(slice []string, value string) bool {
	return IndexOfStr(slice, value) != -1
}

// Equal check if two generic slices are the same.
func Equal(first interface{}, second interface{}) bool {
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
