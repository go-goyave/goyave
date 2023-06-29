package typeutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvert(t *testing.T) {
	type Nested struct {
		C uint `json:"c"`
	}

	type Promoted struct {
		P string `json:"p"`
	}

	type TestStruct struct {
		Promoted
		A      string   `json:"a"`
		D      []string `json:"d"`
		B      float64  `json:"b"`
		Nested Nested   `json:"nested"`
	}

	cases := []struct {
		value   any
		want    any
		wantErr bool
	}{
		{
			value:   map[string]any{"p": "p", "a": "hello", "b": 0.3, "d": []string{"world"}, "c": 456, "nested": map[string]any{"c": 123}},
			want:    &TestStruct{Promoted: Promoted{P: "p"}, A: "hello", B: 0.3, D: []string{"world"}, Nested: Nested{C: 123}},
			wantErr: false,
		},
		{value: &TestStruct{A: "hello"}, want: &TestStruct{A: "hello"}, wantErr: false},
		{value: struct{}{}, want: &TestStruct{}, wantErr: false},
		{value: "string", want: &TestStruct{}, wantErr: true},
		{value: 'a', want: &TestStruct{}, wantErr: true},
		{value: 2, want: &TestStruct{}, wantErr: true},
		{value: 2.5, want: &TestStruct{}, wantErr: true},
		{value: []string{"string"}, want: &TestStruct{}, wantErr: true},
		{value: map[string]any{"a": 1}, want: &TestStruct{}, wantErr: true},
		{value: true, want: &TestStruct{}, wantErr: true},
		{value: nil, want: (*TestStruct)(nil), wantErr: false},
	}

	for _, c := range cases {
		t.Run("TestStruct", func(t *testing.T) {
			res, err := Convert[*TestStruct](c.value)
			assert.Equal(t, c.want, res)
			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

			}
			assert.Equal(t, c.want, res)
		})
	}

	t.Run("string", func(t *testing.T) {
		res, err := Convert[string]("hello")
		assert.Equal(t, "hello", res)
		assert.NoError(t, err)
	})
	t.Run("int", func(t *testing.T) {
		res, err := Convert[int](123)
		assert.Equal(t, 123, res)
		assert.NoError(t, err)
	})
	t.Run("float", func(t *testing.T) {
		res, err := Convert[float64](0.3)
		assert.Equal(t, 0.3, res)
		assert.NoError(t, err)
	})
	t.Run("bool", func(t *testing.T) {
		res, err := Convert[bool](true)
		assert.Equal(t, true, res)
		assert.NoError(t, err)
	})
	t.Run("mismatching types", func(t *testing.T) {
		res, err := Convert[bool]("true")
		assert.Equal(t, false, res)
		assert.Error(t, err)
	})
	t.Run("[]string", func(t *testing.T) {
		res, err := Convert[[]string]([]string{"a", "b", "c"})
		assert.Equal(t, []string{"a", "b", "c"}, res)
		assert.NoError(t, err)
	})
	t.Run("[]any", func(t *testing.T) {
		res, err := Convert[[]any]([]string{"a", "4", "c"})
		assert.Equal(t, []any{"a", "4", "c"}, res)
		assert.NoError(t, err)

		res, err = Convert[[]any]([]any{"a", 4, 4.0, true, []any{"a", "b"}})
		assert.Equal(t, []any{"a", 4, 4.0, true, []any{"a", "b"}}, res)
		assert.NoError(t, err)
	})
}

func TestMustConvert(t *testing.T) {
	assert.Equal(t, 0.3, MustConvert[float64](0.3))

	assert.Panics(t, func() {
		MustConvert[float64]("0.3")
	})
}
