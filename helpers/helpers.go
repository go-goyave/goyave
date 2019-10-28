package helpers

import "reflect"

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
