package typeutil

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, c.want, res)
		})
	}

	t.Run("string", func(t *testing.T) {
		res, err := Convert[string]("hello")
		assert.Equal(t, "hello", res)
		require.NoError(t, err)
	})
	t.Run("int", func(t *testing.T) {
		res, err := Convert[int](123)
		assert.Equal(t, 123, res)
		require.NoError(t, err)
	})
	t.Run("float", func(t *testing.T) {
		res, err := Convert[float64](0.3)
		assert.InEpsilon(t, 0.3, res, 0)
		require.NoError(t, err)
	})
	t.Run("bool", func(t *testing.T) {
		res, err := Convert[bool](true)
		assert.True(t, res)
		require.NoError(t, err)
	})
	t.Run("mismatching types", func(t *testing.T) {
		res, err := Convert[bool]("true")
		assert.False(t, res)
		require.Error(t, err)
	})
	t.Run("[]string", func(t *testing.T) {
		res, err := Convert[[]string]([]string{"a", "b", "c"})
		assert.Equal(t, []string{"a", "b", "c"}, res)
		require.NoError(t, err)
	})
	t.Run("[]any", func(t *testing.T) {
		res, err := Convert[[]any]([]string{"a", "4", "c"})
		assert.Equal(t, []any{"a", "4", "c"}, res)
		require.NoError(t, err)

		res, err = Convert[[]any]([]any{"a", 4, 4.0, true, []any{"a", "b"}})
		assert.Equal(t, []any{"a", 4, 4.0, true, []any{"a", "b"}}, res)
		require.NoError(t, err)
	})
}

func TestMustConvert(t *testing.T) {
	assert.InEpsilon(t, 0.3, MustConvert[float64](0.3), 0)

	assert.Panics(t, func() {
		MustConvert[float64]("0.3")
	})
}

