package validation

import (
	"fmt"
	"math"
	"strconv"
)

const (
	maxIntFloat32 = 16777216
	maxIntFloat64 = 9007199254740992
)

func numberAsFloat64(n any) (float64, bool, error) {
	switch val := n.(type) {
	case float32:
		return float64(val), true, nil
	case float64:
		return val, true, nil
	case int:
		return float64(val), true, nil
	case int8:
		return float64(val), true, nil
	case int16:
		return float64(val), true, nil
	case int32:
		return float64(val), true, nil
	case int64:
		if val > maxIntFloat64 || val < -maxIntFloat64 {
			return float64(val), false, fmt.Errorf("int64 value %d doesn't fit in float64", val)
		}
		return float64(val), true, nil
	case uint:
		if val > maxIntFloat64 {
			return float64(val), false, fmt.Errorf("uint value %d doesn't fit in float64", val)
		}
		return float64(val), true, nil
	case uint8:
		return float64(val), true, nil
	case uint16:
		return float64(val), true, nil
	case uint32:
		return float64(val), true, nil
	case uint64:
		if val > maxIntFloat64 {
			return float64(val), false, fmt.Errorf("uint64, value %d doesn't fit in float64", val)
		}
		return float64(val), true, nil
	}
	return 0, false, nil
}

type float interface {
	float32 | float64
}

type floatValidator[T float] struct{ BaseValidator }

func (v *floatValidator[T]) Validate(ctx *Context) bool {
	switch val := ctx.Value.(type) {
	case T:
		return true
	case float32:
		// float32 -> float64, no check needed
		ctx.Value = T(val)
		return true
	case float64:
		return v.checkFloatRange(ctx, val)
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

func (v *floatValidator[T]) parseString(ctx *Context, val string) bool {
	floatVal, err := strconv.ParseFloat(val, v.getBitSize())
	if err == nil {
		return v.checkFloatRange(ctx, floatVal)
	}
	return false
}

func (v *floatValidator[T]) getBitSize() int {
	var t T
	switch any(t).(type) {
	case float32:
		return 32
	default:
		return 64
	}
}

func (v *floatValidator[T]) min() float64 {
	return -v.max()
}

func (v *floatValidator[T]) max() float64 {
	var t T
	switch any(t).(type) {
	case float32:
		return math.MaxFloat32
	default:
		return math.MaxFloat64
	}
}

func (v *floatValidator[T]) checkFloatRange(ctx *Context, val float64) bool {
	ok := val >= v.min() && val <= v.max()
	if ok {
		ctx.Value = T(val)
	}
	return ok
}

func (v *floatValidator[T]) checkIntRange(ctx *Context, val int) bool {

	var t T
	ok := false
	switch any(t).(type) {
	case float32:
		ok = val <= maxIntFloat32 && val >= -maxIntFloat32
	default:
		ok = val <= maxIntFloat64 && val >= -maxIntFloat64
	}
	if ok {
		ctx.Value = T(val)
	}
	return ok
}

func (v *floatValidator[T]) checkUintRange(ctx *Context, val uint) bool {
	ok := false
	var t T
	switch any(t).(type) {
	case float32:
		ok = val <= maxIntFloat32
	default:
		ok = val <= maxIntFloat64
	}
	if ok {
		ctx.Value = T(val)
	}
	return ok
}

func (v *floatValidator[T]) Name() string {
	return fmt.Sprintf("float%d", v.getBitSize())
}

func (v *floatValidator[T]) IsType() bool { return true }

// Float64Validator validator for the "float64" rule.
type Float64Validator struct{ floatValidator[float64] }

// Float64 the field under validation must be a number
// and fit into Go's `float64` type. If the source number
// is an integer, the validator makes sure `float64` is
// capable of representing it without loss or rounding.
//
// Strings that can be converted to the target type are accepted.
// This rule converts the field to `float64` if it passes.
func Float64() *Float64Validator {
	return &Float64Validator{}
}

// Float32Validator validator for the "float32" rule.
type Float32Validator struct{ floatValidator[float32] }

// Float32 the field under validation must be a number
// and fit into Go's `float32` type. If the source number
// is an integer, the validator makes sure `float32` is
// capable of representing it without loss or rounding.
//
// Strings that can be converted to the target type are accepted.
// This rule converts the field to `float32` if it passes.
func Float32() *Float32Validator {
	return &Float32Validator{}
}
