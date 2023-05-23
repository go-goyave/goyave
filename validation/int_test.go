package validation

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Int()
		assert.NotNil(t, v)
		assert.Equal(t, "int", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: 2, want: true, wantValue: int(2)},
		{value: int(math.MaxInt), want: true, wantValue: int(math.MaxInt)},
		{value: int8(math.MaxInt8), want: true, wantValue: int(math.MaxInt8)},
		{value: int16(math.MaxInt16), want: true, wantValue: int(math.MaxInt16)},
		{value: int32(math.MaxInt32), want: true, wantValue: int(math.MaxInt32)},
		{value: int64(math.MaxInt64), want: true, wantValue: int(math.MaxInt64)},
		{value: int(math.MinInt), want: true, wantValue: int(math.MinInt)},
		{value: int8(math.MinInt8), want: true, wantValue: int(math.MinInt8)},
		{value: int16(math.MinInt16), want: true, wantValue: int(math.MinInt16)},
		{value: int32(math.MinInt32), want: true, wantValue: int(math.MinInt32)},
		{value: int64(math.MinInt64), want: true, wantValue: int(math.MinInt64)},
		{value: uint8(math.MaxUint8), want: true, wantValue: int(math.MaxUint8)},
		{value: uint16(math.MaxUint16), want: true, wantValue: int(math.MaxUint16)},
		{value: uint32(math.MaxUint32), want: true, wantValue: int(math.MaxUint32)},
		{value: uint64(math.MaxUint64), want: false},
		{value: uint(math.MaxUint), want: false},
		{value: uint(123), want: true, wantValue: int(123)},
		{value: uint64(123), want: true, wantValue: int(123)},
		{value: float64(math.MaxInt64), want: false},
		{value: float64(-math.MaxInt64), want: false},
		{value: float64(maxIntFloat64), want: true, wantValue: int(maxIntFloat64)},
		{value: float64(-maxIntFloat64), want: true, wantValue: int(-maxIntFloat64)},
		{value: float32(maxIntFloat32), want: true, wantValue: int(maxIntFloat32)},
		{value: float32(-maxIntFloat32), want: true, wantValue: int(-maxIntFloat32)},
		{value: fmt.Sprintf("%d", math.MaxInt), want: true, wantValue: int(math.MaxInt)},
		{value: fmt.Sprintf("%d", math.MinInt), want: true, wantValue: int(math.MinInt)},
		{value: 2.0, want: true, wantValue: int(2)},
		{value: float32(2.0), want: true, wantValue: int(2)},
		{value: "2", want: true, wantValue: int(2)},
		{value: 2.5, want: false},
		{value: float32(2.5), want: false},
		{value: 'a', want: true, wantValue: int(97)},
		{value: "string", want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Int()
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

func TestInt8Validator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Int8()
		assert.NotNil(t, v)
		assert.Equal(t, "int8", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: 2, want: true, wantValue: int8(2)},
		{value: int(math.MaxInt), want: false},
		{value: int8(math.MaxInt8), want: true, wantValue: int8(math.MaxInt8)},
		{value: int16(math.MaxInt16), want: false},
		{value: int16(math.MaxInt8), want: true, wantValue: int8(math.MaxInt8)},
		{value: int32(math.MaxInt32), want: false},
		{value: int32(math.MaxInt8), want: true, wantValue: int8(math.MaxInt8)},
		{value: int64(math.MaxInt64), want: false},
		{value: int64(math.MaxInt8), want: true, wantValue: int8(math.MaxInt8)},
		{value: int(math.MinInt), want: false},
		{value: int8(math.MinInt8), want: true, wantValue: int8(math.MinInt8)},
		{value: int16(math.MinInt16), want: false},
		{value: int32(math.MinInt32), want: false},
		{value: int64(math.MinInt64), want: false},
		{value: uint8(math.MaxUint8), want: false},
		{value: uint8(math.MaxInt8), want: true, wantValue: int8(math.MaxInt8)},
		{value: uint16(math.MaxUint16), want: false},
		{value: uint16(math.MaxInt8), want: true, wantValue: int8(math.MaxInt8)},
		{value: uint32(math.MaxUint32), want: false},
		{value: uint32(math.MaxInt8), want: true, wantValue: int8(math.MaxInt8)},
		{value: uint64(math.MaxUint64), want: false},
		{value: uint64(math.MaxInt8), want: true, wantValue: int8(math.MaxInt8)},
		{value: uint(math.MaxUint), want: false},
		{value: uint(123), want: true, wantValue: int8(123)},
		{value: uint64(123), want: true, wantValue: int8(123)},
		{value: float64(math.MaxInt8), want: true, wantValue: int8(math.MaxInt8)},
		{value: float64(math.MinInt8), want: true, wantValue: int8(math.MinInt8)},
		{value: float32(math.MaxInt8), want: true, wantValue: int8(math.MaxInt8)},
		{value: float32(math.MinInt8), want: true, wantValue: int8(math.MinInt8)},
		{value: 2.0, want: true, wantValue: int8(2)},
		{value: float32(2.0), want: true, wantValue: int8(2)},
		{value: "2", want: true, wantValue: int8(2)},
		{value: "-2", want: true, wantValue: int8(-2)},
		{value: 2.5, want: false},
		{value: float32(2.5), want: false},
		{value: 'a', want: true, wantValue: int8(97)},
		{value: "string", want: false},
		{value: fmt.Sprintf("%d", math.MaxInt), want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Int8()
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

func TestInt16Validator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Int16()
		assert.NotNil(t, v)
		assert.Equal(t, "int16", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: 2, want: true, wantValue: int16(2)},
		{value: int(math.MaxInt), want: false},
		{value: int8(math.MaxInt8), want: true, wantValue: int16(math.MaxInt8)},
		{value: int16(math.MaxInt16), want: true, wantValue: int16(math.MaxInt16)},
		{value: int32(math.MaxInt32), want: false},
		{value: int32(math.MaxInt16), want: true, wantValue: int16(math.MaxInt16)},
		{value: int64(math.MaxInt64), want: false},
		{value: int64(math.MaxInt16), want: true, wantValue: int16(math.MaxInt16)},
		{value: int(math.MinInt), want: false},
		{value: int8(math.MinInt8), want: true, wantValue: int16(math.MinInt8)},
		{value: int16(math.MinInt16), want: true, wantValue: int16(math.MinInt16)},
		{value: int32(math.MinInt32), want: false},
		{value: int32(math.MinInt16), want: true, wantValue: int16(math.MinInt16)},
		{value: int64(math.MinInt64), want: false},
		{value: int64(math.MinInt16), want: true, wantValue: int16(math.MinInt16)},
		{value: uint8(math.MaxUint8), want: true, wantValue: int16(math.MaxUint8)},
		{value: uint16(math.MaxUint16), want: false},
		{value: uint16(math.MaxUint8), want: true, wantValue: int16(math.MaxUint8)},
		{value: uint32(math.MaxUint32), want: false},
		{value: uint32(math.MaxUint8), want: true, wantValue: int16(math.MaxUint8)},
		{value: uint64(math.MaxUint64), want: false},
		{value: uint64(math.MaxUint8), want: true, wantValue: int16(math.MaxUint8)},
		{value: uint(math.MaxUint), want: false},
		{value: uint(123), want: true, wantValue: int16(123)},
		{value: uint64(123), want: true, wantValue: int16(123)},
		{value: float64(math.MaxInt16), want: true, wantValue: int16(math.MaxInt16)},
		{value: float64(math.MinInt16), want: true, wantValue: int16(math.MinInt16)},
		{value: float32(math.MaxInt16), want: true, wantValue: int16(math.MaxInt16)},
		{value: float32(math.MinInt16), want: true, wantValue: int16(math.MinInt16)},
		{value: 2.0, want: true, wantValue: int16(2)},
		{value: float32(2.0), want: true, wantValue: int16(2)},
		{value: "2", want: true, wantValue: int16(2)},
		{value: "-2", want: true, wantValue: int16(-2)},
		{value: 2.5, want: false},
		{value: float32(2.5), want: false},
		{value: 'a', want: true, wantValue: int16(97)},
		{value: "string", want: false},
		{value: fmt.Sprintf("%d", math.MaxInt), want: false},
		{value: fmt.Sprintf("%d", math.MinInt), want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Int16()
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

func TestInt32Validator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Int32()
		assert.NotNil(t, v)
		assert.Equal(t, "int32", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: 2, want: true, wantValue: int32(2)},
		{value: int(math.MaxInt), want: false},
		{value: int8(math.MaxInt8), want: true, wantValue: int32(math.MaxInt8)},
		{value: int16(math.MaxInt16), want: true, wantValue: int32(math.MaxInt16)},
		{value: int32(math.MaxInt32), want: true, wantValue: int32(math.MaxInt32)},
		{value: int64(math.MaxInt64), want: false},
		{value: int64(math.MaxInt32), want: true, wantValue: int32(math.MaxInt32)},
		{value: int(math.MinInt), want: false},
		{value: int8(math.MinInt8), want: true, wantValue: int32(math.MinInt8)},
		{value: int16(math.MinInt16), want: true, wantValue: int32(math.MinInt16)},
		{value: int32(math.MinInt32), want: true, wantValue: int32(math.MinInt32)},
		{value: int64(math.MinInt64), want: false},
		{value: int64(math.MinInt32), want: true, wantValue: int32(math.MinInt32)},
		{value: uint8(math.MaxUint8), want: true, wantValue: int32(math.MaxUint8)},
		{value: uint16(math.MaxUint16), want: true, wantValue: int32(math.MaxUint16)},
		{value: uint32(math.MaxUint32), want: false},
		{value: uint64(math.MaxUint64), want: false},
		{value: uint32(math.MaxUint16), want: true, wantValue: int32(math.MaxUint16)},
		{value: uint64(math.MaxUint16), want: true, wantValue: int32(math.MaxUint16)},
		{value: uint(math.MaxUint), want: false},
		{value: uint(123), want: true, wantValue: int32(123)},
		{value: uint64(123), want: true, wantValue: int32(123)},
		{value: float64(math.MaxInt32), want: true, wantValue: int32(math.MaxInt32)},
		{value: float64(math.MinInt32), want: true, wantValue: int32(math.MinInt32)},
		{value: float32(maxIntFloat32), want: true, wantValue: int32(maxIntFloat32)},
		{value: float32(-maxIntFloat32), want: true, wantValue: int32(-maxIntFloat32)},
		{value: 2.0, want: true, wantValue: int32(2)},
		{value: float32(2.0), want: true, wantValue: int32(2)},
		{value: "2", want: true, wantValue: int32(2)},
		{value: "-2", want: true, wantValue: int32(-2)},
		{value: fmt.Sprintf("%d", math.MaxInt), want: false},
		{value: fmt.Sprintf("%d", math.MinInt), want: false},
		{value: 2.5, want: false},
		{value: float32(2.5), want: false},
		{value: 'a', want: true, wantValue: int32(97)},
		{value: "string", want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Int32()
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

func TestInt64Validator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Int64()
		assert.NotNil(t, v)
		assert.Equal(t, "int64", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: 2, want: true, wantValue: int64(2)},
		{value: int(math.MaxInt), want: true, wantValue: int64(math.MaxInt)},
		{value: int8(math.MaxInt8), want: true, wantValue: int64(math.MaxInt8)},
		{value: int16(math.MaxInt16), want: true, wantValue: int64(math.MaxInt16)},
		{value: int32(math.MaxInt32), want: true, wantValue: int64(math.MaxInt32)},
		{value: int64(math.MaxInt64), want: true, wantValue: int64(math.MaxInt64)},
		{value: int(math.MinInt), want: true, wantValue: int64(math.MinInt)},
		{value: int8(math.MinInt8), want: true, wantValue: int64(math.MinInt8)},
		{value: int16(math.MinInt16), want: true, wantValue: int64(math.MinInt16)},
		{value: int32(math.MinInt32), want: true, wantValue: int64(math.MinInt32)},
		{value: int64(math.MinInt64), want: true, wantValue: int64(math.MinInt64)},
		{value: uint8(math.MaxUint8), want: true, wantValue: int64(math.MaxUint8)},
		{value: uint16(math.MaxUint16), want: true, wantValue: int64(math.MaxUint16)},
		{value: uint32(math.MaxUint32), want: true, wantValue: int64(math.MaxUint32)},
		{value: uint64(math.MaxUint64), want: false},
		{value: uint(math.MaxUint), want: false},
		{value: uint(123), want: true, wantValue: int64(123)},
		{value: uint64(123), want: true, wantValue: int64(123)},
		{value: float64(math.MaxInt64), want: false},
		{value: float64(-math.MaxInt64), want: false},
		{value: float64(maxIntFloat64), want: true, wantValue: int64(maxIntFloat64)},
		{value: float64(-maxIntFloat64), want: true, wantValue: int64(-maxIntFloat64)},
		{value: float32(maxIntFloat32), want: true, wantValue: int64(maxIntFloat32)},
		{value: float32(-maxIntFloat32), want: true, wantValue: int64(-maxIntFloat32)},
		{value: fmt.Sprintf("%d", math.MaxInt), want: true, wantValue: int64(math.MaxInt)},
		{value: fmt.Sprintf("%d", math.MinInt), want: true, wantValue: int64(math.MinInt)},
		{value: 2.0, want: true, wantValue: int64(2)},
		{value: float32(2.0), want: true, wantValue: int64(2)},
		{value: "2", want: true, wantValue: int64(2)},
		{value: 2.5, want: false},
		{value: float32(2.5), want: false},
		{value: 'a', want: true, wantValue: int64(97)},
		{value: "string", want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Int64()
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

func TestUintValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Uint()
		assert.NotNil(t, v)
		assert.Equal(t, "uint", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: 2, want: true, wantValue: uint(2)},
		{value: int(math.MaxInt), want: true, wantValue: uint(math.MaxInt)},
		{value: int8(math.MaxInt8), want: true, wantValue: uint(math.MaxInt8)},
		{value: int16(math.MaxInt16), want: true, wantValue: uint(math.MaxInt16)},
		{value: int32(math.MaxInt32), want: true, wantValue: uint(math.MaxInt32)},
		{value: int64(math.MaxInt64), want: true, wantValue: uint(math.MaxInt64)},
		{value: int(math.MinInt), want: false},
		{value: int8(math.MinInt8), want: false},
		{value: int16(math.MinInt16), want: false},
		{value: int32(math.MinInt32), want: false},
		{value: int64(math.MinInt64), want: false},
		{value: uint8(math.MaxUint8), want: true, wantValue: uint(math.MaxUint8)},
		{value: uint16(math.MaxUint16), want: true, wantValue: uint(math.MaxUint16)},
		{value: uint32(math.MaxUint32), want: true, wantValue: uint(math.MaxUint32)},
		{value: uint64(math.MaxUint64), want: true, wantValue: uint(math.MaxUint64)},
		{value: uint(math.MaxUint), want: true, wantValue: uint(math.MaxUint)},
		{value: float64(math.MaxInt64), want: false},
		{value: float64(-math.MaxInt64), want: false},
		{value: float64(maxIntFloat64), want: true, wantValue: uint(maxIntFloat64)},
		{value: float64(-maxIntFloat64), want: false},
		{value: float32(maxIntFloat32), want: true, wantValue: uint(maxIntFloat32)},
		{value: float32(-maxIntFloat32), want: false},
		{value: float32(math.MaxInt32), want: false},
		{value: float32(math.MinInt32), want: false},
		{value: fmt.Sprintf("%d", uint(math.MaxUint)), want: true, wantValue: uint(math.MaxUint)},
		{value: fmt.Sprintf("%d", math.MinInt), want: false},
		{value: 2.0, want: true, wantValue: uint(2)},
		{value: float32(2.0), want: true, wantValue: uint(2)},
		{value: "2", want: true, wantValue: uint(2)},
		{value: 2.5, want: false},
		{value: float32(2.5), want: false},
		{value: 'a', want: true, wantValue: uint(97)},
		{value: "string", want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Uint()
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

func TestUint8Validator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Uint8()
		assert.NotNil(t, v)
		assert.Equal(t, "uint8", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: 2, want: true, wantValue: uint8(2)},
		{value: int(math.MaxInt), want: false},
		{value: int(math.MaxInt8), want: true, wantValue: uint8(math.MaxInt8)},
		{value: int8(math.MaxInt8), want: true, wantValue: uint8(math.MaxInt8)},
		{value: int16(math.MaxInt8), want: true, wantValue: uint8(math.MaxInt8)},
		{value: int32(math.MaxInt8), want: true, wantValue: uint8(math.MaxInt8)},
		{value: int64(math.MaxInt8), want: true, wantValue: uint8(math.MaxInt8)},
		{value: int16(math.MaxInt16), want: false},
		{value: int32(math.MaxInt32), want: false},
		{value: int64(math.MaxInt64), want: false},
		{value: int(math.MinInt), want: false},
		{value: int8(math.MinInt8), want: false},
		{value: int16(math.MinInt16), want: false},
		{value: int32(math.MinInt32), want: false},
		{value: int64(math.MinInt64), want: false},
		{value: uint8(math.MaxUint8), want: true, wantValue: uint8(math.MaxUint8)},
		{value: uint16(math.MaxUint8), want: true, wantValue: uint8(math.MaxUint8)},
		{value: uint32(math.MaxUint8), want: true, wantValue: uint8(math.MaxUint8)},
		{value: uint64(math.MaxUint8), want: true, wantValue: uint8(math.MaxUint8)},
		{value: uint(math.MaxUint8), want: true, wantValue: uint8(math.MaxUint8)},
		{value: uint16(math.MaxUint16), want: false},
		{value: uint32(math.MaxUint32), want: false},
		{value: uint64(math.MaxUint64), want: false},
		{value: uint(math.MaxUint), want: false},
		{value: float64(math.MaxInt64), want: false},
		{value: float64(-math.MaxInt64), want: false},
		{value: float32(maxIntFloat32), want: false},
		{value: float32(-maxIntFloat32), want: false},
		{value: float64(math.MaxUint8), want: true, wantValue: uint8(math.MaxUint8)},
		{value: float64(-math.MaxUint8), want: false},
		{value: float32(math.MaxUint8), want: true, wantValue: uint8(math.MaxUint8)},
		{value: float32(-math.MaxUint8), want: false},
		{value: fmt.Sprintf("%d", uint(math.MaxUint8)), want: true, wantValue: uint8(math.MaxUint8)},
		{value: fmt.Sprintf("%d", math.MinInt8), want: false},
		{value: 2.0, want: true, wantValue: uint8(2)},
		{value: float32(2.0), want: true, wantValue: uint8(2)},
		{value: "2", want: true, wantValue: uint8(2)},
		{value: 2.5, want: false},
		{value: float32(2.5), want: false},
		{value: 'a', want: true, wantValue: uint8(97)},
		{value: "string", want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Uint8()
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

func TestUint16Validator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Uint16()
		assert.NotNil(t, v)
		assert.Equal(t, "uint16", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: 2, want: true, wantValue: uint16(2)},
		{value: int(math.MaxInt), want: false},
		{value: int(math.MaxInt16), want: true, wantValue: uint16(math.MaxInt16)},
		{value: int8(math.MaxInt8), want: true, wantValue: uint16(math.MaxInt8)},
		{value: int16(math.MaxInt16), want: true, wantValue: uint16(math.MaxInt16)},
		{value: int32(math.MaxInt16), want: true, wantValue: uint16(math.MaxInt16)},
		{value: int64(math.MaxInt16), want: true, wantValue: uint16(math.MaxInt16)},
		{value: int32(math.MaxInt32), want: false},
		{value: int64(math.MaxInt64), want: false},
		{value: int(math.MinInt), want: false},
		{value: int8(math.MinInt8), want: false},
		{value: int16(math.MinInt16), want: false},
		{value: int32(math.MinInt32), want: false},
		{value: int64(math.MinInt64), want: false},
		{value: uint8(math.MaxUint8), want: true, wantValue: uint16(math.MaxUint8)},
		{value: uint16(math.MaxUint16), want: true, wantValue: uint16(math.MaxUint16)},
		{value: uint32(math.MaxUint16), want: true, wantValue: uint16(math.MaxUint16)},
		{value: uint64(math.MaxUint16), want: true, wantValue: uint16(math.MaxUint16)},
		{value: uint(math.MaxUint16), want: true, wantValue: uint16(math.MaxUint16)},
		{value: uint32(math.MaxUint32), want: false},
		{value: uint64(math.MaxUint64), want: false},
		{value: uint(math.MaxUint), want: false},
		{value: float64(math.MaxInt64), want: false},
		{value: float64(-math.MaxInt64), want: false},
		{value: float32(maxIntFloat32), want: false},
		{value: float32(-maxIntFloat32), want: false},
		{value: float64(math.MaxUint16), want: true, wantValue: uint16(math.MaxUint16)},
		{value: float64(-math.MaxUint16), want: false},
		{value: float32(math.MaxUint16), want: true, wantValue: uint16(math.MaxUint16)},
		{value: float32(-math.MaxUint16), want: false},
		{value: fmt.Sprintf("%d", uint(math.MaxUint16)), want: true, wantValue: uint16(math.MaxUint16)},
		{value: fmt.Sprintf("%d", math.MinInt16), want: false},
		{value: 2.0, want: true, wantValue: uint16(2)},
		{value: float32(2.0), want: true, wantValue: uint16(2)},
		{value: "2", want: true, wantValue: uint16(2)},
		{value: 2.5, want: false},
		{value: float32(2.5), want: false},
		{value: 'a', want: true, wantValue: uint16(97)},
		{value: "string", want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Uint16()
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

func TestUint32Validator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Uint32()
		assert.NotNil(t, v)
		assert.Equal(t, "uint32", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: 2, want: true, wantValue: uint32(2)},
		{value: int(math.MaxInt), want: false},
		{value: int(math.MaxInt32), want: true, wantValue: uint32(math.MaxInt32)},
		{value: int8(math.MaxInt8), want: true, wantValue: uint32(math.MaxInt8)},
		{value: int16(math.MaxInt16), want: true, wantValue: uint32(math.MaxInt16)},
		{value: int32(math.MaxInt32), want: true, wantValue: uint32(math.MaxInt32)},
		{value: int64(math.MaxInt32), want: true, wantValue: uint32(math.MaxInt32)},
		{value: int64(math.MaxInt64), want: false},
		{value: int(math.MinInt), want: false},
		{value: int8(math.MinInt8), want: false},
		{value: int16(math.MinInt16), want: false},
		{value: int32(math.MinInt32), want: false},
		{value: int64(math.MinInt64), want: false},
		{value: uint8(math.MaxUint8), want: true, wantValue: uint32(math.MaxUint8)},
		{value: uint16(math.MaxUint16), want: true, wantValue: uint32(math.MaxUint16)},
		{value: uint32(math.MaxUint32), want: true, wantValue: uint32(math.MaxUint32)},
		{value: uint64(math.MaxUint32), want: true, wantValue: uint32(math.MaxUint32)},
		{value: uint(math.MaxUint32), want: true, wantValue: uint32(math.MaxUint32)},
		{value: uint64(math.MaxUint64), want: false},
		{value: uint(math.MaxUint), want: false},
		{value: float64(math.MaxUint32), want: true, wantValue: uint32(math.MaxUint32)},
		{value: float64(-math.MaxInt32), want: false},
		{value: float32(maxIntFloat32), want: true, wantValue: uint32(maxIntFloat32)},
		{value: float32(-maxIntFloat32), want: false},
		{value: float64(math.MaxUint16), want: true, wantValue: uint32(math.MaxUint16)},
		{value: float64(-math.MaxUint16), want: false},
		{value: float32(math.MaxUint16), want: true, wantValue: uint32(math.MaxUint16)},
		{value: float32(-math.MaxUint16), want: false},
		{value: fmt.Sprintf("%d", uint(math.MaxUint32)), want: true, wantValue: uint32(math.MaxUint32)},
		{value: fmt.Sprintf("%d", math.MinInt32), want: false},
		{value: 2.0, want: true, wantValue: uint32(2)},
		{value: float32(2.0), want: true, wantValue: uint32(2)},
		{value: "2", want: true, wantValue: uint32(2)},
		{value: 2.5, want: false},
		{value: float32(2.5), want: false},
		{value: 'a', want: true, wantValue: uint32(97)},
		{value: "string", want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Uint32()
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

func TestUint64Validator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Uint64()
		assert.NotNil(t, v)
		assert.Equal(t, "uint64", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value     any
		wantValue any
		want      bool
	}{
		{value: 2, want: true, wantValue: uint64(2)},
		{value: int(math.MaxInt), want: true, wantValue: uint64(math.MaxInt)},
		{value: int8(math.MaxInt8), want: true, wantValue: uint64(math.MaxInt8)},
		{value: int16(math.MaxInt16), want: true, wantValue: uint64(math.MaxInt16)},
		{value: int32(math.MaxInt32), want: true, wantValue: uint64(math.MaxInt32)},
		{value: int64(math.MaxInt64), want: true, wantValue: uint64(math.MaxInt64)},
		{value: int(math.MinInt), want: false},
		{value: int8(math.MinInt8), want: false},
		{value: int16(math.MinInt16), want: false},
		{value: int32(math.MinInt32), want: false},
		{value: int64(math.MinInt64), want: false},
		{value: uint8(math.MaxUint8), want: true, wantValue: uint64(math.MaxUint8)},
		{value: uint16(math.MaxUint16), want: true, wantValue: uint64(math.MaxUint16)},
		{value: uint32(math.MaxUint32), want: true, wantValue: uint64(math.MaxUint32)},
		{value: uint64(math.MaxUint64), want: true, wantValue: uint64(math.MaxUint64)},
		{value: uint(math.MaxUint), want: true, wantValue: uint64(math.MaxUint)},
		{value: float64(math.MaxInt64), want: false},
		{value: float64(-math.MaxInt64), want: false},
		{value: float32(math.MaxInt32), want: false},
		{value: float32(math.MinInt32), want: false},
		{value: float64(maxIntFloat64), want: true, wantValue: uint64(maxIntFloat64)},
		{value: float64(-maxIntFloat64), want: false},
		{value: float32(maxIntFloat32), want: true, wantValue: uint64(maxIntFloat32)},
		{value: float32(-maxIntFloat32), want: false},
		{value: fmt.Sprintf("%d", uint(math.MaxUint)), want: true, wantValue: uint64(math.MaxUint)},
		{value: fmt.Sprintf("%d", math.MinInt), want: false},
		{value: 2.0, want: true, wantValue: uint64(2)},
		{value: float32(2.0), want: true, wantValue: uint64(2)},
		{value: "2", want: true, wantValue: uint64(2)},
		{value: 2.5, want: false},
		{value: float32(2.5), want: false},
		{value: 'a', want: true, wantValue: uint64(97)},
		{value: "string", want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Uint64()
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