func TestCopy(t *testing.T) {
	type Nested struct {
		C uint `json:"c"`
	}

	type Promoted struct {
		P string `json:"p"`
	}

	type TestStruct struct {
		Promoted
		Undefined    Undefined[string]    `json:"undefined"`
		UndefinedPtr Undefined[*string]   `json:"undefinedPtr"`
		A            string               `json:"a"`
		Ptr          *string              `json:"ptr"`
		D            []string             `json:"d"`
		B            float64              `json:"b"`
		Scanner      Undefined[testInt64] `json:"scanner"`
		Nested       Nested               `json:"nested"`
	}

	cases := []struct {
		model     *TestStruct
		dto       any
		want      *TestStruct
		desc      string
		wantPanic bool
	}{
		{
			desc: "base",
			model: &TestStruct{
				A: "test",
				D: []string{"test1", "test2"},
				B: 1,
			},
			dto: struct {
				A string
				D []string
			}{A: "override", D: []string{"override1", "override2"}},
			want: &TestStruct{
				A: "override",
				D: []string{"override1", "override2"},
				B: 1,
			},
		},
		{
			desc:  "base_at_zero",
			model: &TestStruct{},
			dto: struct {
				B float64
			}{B: 1.234},
			want: &TestStruct{
				B: 1.234,
			},
		},
		{
			desc: "promoted",
			model: &TestStruct{
				A: "test",
				Promoted: Promoted{
					P: "promoted",
				},
			},
			dto: struct {
				A string
				P string
			}{A: "override", P: "promoted override"},
			want: &TestStruct{
				A: "override",
				Promoted: Promoted{
					P: "promoted override",
				},
			},
		},
		{
			desc: "promoted_dto",
			model: &TestStruct{
				A: "test",
				Promoted: Promoted{
					P: "promoted",
				},
			},
			dto: struct {
				A        string
				Promoted struct {
					P string
				}
			}{A: "override", Promoted: struct {
				P string
			}{
				P: "promoted override",
			}},
			want: &TestStruct{
				A: "override",
				Promoted: Promoted{
					P: "promoted override",
				},
			},
		},
		{
			desc: "ignore_empty",
			model: &TestStruct{
				A: "test",
				D: []string{"test1", "test2"},
				B: 0,
			},
			dto: struct {
				Ptr Undefined[*string]
				A   string
				D   []string
				B   float64
			}{A: "", B: 0, D: nil, Ptr: Undefined[*string]{}},
			want: &TestStruct{
				A: "test",
				D: []string{"test1", "test2"},
				B: 0,
			},
		},
		{
			desc: "deep",
			model: &TestStruct{
				Nested: Nested{
					C: 2,
				},
			},
			dto: struct {
				C      uint
				Nested struct {
					C uint
				}
			}{C: 3, Nested: struct{ C uint }{C: 4}},
			want: &TestStruct{
				Nested: Nested{
					C: 4,
				},
			},
		},
		{
			desc: "undefined_field_zero_value",
			model: &TestStruct{
				B: 1,
			},
			dto: struct{ B Undefined[float64] }{B: NewUndefined(0.0)},
			want: &TestStruct{
				B: 0,
			},
		},
		{
			desc: "undefined_field",
			model: &TestStruct{
				B: 1,
			},
			dto: struct{ B Undefined[float64] }{B: NewUndefined(1.234)},
			want: &TestStruct{
				B: 1.234,
			},
		},
		{
			desc:  "undefined_slice",
			model: &TestStruct{},
			dto:   struct{ D Undefined[[]string] }{D: NewUndefined([]string{"a", "b", "c"})},
			want: &TestStruct{
				D: []string{"a", "b", "c"},
			},
		},
		{
			desc:  "undefined_struct",
			model: &TestStruct{},
			dto:   struct{ Nested Undefined[struct{ C uint }] }{Nested: NewUndefined(struct{ C uint }{C: 4})},
			want: &TestStruct{
				Nested: Nested{
					C: 4,
				},
			},
		},
		{
			desc: "undefined_nil",
			model: &TestStruct{
				A:   "not nil",
				Ptr: lo.ToPtr("not nil"),
			},
			dto: struct {
				A   Undefined[*string]
				Ptr Undefined[*string]
			}{
				A:   NewUndefined[*string](nil),
				Ptr: NewUndefined[*string](nil),
			},
			want: &TestStruct{
				A:   "not nil",
				Ptr: nil,
			},
		},
		{
			desc: "undefined_to_undefined",
			model: &TestStruct{
				Undefined: NewUndefined("value"),
			},
			dto: struct {
				Undefined Undefined[string]
			}{
				Undefined: NewUndefined("override"),
			},
			want: &TestStruct{
				Undefined: NewUndefined("override"),
			},
		},
		{
			desc: "undefined_ptr_to_undefined",
			model: &TestStruct{
				Undefined: NewUndefined("value"),
			},
			dto: struct {
				Undefined Undefined[*string]
			}{
				Undefined: NewUndefined(lo.ToPtr("override")),
			},
			want: &TestStruct{
				Undefined: NewUndefined("override"),
			},
		},
		{
			desc: "ptr_to_undefined",
			model: &TestStruct{
				Undefined: NewUndefined("value"),
			},
			dto: struct {
				Undefined *string
			}{
				Undefined: lo.ToPtr("override"),
			},
			want: &TestStruct{
				Undefined: NewUndefined("override"),
			},
		},
		{
			desc: "undefined_ptr_to_undefined",
			model: &TestStruct{
				Undefined: NewUndefined("value"),
			},
			dto: struct {
				Undefined Undefined[*string]
			}{
				Undefined: NewUndefined(lo.ToPtr("override")),
			},
			want: &TestStruct{
				Undefined: NewUndefined("override"),
			},
		},
		{
			desc: "undefined_to_undefined_incompatible_types",
			model: &TestStruct{
				Undefined: NewUndefined("value"),
			},
			dto: struct {
				Undefined Undefined[int]
			}{
				Undefined: NewUndefined(123),
			},
			want: &TestStruct{
				Undefined: NewUndefined("value"), // The value has not been overridden because of incompatible types
			},
		},
		{
			desc: "scanner_undefined",
			model: &TestStruct{
				Scanner: NewUndefined(testInt64{Val: 123}),
			},
			dto: struct {
				Scanner int64
			}{
				Scanner: 456,
			},
			want: &TestStruct{
				Scanner: NewUndefined(testInt64{Val: 456}),
			},
		},
		{
			desc: "undefined_scanner_incompatible",
			model: &TestStruct{
				Scanner: NewUndefined(testInt64{Val: 123}),
			},
			dto: struct {
				Scanner string
			}{
				Scanner: "456",
			},
			want: &TestStruct{
				Scanner: NewUndefined(testInt64{Val: 123}),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			if c.wantPanic {
				assert.Panics(t, func() {
					Copy(c.model, c.dto)
				})
				return
			}
			res := Copy(c.model, c.dto)
			assert.Equal(t, c.want, res)
		})
	}
}
