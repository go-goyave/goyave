package typeutil

import (
	"encoding/json/v2"

	"goyave.dev/copier"
	"goyave.dev/goyave/v5/util/errors"
)

// Convert anything into the desired type using JSON marshaling and unmarshaling.
// TODO interface so the marshaler/unmarshaler can be swapped?
func Convert[T any](data any, opts ...json.Options) (T, error) {
	if v, ok := data.(T); ok {
		return v, nil
	}

	var result T
	buf, err := json.Marshal(data, opts...)
	if err != nil {
		return result, errors.NewSkip(err, 3)
	}
	err = json.Unmarshal(buf, &result, opts...)
	return result, errors.NewSkip(err, 3)
}

// MustConvert anything into the desired type using JSON marshaling and unmarshaling.
// Panics if it fails.
func MustConvert[T any](data any, opts ...json.Options) T {
	res, err := Convert[T](data, opts...)
	if err != nil {
		panic(err)
	}
	return res
}

// Copy deep-copy a DTO's non-zero fields to the given model. The model is updated in-place and returned.
// Field names are matched in a case sensitive way.
// If you need to copy a zero-value (empty string, `false`, 0, etc) into the destination model, your DTO
// can take advantage of `typeutil.Undefined`.
// Panics if an error occurs.
func Copy[T, D any](model *T, dto D) *T {
	err := copier.CopyWithOption(model, dto, copier.Option{IgnoreEmpty: true, DeepCopy: true, CaseSensitive: true})
	if err != nil {
		panic(errors.NewSkip(err, 3))
	}
	return model
}
