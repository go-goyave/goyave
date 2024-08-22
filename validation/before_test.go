package validation

import (
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/lang"
)

func TestBeforeValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		now := time.Now()
		v := Before(now)
		assert.NotNil(t, v)
		assert.Equal(t, "before", v.Name())
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
		{ref: ref, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T09:07:42Z")), want: true},
		{ref: ref, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:08:42Z")), want: false},
		{ref: ref, value: ref, want: false},
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
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Before(c.ref)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}

func TestBeforeEqualValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		now := time.Now()
		v := BeforeEqual(now)
		assert.NotNil(t, v)
		assert.Equal(t, "before_equal", v.Name())
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
		{ref: ref, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T09:07:42Z")), want: true},
		{ref: ref, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:08:42Z")), want: false},
		{ref: ref, value: ref, want: true},
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
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := BeforeEqual(c.ref)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}

func TestBeforeFieldValidator(t *testing.T) {
	path := "object.field[]"
	t.Run("Constructor", func(t *testing.T) {
		v := BeforeField(path)
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "before", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":date", "field"}, v.MessagePlaceholders(&Context{}))

		assert.Panics(t, func() {
			BeforeField("invalid[path.")
		})
	})

	ref1 := lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z"))
	ref2 := lo.Must(time.Parse(time.RFC3339, "2023-03-16T10:07:42Z"))

	dataSingle := makeBeforeFieldData(ref1)
	dataTwo := makeBeforeFieldData(ref1, ref2)
	dataNotTime := makeBeforeFieldData(ref1, "string")

	cases := []struct {
		data  map[string]any
		value any
		want  bool
	}{
		{data: dataSingle, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T09:07:42Z")), want: true},
		{data: dataSingle, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:08:42Z")), want: false},
		{data: dataSingle, value: ref1, want: false},
		{data: dataTwo, value: lo.Must(time.Parse(time.RFC3339, "2023-03-16T10:06:42Z")), want: false},
		{data: dataTwo, value: ref1, want: false},
		{data: dataTwo, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:06:42Z")), want: true},
		{data: dataSingle, value: "string", want: false},
		{data: dataSingle, value: 'a', want: false},
		{data: dataSingle, value: 2, want: false},
		{data: dataSingle, value: 2.5, want: false},
		{data: dataSingle, value: []string{"string"}, want: false},
		{data: dataSingle, value: map[string]any{"a": 1}, want: false},
		{data: dataSingle, value: true, want: false},
		{data: dataSingle, value: nil, want: false},
		{data: dataNotTime, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T09:07:42Z")), want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := BeforeField(path)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
				Data:  c.data,
			}))
		})
	}
}

func TestBeforeEqualFieldValidator(t *testing.T) {
	path := "object.field[]"
	t.Run("Constructor", func(t *testing.T) {
		v := BeforeEqualField(path)
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "before_equal", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":date", "field"}, v.MessagePlaceholders(&Context{}))

		assert.Panics(t, func() {
			BeforeEqualField("invalid[path.")
		})
	})

	ref1 := lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:07:42Z"))
	ref2 := lo.Must(time.Parse(time.RFC3339, "2023-03-16T10:07:42Z"))

	dataSingle := makeBeforeFieldData(ref1)
	dataTwo := makeBeforeFieldData(ref1, ref2)
	dataNotTime := makeBeforeFieldData(ref1, "string")
	dataEmpty := makeBeforeFieldData()

	cases := []struct {
		data  map[string]any
		value any
		want  bool
	}{
		{data: dataSingle, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T09:07:42Z")), want: true},
		{data: dataSingle, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:08:42Z")), want: false},
		{data: dataTwo, value: lo.Must(time.Parse(time.RFC3339, "2023-03-16T10:06:42Z")), want: false},
		{data: dataTwo, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T10:06:42Z")), want: true},

		{data: dataSingle, value: ref1, want: true},
		{data: dataTwo, value: ref2, want: false},
		{data: dataTwo, value: ref1, want: true},

		{data: dataSingle, value: "string", want: false},
		{data: dataSingle, value: 'a', want: false},
		{data: dataSingle, value: 2, want: false},
		{data: dataSingle, value: 2.5, want: false},
		{data: dataSingle, value: []string{"string"}, want: false},
		{data: dataSingle, value: map[string]any{"a": 1}, want: false},
		{data: dataSingle, value: true, want: false},
		{data: dataSingle, value: nil, want: false},
		{data: dataNotTime, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T09:07:42Z")), want: false},
		{data: dataEmpty, value: lo.Must(time.Parse(time.RFC3339, "2023-03-15T09:07:42Z")), want: true},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := BeforeEqualField(path)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
				Data:  c.data,
			}))
		})
	}
}

func makeBeforeFieldData(ref ...any) map[string]any {
	return map[string]any{
		"object": map[string]any{
			"field": ref,
		},
	}
}
