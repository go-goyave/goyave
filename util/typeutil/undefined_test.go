package typeutil

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/copier"
)

type testInt64 struct {
	Val int64
}

func (i *testInt64) UnmarshalText(text []byte) error {
	val, err := strconv.ParseInt(string(text), 10, 64)
	if err != nil {
		return err
	}
	i.Val = val
	return nil
}

func (i testInt64) CopyValue() any {
	return i.Val
}

func (i *testInt64) Scan(src any) error {
	val, ok := src.(int64)
	if !ok {
		return fmt.Errorf("src %#v is not int64", src)
	}
	i.Val = val
	return nil
}

type errValuer struct{}

func (e errValuer) Value() (driver.Value, error) {
	return nil, fmt.Errorf("errValuer")
}

func TestUndefined(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		u := NewUndefined("hello")
		assert.Equal(t, Undefined[string]{Val: "hello", Present: true}, u)
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		u := &Undefined[int64]{}

		require.NoError(t, json.Unmarshal([]byte("123456789"), u))
		assert.Equal(t, &Undefined[int64]{Val: 123456789, Present: true}, u)

		u = &Undefined[int64]{}
		require.Error(t, json.Unmarshal([]byte("\"notint\""), u))
		assert.Equal(t, &Undefined[int64]{Val: 0, Present: false}, u)
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		u := &Undefined[int64]{} // Not a text unmarshaler
		require.Error(t, u.UnmarshalText([]byte("123456789")))
		assert.Equal(t, &Undefined[int64]{Val: 0, Present: true}, u)

		u2 := &Undefined[testInt64]{}
		require.NoError(t, u2.UnmarshalText([]byte("123456789")))
		assert.Equal(t, &Undefined[testInt64]{Val: testInt64{Val: 123456789}, Present: true}, u2)

		u3 := &Undefined[testInt64]{}
		require.Error(t, u3.UnmarshalText([]byte("notint")))
		assert.Equal(t, &Undefined[testInt64]{Val: testInt64{Val: 0}, Present: true}, u3)
	})

	t.Run("IsZero", func(t *testing.T) {
		u := NewUndefined("hello")
		assert.False(t, u.IsZero())
		u.Present = false
		assert.True(t, u.IsZero())
	})

	t.Run("IsPresent", func(t *testing.T) {
		u := NewUndefined("hello")
		assert.True(t, u.IsPresent())
		u.Present = false
		assert.False(t, u.IsPresent())
	})

	t.Run("Value", func(t *testing.T) {

		cases := []struct {
			undefined driver.Valuer
			want      driver.Value
			wantErr   bool
		}{
			{undefined: NewUndefined("hello"), want: "hello", wantErr: false},
			{undefined: Undefined[string]{}, want: nil, wantErr: false},
			{undefined: Undefined[string]{Val: "hello", Present: false}, want: nil, wantErr: false},
			{undefined: NewUndefined([]string{"a", "b"}), want: []string{"a", "b"}, wantErr: false}, // Doesn't implement driver.Valuer
			{undefined: NewUndefined(sql.NullInt64{Int64: 123456789, Valid: true}), want: int64(123456789), wantErr: false},
			{undefined: NewUndefined(sql.NullInt64{}), want: nil, wantErr: false},
			{undefined: NewUndefined(errValuer{}), want: nil, wantErr: true},
		}

		for _, c := range cases {
			c := c
			v, err := c.undefined.Value()
			assert.Equal(t, c.want, v)
			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		}

	})

	t.Run("CopyValue", func(t *testing.T) {
		cases := []struct {
			undefined copier.Valuer
			want      any
		}{
			{undefined: NewUndefined("hello"), want: "hello"},
			{undefined: Undefined[string]{}, want: nil},
			{undefined: NewUndefined(testInt64{Val: 1234}), want: int64(1234)},
		}

		for _, c := range cases {
			c := c
			assert.Equal(t, c.want, c.undefined.CopyValue())
		}
	})

	t.Run("Scan", func(t *testing.T) {
		cases := []struct {
			undefined sql.Scanner
			value     any
			want      any
			wantErr   error
		}{
			{undefined: &Undefined[string]{}, value: "hello", want: lo.ToPtr(NewUndefined("hello"))},
			{undefined: &Undefined[string]{}, value: lo.ToPtr("hello"), want: lo.ToPtr(NewUndefined("hello"))},
			{undefined: &Undefined[testInt64]{}, value: int64(123), want: lo.ToPtr(NewUndefined(testInt64{Val: 123}))},
			{undefined: &Undefined[testInt64]{}, value: "hello", want: &Undefined[testInt64]{Present: true}, wantErr: fmt.Errorf("src \"hello\" is not int64")},                                      // Error coming from testInt64
			{undefined: &Undefined[int64]{}, value: "hello", want: &Undefined[int64]{Present: true}, wantErr: fmt.Errorf("typeutil.Undefined: Scan() incompatible types (src: string, dst: int64)")}, // Error coming from Undefined
			{undefined: &Undefined[int64]{Val: 123}, value: nil, want: &Undefined[int64]{Present: true, Val: 0}},
			{undefined: &Undefined[*int64]{Val: lo.ToPtr(int64(123))}, value: nil, want: &Undefined[*int64]{Present: true, Val: nil}},
		}

		for _, c := range cases {
			c := c
			err := c.undefined.Scan(c.value)
			if c.wantErr != nil {
				require.ErrorContains(t, err, c.wantErr.Error())
			}

			assert.Equal(t, c.want, c.undefined)
		}
	})

	t.Run("Default", func(t *testing.T) {
		u := NewUndefined("hello")
		assert.Equal(t, "hello", u.Default("world"))
		u.Present = false
		assert.Equal(t, "world", u.Default("world"))
	})
}
