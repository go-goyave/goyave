package validation

import (
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestFloat32Validator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Float32()
		assert.NotNil(t, v)
		assert.Equal(t, "float32", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: 2, want: true, wantValue: float32(2.0)},
		{value: 2.5, want: true, wantValue: float32(2.5)},
		{value: float32(2.5), want: true, wantValue: float32(2.5)},
		{value: uint(2), want: true, wantValue: float32(2.0)},
		{value: 'a', want: true, wantValue: float32(97.0)},
		{value: "2.5", want: true, wantValue: float32(2.5)},
		{value: strconv.FormatFloat(math.MaxFloat32, 'f', 24, 64), want: true, wantValue: float32(math.MaxFloat32)},
		{value: strconv.FormatFloat(-math.MaxFloat32, 'f', 24, 64), want: true, wantValue: float32(-math.MaxFloat32)},
		{value: strconv.FormatFloat(math.MaxFloat64, 'f', 24, 64), want: false},
		{value: strconv.FormatFloat(-math.MaxFloat64, 'f', 24, 64), want: false},
		{value: uint8(math.MaxUint8), want: true, wantValue: float32(math.MaxUint8)},
		{value: uint16(math.MaxUint16), want: true, wantValue: float32(math.MaxUint16)},
		{value: uint32(maxIntFloat32), want: true, wantValue: float32(maxIntFloat32)},
		{value: uint32(math.MaxUint32), want: false},
		{value: uint64(math.MaxUint64), want: false},
		{value: uint(math.MaxUint), want: false},
		{value: int8(math.MaxInt8), want: true, wantValue: float32(math.MaxInt8)},
		{value: int16(math.MaxInt16), want: true, wantValue: float32(math.MaxInt16)},
		{value: int32(maxIntFloat32), want: true, wantValue: float32(maxIntFloat32)},
		{value: int32(math.MaxInt32), want: false},
		{value: int64(math.MaxInt64), want: false},
		{value: int(math.MaxInt), want: false},
		{value: int8(math.MinInt8), want: true, wantValue: float32(math.MinInt8)},
		{value: int16(math.MinInt16), want: true, wantValue: float32(math.MinInt16)},
		{value: int32(-maxIntFloat32), want: true, wantValue: float32(-maxIntFloat32)},
		{value: int32(math.MinInt32), want: false},
		{value: int64(math.MinInt64), want: false},
		{value: int(math.MinInt), want: false},
		{value: math.MaxFloat64, want: false},
		{value: math.MaxFloat32, want: true, wantValue: float32(math.MaxFloat32)},
		{value: -math.MaxFloat32, want: true, wantValue: float32(-math.MaxFloat32)},
		{value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:06:42Z")), want: false},
		{value: "string", want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Float32()
			ctx := &Context{
				Value: c.value,
			}
			ok := v.Validate(ctx)
			assert.Equal(t, c.want, ok)
			if c.wantValue != nil {
				assert.Equal(t, c.wantValue, ctx.Value)
			}
		})
	}
}

func TestFloat64Validator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Float64()
		assert.NotNil(t, v)
		assert.Equal(t, "float64", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: 2, want: true, wantValue: float64(2.0)},
		{value: 2.5, want: true, wantValue: float64(2.5)},
		{value: float64(2.5), want: true, wantValue: float64(2.5)},
		{value: float32(math.MaxFloat32), want: true, wantValue: float64(math.MaxFloat32)},
		{value: uint(2), want: true, wantValue: float64(2.0)},
		{value: 'a', want: true, wantValue: float64(97.0)},
		{value: "2.5", want: true, wantValue: float64(2.5)},
		{value: strconv.FormatFloat(math.MaxFloat64, 'f', 24, 64), want: true, wantValue: float64(math.MaxFloat64)},
		{value: strconv.FormatFloat(-math.MaxFloat64, 'f', 24, 64), want: true, wantValue: float64(-math.MaxFloat64)},
		{value: uint8(math.MaxUint8), want: true, wantValue: float64(math.MaxUint8)},
		{value: uint16(math.MaxUint16), want: true, wantValue: float64(math.MaxUint16)},
		{value: uint32(math.MaxUint32), want: true, wantValue: float64(math.MaxUint32)},
		{value: uint64(math.MaxUint64), want: false},
		{value: uint(math.MaxUint), want: false},
		{value: int8(math.MaxInt8), want: true, wantValue: float64(math.MaxInt8)},
		{value: int16(math.MaxInt16), want: true, wantValue: float64(math.MaxInt16)},
		{value: int32(math.MaxInt32), want: true, wantValue: float64(math.MaxInt32)},
		{value: int64(math.MaxInt64), want: false},
		{value: int64(maxIntFloat64), want: true, wantValue: float64(maxIntFloat64)},
		{value: int(math.MaxInt), want: false},
		{value: int8(math.MinInt8), want: true, wantValue: float64(math.MinInt8)},
		{value: int16(math.MinInt16), want: true, wantValue: float64(math.MinInt16)},
		{value: int32(math.MinInt32), want: true, wantValue: float64(math.MinInt32)},
		{value: int64(math.MinInt64), want: false},
		{value: int64(-maxIntFloat64), want: true, wantValue: float64(-maxIntFloat64)},
		{value: int(math.MinInt), want: false},
		{value: math.MaxFloat64, want: true, wantValue: math.MaxFloat64},
		{value: -math.MaxFloat64, want: true, wantValue: -math.MaxFloat64},
		{value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:06:42Z")), want: false},
		{value: "string", want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Float64()
			ctx := &Context{
				Value: c.value,
			}
			ok := v.Validate(ctx)
			assert.Equal(t, c.want, ok)
			if c.wantValue != nil {
				assert.Equal(t, c.wantValue, ctx.Value)
			}
		})
	}
}
