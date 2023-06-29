package typeutil

import (
	"bytes"
	"encoding/json"
)

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
