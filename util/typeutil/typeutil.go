package typeutil

import (
	"bytes"
	"encoding/json"

	"goyave.dev/goyave/v5/util/errors"
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
		return result, errors.NewSkip(err, 3)
	}
	err := decoder.Decode(&result)
	if err != nil {
		err = errors.NewSkip(err, 3)
	}
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
