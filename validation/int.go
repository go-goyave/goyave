package validation

import (
	"fmt"
	"math"
	"strconv"
)

type integer interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64
}

type intValidator[T integer] struct{ BaseValidator }

func (v *intValidator[T]) Validate(ctx *Context) bool {
	switch val := ctx.Value.(type) {
	case T:
		return true
	case float32:
		return v.checkFloat32Range(ctx, val)
	case float64:
		return v.checkFloat64Range(ctx, val)
	case string:
		return v.parseString(ctx, val)
	case int:
		return v.checkIntRange(ctx, val)
	case int8:
		return v.checkIntRange(ctx, int(val))
	case int16:
		return v.checkIntRange(ctx, int(val))
	case int32:
		return v.checkIntRange(ctx, int(val))
	case int64:
		return v.checkIntRange(ctx, int(val))
	case uint:
		return v.checkUintRange(ctx, val)
	case uint8:
		return v.checkUintRange(ctx, uint(val))
	case uint16:
		return v.checkUintRange(ctx, uint(val))
	case uint32:
		return v.checkUintRange(ctx, uint(val))
	case uint64:
		return v.checkUintRange(ctx, uint(val))
	}

	return false
}

func (v *intValidator[T]) isUnsigned() bool {
	var t T
	switch any(t).(type) {
	case uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func (v *intValidator[T]) getBitSize() int {
	var t T
	switch any(t).(type) {
	case int8, uint8:
		return 8
	case int16, uint16:
		return 16
	case int32, uint32:
		return 32
	case int64, uint64:
		return 64
	}
	return strconv.IntSize
}

func (v *intValidator[T]) checkFloat64Range(ctx *Context, val float64) bool {
	if val > maxIntFloat64 || val < -maxIntFloat64 {
		return false
	}
	ok := !((v.isUnsigned() && val < 0) || math.Abs(val-float64(T(val))) > 0 || (val < 0 && int(val) < v.min()) || (val > 0 && uint(val) > v.max()))
	if ok {
		ctx.Value = T(val)
	}
	return ok
}

func (v *intValidator[T]) checkFloat32Range(ctx *Context, val float32) bool {
	if val > maxIntFloat32 || val < -maxIntFloat32 {
		return false
	}
	ok := !((v.isUnsigned() && val < 0) || math.Abs(float64(val-float32(T(val)))) > 0 || (val < 0 && int(val) < v.min()) || (val > 0 && uint(val) > v.max()))
	if ok {
		ctx.Value = T(val)
	}
	return ok
}

func (v *intValidator[T]) checkIntRange(ctx *Context, val int) bool {
	ok := false
	if v.isUnsigned() {
		ok = val >= v.min() && uint(val) <= v.max()
	} else {
		ok = val >= v.min() && val <= int(v.max())
	}
	if ok {
		ctx.Value = T(val)
	}
	return ok
}

func (v *intValidator[T]) checkUintRange(ctx *Context, val uint) bool {
	ok := val <= v.max()
	if ok {
		ctx.Value = T(val)
	}
	return ok
}

func (v *intValidator[T]) min() int {
	if v.isUnsigned() {
		return 0
	}
	bitSize := v.getBitSize() - 1
	return -1 << bitSize
}

func (v *intValidator[T]) max() uint {
	bitSize := v.getBitSize()
	if !v.isUnsigned() {
		bitSize--
	}
	return 1<<bitSize - 1
}

func (v *intValidator[T]) parseString(ctx *Context, val string) bool {
	if v.isUnsigned() {
		intVal, err := strconv.ParseUint(val, 10, v.getBitSize())
		if err == nil {
			ctx.Value = T(intVal)
		}
		return err == nil
	}

	intVal, err := strconv.ParseInt(val, 10, v.getBitSize())
	if err == nil {
		ctx.Value = T(intVal)
	}
	return err == nil
}

func (v *intValidator[T]) Name() string {
	var t T
	switch any(t).(type) {
	case int:
		return "int"
	case uint:
		return "uint"
	}
	format := "int%d"
	if v.isUnsigned() {
		format = "uint%d"
	}
	return fmt.Sprintf(format, v.getBitSize())
}

// IsType returns true
func (v *intValidator[T]) IsType() bool { return true }

// IntValidator validator for the "int" rule.
type IntValidator struct{ intValidator[int] }

// Int the field under validation must be an integer
// and fit into Go's `int` type. If the source number is
// a float, the validator makes sure the value is within
// the range of integers that the float can accurately represent.
//
// Floats are only accepted if they don't have a decimal.
// Strings that can be converted to the target type are accepted.
// This rule converts the field to `int` if it passes.
func Int() *IntValidator {
	return &IntValidator{}
}

// Int8Validator validator for the "int8" rule.
type Int8Validator struct{ intValidator[int8] }

// Int8 the field under validation must be an integer
// and fit into Go's `int8` type. If the source number is
// a float, the validator makes sure the value is within
// the range of integers that the float can accurately represent.
//
// Floats are only accepted if they don't have a decimal.
// Strings that can be converted to the target type are accepted.
// This rule converts the field to `int8` if it passes.
func Int8() *Int8Validator {
	return &Int8Validator{}
}

// Int16Validator validator for the "int16" rule.
type Int16Validator struct{ intValidator[int16] }

// Int16 the field under validation must be an integer
// and fit into Go's `int16` type. If the source number is
// a float, the validator makes sure the value is within
// the range of integers that the float can accurately represent.
//
// Floats are only accepted if they don't have a decimal.
// Strings that can be converted to the target type are accepted.
// This rule converts the field to `int16` if it passes.
func Int16() *Int16Validator {
	return &Int16Validator{}
}

// Int32Validator validator for the "int32" rule.
type Int32Validator struct{ intValidator[int32] }

// Int32 the field under validation must be an integer
// and fit into Go's `int32` type. If the source number is
// a float, the validator makes sure the value is within
// the range of integers that the float can accurately represent.
//
// Floats are only accepted if they don't have a decimal.
// Strings that can be converted to the target type are accepted.
// This rule converts the field to `int32` if it passes.
func Int32() *Int32Validator {
	return &Int32Validator{}
}

// Int64Validator validator for the "int64" rule.
type Int64Validator struct{ intValidator[int64] }

// Int64 the field under validation must be an integer
// and fit into Go's `int64` type. If the source number is
// a float, the validator makes sure the value is within
// the range of integers that the float can accurately represent.
//
// Floats are only accepted if they don't have a decimal.
// Strings that can be converted to the target type are accepted.
// This rule converts the field to `int64` if it passes.
func Int64() *Int64Validator {
	return &Int64Validator{}
}

// UintValidator validator for the "uint" rule.
type UintValidator struct{ intValidator[uint] }

// Uint the field under validation must be a positive integer
// and fit into Go's `uint` type. If the source number is
// a float, the validator makes sure the value is within
// the range of integers that the float can accurately represent.
//
// Floats are only accepted if they don't have a decimal.
// Strings that can be converted to the target type are accepted.
// This rule converts the field to `uint` if it passes.
func Uint() *UintValidator {
	return &UintValidator{}
}

// Uint8Validator validator for the "uint8" rule.
type Uint8Validator struct{ intValidator[uint8] }

// Uint8 the field under validation must be a positive integer
// and fit into Go's `uint8` type. If the source number is
// a float, the validator makes sure the value is within
// the range of integers that the float can accurately represent.
//
// Floats are only accepted if they don't have a decimal.
// Strings that can be converted to the target type are accepted.
// This rule converts the field to `uint8` if it passes.
func Uint8() *Uint8Validator {
	return &Uint8Validator{}
}

// Uint16Validator validator for the "uint16" rule.
type Uint16Validator struct{ intValidator[uint16] }

// Uint16 the field under validation must be a positive integer
// and fit into Go's `uint16` type. If the source number is
// a float, the validator makes sure the value is within
// the range of integers that the float can accurately represent.
//
// Floats are only accepted if they don't have a decimal.
// Strings that can be converted to the target type are accepted.
// This rule converts the field to `uint16` if it passes.
func Uint16() *Uint16Validator {
	return &Uint16Validator{}
}

// Uint32Validator validator for the "uint32" rule.
type Uint32Validator struct{ intValidator[uint32] }

// Uint32 the field under validation must be a positive integer
// and fit into Go's `uint32` type. If the source number is
// a float, the validator makes sure the value is within
// the range of integers that the float can accurately represent.
//
// Floats are only accepted if they don't have a decimal.
// Strings that can be converted to the target type are accepted.
// This rule converts the field to `uint32` if it passes.
func Uint32() *Uint32Validator {
	return &Uint32Validator{}
}

// Uint64Validator validator for the "uint64" rule.
type Uint64Validator struct{ intValidator[uint64] }

// Uint64 the field under validation must be a positive integer
// and fit into Go's `uint64` type. If the source number is
// a float, the validator makes sure the value is within
// the range of integers that the float can accurately represent.
//
// Floats are only accepted if they don't have a decimal.
// Strings that can be converted to the target type are accepted.
// This rule converts the field to `uint64` if it passes.
func Uint64() *Uint64Validator {
	return &Uint64Validator{}
}
