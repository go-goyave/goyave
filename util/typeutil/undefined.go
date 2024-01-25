package typeutil

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"

	"goyave.dev/copier"
	"goyave.dev/goyave/v5/util/errors"
)

// Undefined utility type wrapping a generic value used to differentiate
// between the absence of a field and its zero value, without using pointers.
//
// This is especially useful when using wrappers such as `sql.NullString`, which
// are structures that encode/decode to a non-struct value. When working with
// requests that may or may not contain a field that is a nullable value, you cannot
// use pointers to define the presence or absence of this kind of structure. Thus the
// case where the field is absent (zero-value) and where the field is present but has
// a null value are indistinguishable.
//
// This type only implements:
//   - `encoding.TextUnmarshaler`
//   - `json.Unmarshaler`
//   - `driver.Valuer`
//
// Because it only implements "read"-related interfaces, it is not recommended to use it
// for responses or for scanning database results. For these use-cases, it is recommended
// to use pointers for the field types with the json tag "omitempty".
type Undefined[T any] struct {
	Val     T
	Present bool
}

// NewUndefined creates a new `Undefined` wrapper with `Present` set to `true`.
func NewUndefined[T any](val T) Undefined[T] {
	return Undefined[T]{
		Val:     val,
		Present: true,
	}
}

// UnmarshalJSON implements json.Unmarshaler.
// On successful unmarshal of the underlying value, sets the `Present` field to `true`.
func (u *Undefined[T]) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &u.Val); err != nil {
		return errors.Errorf("typeutil.Undefined: couldn't unmarshal JSON: %w", err)
	}

	u.Present = true
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// If the input is a blank string, `Present` is set to `false`, otherwise `true`.
// This implementation will return an error if the underlying value doesn't implement
// `encoding.TextUnmarshaler`.
func (u *Undefined[T]) UnmarshalText(text []byte) error {
	u.Present = len(text) > 0
	if textUnmarshaler, ok := any(&u.Val).(encoding.TextUnmarshaler); ok {
		if err := textUnmarshaler.UnmarshalText(text); err != nil {
			return errors.New(err)
		}
		u.Present = true
		return nil
	}

	return errors.New("typeutil.Undefined: cannot unmarshal text: underlying value doesn't implement encoding.TextUnmarshaler")
}

// IsZero returns true for non-present values, for potential future omitempty support.
func (u Undefined[T]) IsZero() bool {
	return !u.Present
}

// IsPresent returns true for present values.
func (u Undefined[T]) IsPresent() bool {
	return u.Present
}

// Value implements the driver sql.Valuer interface.
func (u Undefined[T]) Value() (driver.Value, error) {
	if !u.Present {
		return nil, nil
	}

	if valuer, ok := any(u.Val).(driver.Valuer); ok {
		v, err := valuer.Value()
		return v, errors.New(err)
	}
	return u.Val, nil
}

// Scan implementation of `sql.Scanner` meant to be able to support copying
// from and to `Undefined` structures with `typeutil.Copy`.
func (u *Undefined[T]) Scan(src any) error {
	u.Present = true

	if scanner, ok := any(&u.Val).(sql.Scanner); ok {
		return errors.New(scanner.Scan(src))
	}

	switch val := src.(type) {
	case T:
		u.Val = val
	case *T:
		u.Val = *val
	case nil:
		// Set to zero-value
		var t T
		u.Val = t
	default:
		var t T
		return errors.Errorf("typeutil.Undefined: Scan() incompatible types (src: %T, dst: %T)", src, t)
	}
	return nil
}

// CopyValue implements the copier.Valuer interface.
func (u Undefined[T]) CopyValue() any {
	if !u.Present {
		return nil
	}

	if valuer, ok := any(u.Val).(copier.Valuer); ok {
		return valuer.CopyValue()
	}
	return u.Val
}

// Default return the value if present, otherwise returns the given default value.
func (u Undefined[T]) Default(defaultValue T) T {
	if u.Present {
		return u.Val
	}
	return defaultValue
}
