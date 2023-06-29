package validation

import (
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/lang"
)

func TestDateValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Date()
		assert.NotNil(t, v)
		assert.Equal(t, "date", v.Name())
		assert.True(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
		assert.Equal(t, []string{time.DateOnly}, v.Formats)

		v = Date(time.RFC3339, time.RFC3339Nano)
		assert.NotNil(t, v)
		assert.Empty(t, v.MessagePlaceholders(&Context{}))
		assert.Equal(t, []string{time.RFC3339, time.RFC3339Nano}, v.Formats)
	})

	ref3339 := lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z"))
	ref3339Nano := lo.Must(time.Parse(time.RFC3339Nano, "2023-03-15T10:07:42.123456789Z"))
	formats := []string{time.RFC3339, time.RFC3339Nano}
	cases := []struct {
		value     any
		wantValue any
		formats   []string
		want      bool
	}{
		{formats: formats, value: "2023-03-15T09:07:42Z", want: true, wantValue: lo.Must(time.Parse(time.RFC3339, "2023-03-15T09:07:42Z"))},
		{formats: formats, value: "2023-03-15T09:07:42.123456789Z", want: true, wantValue: lo.Must(time.Parse(time.RFC3339Nano, "2023-03-15T09:07:42.123456789Z"))},
		{formats: formats, value: ref3339, want: true, wantValue: ref3339},
		{formats: formats, value: ref3339Nano, want: true, wantValue: ref3339Nano},
		{formats: formats, value: "string", want: false},
		{formats: formats, value: "2023-03-15", want: false},
		{formats: []string{}, value: "2023-03-15", want: true, wantValue: lo.Must(time.Parse(time.DateOnly, "2023-03-15"))},
		{formats: formats, value: 'a', want: false},
		{formats: formats, value: 2, want: false},
		{formats: formats, value: 2.5, want: false},
		{formats: formats, value: []string{"string"}, want: false},
		{formats: formats, value: map[string]any{"a": 1}, want: false},
		{formats: formats, value: true, want: false},
		{formats: formats, value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Date(c.formats...)
			ctx := &Context{
				Value: c.value,
			}
			ok := v.Validate(ctx)
			if assert.Equal(t, c.want, ok) && ok {
				assert.Equal(t, c.wantValue, ctx.Value)
			}
		})
	}
}

func TestDateEqualsValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		now := time.Now()
		v := DateEquals(now)
		assert.NotNil(t, v)
		assert.Equal(t, "date_equals", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":date", now.Format(time.RFC3339)}, v.MessagePlaceholders(&Context{}))
	})

	ref := lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z"))
	cases := []struct {
		ref   time.Time
		value any
		want  bool
	}{
		{ref: ref, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z")), want: true},
		{ref: ref, value: ref, want: true},
		{ref: ref, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:08:42Z")), want: false},
		{ref: ref, value: "string", want: false},
		{ref: ref, value: 'a', want: false},
		{ref: ref, value: 2, want: false},
		{ref: ref, value: 2.5, want: false},
		{ref: ref, value: []string{"string"}, want: false},
		{ref: ref, value: map[string]any{"a": 1}, want: false},
		{ref: ref, value: true, want: false},
		{ref: ref, value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := DateEquals(c.ref)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}

func TestEqualsFieldValidator(t *testing.T) {
	path := "object.field[]"
	t.Run("Constructor", func(t *testing.T) {
		v := DateEqualsField(path)
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "date_equals", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":date", "field"}, v.MessagePlaceholders(&Context{}))

		assert.Panics(t, func() {
			DateEqualsField("invalid[path.")
		})
	})

	ref1 := lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z"))
	ref2 := lo.Must(time.Parse(time.RFC3339, "2023-03-16T10:07:42Z"))

	dataSingle := makeEqualsFieldData(ref1)
	dataTwo := makeEqualsFieldData(ref1, ref2)
	dataTwoSame := makeEqualsFieldData(ref1, ref1)
	dataNotTime := makeEqualsFieldData(ref1, "string")
	dataEmpty := makeEqualsFieldData()

	cases := []struct {
		data  map[string]any
		value any
		want  bool
	}{
		{data: dataSingle, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z")), want: true},
		{data: dataSingle, value: ref1, want: true},
		{data: dataSingle, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:08:42Z")), want: false},
		{data: dataTwo, value: lo.Must(time.Parse(time.RFC3339, "2023-03-16T10:07:42Z")), want: false},
		{data: dataTwo, value: ref2, want: false},
		{data: dataTwo, value: ref1, want: false},
		{data: dataTwoSame, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z")), want: true},
		{data: dataSingle, value: "string", want: false},
		{data: dataSingle, value: 'a', want: false},
		{data: dataSingle, value: 2, want: false},
		{data: dataSingle, value: 2.5, want: false},
		{data: dataSingle, value: []string{"string"}, want: false},
		{data: dataSingle, value: map[string]any{"a": 1}, want: false},
		{data: dataSingle, value: true, want: false},
		{data: dataSingle, value: nil, want: false},
		{data: dataNotTime, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z")), want: false},
		{data: dataEmpty, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z")), want: true},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := DateEqualsField(path)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
				Data:  c.data,
			}))
		})
	}
}

func makeEqualsFieldData(ref ...any) map[string]any {
	return map[string]any{
		"object": map[string]any{
			"field": ref,
		},
	}
}
