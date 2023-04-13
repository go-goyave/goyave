package typeutil

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestUndefined(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		u := NewUndefined("hello")
		assert.Equal(t, Undefined[string]{Val: "hello", Present: true}, u)
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		u := &Undefined[int64]{}

		assert.NoError(t, json.Unmarshal([]byte("123456789"), u))
		assert.Equal(t, &Undefined[int64]{Val: 123456789, Present: true}, u)

		u = &Undefined[int64]{}
		assert.Error(t, json.Unmarshal([]byte("\"notint\""), u))
		assert.Equal(t, &Undefined[int64]{Val: 0, Present: false}, u)
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		u := &Undefined[int64]{} // Not a text unmarshaler
		assert.Error(t, u.UnmarshalText([]byte("123456789")))
		assert.Equal(t, &Undefined[int64]{Val: 0, Present: true}, u)

		u2 := &Undefined[testInt64]{}
		assert.NoError(t, u2.UnmarshalText([]byte("123456789")))
		assert.Equal(t, &Undefined[testInt64]{Val: testInt64{Val: 123456789}, Present: true}, u2)

		u3 := &Undefined[testInt64]{}
		assert.Error(t, u3.UnmarshalText([]byte("notint")))
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
		}

		for _, c := range cases {
			v, err := c.undefined.Value()
			assert.Equal(t, c.want, v)
			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		}

	})
}
