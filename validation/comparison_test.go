package validation

import (
	"math"
	"mime/multipart"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/lang"
	"goyave.dev/goyave/v5/util/fsutil"
)

func makeComparisonData(ref ...any) map[string]any {
	return map[string]any{
		"object": map[string]any{
			"field": ref,
		},
	}
}

func TestGreaterThanValidator(t *testing.T) {
	path := "object.field[]"
	t.Run("Constructor", func(t *testing.T) {
		v := GreaterThan(path)
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "greater_than", v.Name())
		assert.False(t, v.IsType())
		assert.True(t, v.IsTypeDependent())
		assert.Equal(t, []string{":other", "field"}, v.MessagePlaceholders(&Context{}))

		assert.Panics(t, func() {
			GreaterThan("invalid[path.")
		})
	})

	largeFile := fsutil.File{
		Header: &multipart.FileHeader{
			Size: 4 * 1024,
		},
	}

	cases := []struct {
		value any
		data  map[string]any
		desc  string
		want  bool
	}{
		{desc: "length of two strings ok", data: makeComparisonData("ab"), value: "abc", want: true},
		{desc: "length of two strings nok", data: makeComparisonData("def"), value: "abc", want: false},
		{desc: "length of many strings ok", data: makeComparisonData("ab", "de"), value: "abc", want: true},
		{desc: "length of many strings nok", data: makeComparisonData("ab", "def"), value: "abc", want: false},
		{desc: "value of two int ok", data: makeComparisonData(3), value: 4, want: true},
		{desc: "value of two int nok", data: makeComparisonData(3), value: 2, want: false},
		{desc: "value of many int ok", data: makeComparisonData(3, 4), value: 5, want: true},
		{desc: "value of many int nok", data: makeComparisonData(3, 5), value: 4, want: false},
		{desc: "value uint overflow", data: nil, value: uint(math.MaxInt64), want: false},
		{desc: "compared value uint overflow", data: makeComparisonData(3, uint(math.MaxInt64)), value: 4, want: false},
		{desc: "value uint64 overflow", data: nil, value: uint64(math.MaxInt64), want: false},
		{desc: "compared value uint64 overflow", data: makeComparisonData(3, uint64(math.MaxInt64)), value: 4, want: false},
		{desc: "value of two float ok", data: makeComparisonData(3.045), value: 3.0451, want: true},
		{desc: "value of two float nok", data: makeComparisonData(3.045), value: 3.04499999, want: false},
		{desc: "value of many float ok", data: makeComparisonData(3.045, 3.046), value: 3.0461, want: true},
		{desc: "value of many float nok", data: makeComparisonData(3.045, 3.046), value: 3.0451, want: false},
		{desc: "float with int ok", data: makeComparisonData(3), value: 3.01, want: true},
		{desc: "float with int nok", data: makeComparisonData(3), value: 2.99, want: false},
		{desc: "float with many int ok", data: makeComparisonData(3, 4), value: 4.01, want: true},
		{desc: "float with many int nok", data: makeComparisonData(3, 4), value: 3.01, want: false},
		{desc: "int with float ok", data: makeComparisonData(2.99), value: 3, want: true},
		{desc: "int with float nok", data: makeComparisonData(3.01), value: 3, want: false},
		{desc: "int with many float ok", data: makeComparisonData(2.99, 3.99), value: 4, want: true},
		{desc: "int with many float nok", data: makeComparisonData(2.99, 3.99), value: 3, want: false},
		{desc: "string with number ok", data: makeComparisonData(2), value: "abc", want: true},
		{desc: "string with number nok", data: makeComparisonData(3), value: "abc", want: false},
		{desc: "string with many numbers ok", data: makeComparisonData(1, 2), value: "abc", want: true},
		{desc: "string with many numbers nok", data: makeComparisonData(2, 3), value: "abc", want: false},
		{desc: "number with string ok", data: makeComparisonData("abc"), value: 3.5, want: true},
		{desc: "number with string nok", data: makeComparisonData("abcd"), value: 4.0, want: false},
		{desc: "number with many strings ok", data: makeComparisonData("ab", "abc"), value: 3.5, want: true},
		{desc: "number with many strings nok", data: makeComparisonData("abc", "defg"), value: 3.5, want: false},
		{desc: "number with array ok", data: makeComparisonData([]string{"a", "b", "c"}), value: 3.5, want: true},
		{desc: "number with array nok", data: makeComparisonData([]string{"a", "b"}), value: 2.0, want: false},
		{desc: "number with many arrays ok", data: makeComparisonData([]string{"a", "b", "c"}, []string{"a", "b"}), value: 3.5, want: true},
		{desc: "number with many arrays nok", data: makeComparisonData([]string{"a", "b", "c"}, []string{"a", "b", "c", "d"}), value: 3.5, want: false},
		{desc: "string with array ok", data: makeComparisonData([]string{"a", "b"}), value: "abc", want: true},
		{desc: "string with array nok", data: makeComparisonData([]string{"a", "b", "c"}), value: "abc", want: false},
		{desc: "string with many arrays ok", data: makeComparisonData([]string{"a", "b", "c"}, []string{"a", "b"}), value: "abcd", want: true},
		{desc: "string with many arrays nok", data: makeComparisonData([]string{"a", "b", "c"}, []string{"a", "b"}), value: "abc", want: false},
		{desc: "array with number ok", data: makeComparisonData(1.6), value: []string{"a", "b"}, want: true},
		{desc: "array with number nok", data: makeComparisonData(2.0), value: []string{"a", "b"}, want: false},
		{desc: "array with many numbers ok", data: makeComparisonData(1.6, 1), value: []string{"a", "b"}, want: true},
		{desc: "array with many numbers nok", data: makeComparisonData(1.6, 2.1), value: []string{"a", "b"}, want: false},
		{desc: "array with string ok", data: makeComparisonData("a"), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with string nok", data: makeComparisonData("abc"), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with many strings ok", data: makeComparisonData("ab", "de"), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with many strings nok", data: makeComparisonData("abc", "de"), value: []string{"a", "b", "c"}, want: false},
		{desc: "number with number of keys in a map ok", data: makeComparisonData(map[string]any{"a": 0}), value: 2, want: true},
		{desc: "number with number of keys in a map nok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}), value: 2, want: false},
		{desc: "number with number of keys in many maps ok", data: makeComparisonData(map[string]any{"a": 0}, map[string]any{"c": 0}), value: 2, want: true},
		{desc: "number with number of keys in many maps nok", data: makeComparisonData(map[string]any{"a": 0}, map[string]any{"c": 0, "d": 1}), value: 2, want: false},
		{desc: "string with number of keys in a map ok", data: makeComparisonData(map[string]any{"a": 0}), value: "ab", want: true},
		{desc: "string with number of keys in a map nok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}), value: "ab", want: false},
		{desc: "string with number of keys in many maps ok", data: makeComparisonData(map[string]any{"a": 0}, map[string]any{"c": 0}), value: "ab", want: true},
		{desc: "string with number of keys in many maps nok", data: makeComparisonData(map[string]any{"a": 0}, map[string]any{"c": 0, "d": 1}), value: "ab", want: false},
		{desc: "number of keys in a map with number ok", data: makeComparisonData(1.5), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with number nok", data: makeComparisonData(2), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "number of keys in a map with many numbers ok", data: makeComparisonData(1.5, 1.9), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with many numbers nok", data: makeComparisonData(1.5, 2), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "number of keys in a map with string ok", data: makeComparisonData("a"), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with string nok", data: makeComparisonData("ab"), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "number of keys in a map with many strings ok", data: makeComparisonData("a", "b"), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with many strings nok", data: makeComparisonData("a", "bc"), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with object ok", data: makeComparisonData(map[string]any{"c": 0}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with object nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with many objects ok", data: makeComparisonData(map[string]any{"c": 0}, map[string]any{"d": 0}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with many objects nok", data: makeComparisonData(map[string]any{"c": 0}, map[string]any{"d": 0, "e": 1}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with array ok", data: makeComparisonData([]string{"a"}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with array nok", data: makeComparisonData([]string{"a", "b"}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with many arrays ok", data: makeComparisonData([]string{"a"}, []string{"b"}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with many arrays nok", data: makeComparisonData([]string{"a"}, []string{"b", "c"}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "array with object ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with object nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2}), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with many objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}, map[string]any{"e": 2}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with many objects nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}, map[string]any{"e": 2, "f": 3, "g": 4}), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with array ok", data: makeComparisonData([]string{"d", "e"}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with array nok", data: makeComparisonData([]string{"d", "e", "f"}), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with many arrays ok", data: makeComparisonData([]string{"d", "e"}, []string{"f", "g"}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with many arrays nok", data: makeComparisonData([]string{"d", "e"}, []string{"f", "g", "h"}), value: []string{"a", "b", "c"}, want: false},
		{desc: "file size with number ok", data: makeComparisonData(3), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with number nok", data: makeComparisonData(4), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with many numbers ok", data: makeComparisonData(3, 3.99), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with many numbers nok", data: makeComparisonData(3, 4.0), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with array ok", data: makeComparisonData([]string{"a", "b", "c"}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with array nok", data: makeComparisonData([]string{"a", "b", "c", "d"}), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with many arrays ok", data: makeComparisonData([]string{"a", "b"}, []string{"a", "b", "c"}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with many arrays nok", data: makeComparisonData([]string{"a", "b"}, []string{"a", "b", "c", "d"}), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with objects nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2, "f": 3}), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with many objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}, map[string]any{"e": 2, "f": 3}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with many objects nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}, map[string]any{"e": 2, "f": 3, "g": 4, "h": 5}), value: []fsutil.File{largeFile}, want: false},
		{desc: "number with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: 5, want: true},
		{desc: "number with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: 4, want: false},
		{desc: "string with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: "abcde", want: true},
		{desc: "string with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: "abcd", want: false},
		{desc: "array with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: []string{"a", "b", "c", "d", "e"}, want: true},
		{desc: "array with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: []string{"a", "b", "c", "d"}, want: false},
		{desc: "object with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: map[string]any{"a": 0, "b": 1, "c": 2, "d": 3, "e": 4}, want: true},
		{desc: "object with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: map[string]any{"e": 2, "f": 3, "g": 4, "h": 5}, want: false},
		{desc: "empty array", data: makeComparisonData(), value: "abc", want: true},

		{desc: "cannot validate bool", data: makeComparisonData(false), value: true, want: true},
		{desc: "cannot validate time", data: makeComparisonData(time.Now()), value: time.Now(), want: true},
		{desc: "cannot validate nil", data: makeComparisonData(nil), value: nil, want: true},
		{desc: "cannot validate nil", data: makeComparisonData(nil), value: 1, want: true},
		{desc: "cannot validate nil", data: makeComparisonData(1), value: nil, want: true},

		{desc: "not found", data: map[string]any{"object": map[string]any{}}, value: "abc", want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			v := GreaterThan(path)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
				Data:  c.data,
			}))
		})
	}
}

func TestGreaterThanEqualValidator(t *testing.T) {
	path := "object.field[]"
	t.Run("Constructor", func(t *testing.T) {
		v := GreaterThanEqual(path)
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "greater_than_equal", v.Name())
		assert.False(t, v.IsType())
		assert.True(t, v.IsTypeDependent())
		assert.Equal(t, []string{":other", "field"}, v.MessagePlaceholders(&Context{}))

		assert.Panics(t, func() {
			GreaterThanEqual("invalid[path.")
		})
	})

	largeFile := fsutil.File{
		Header: &multipart.FileHeader{
			Size: 4 * 1024,
		},
	}

	cases := []struct {
		value any
		data  map[string]any
		desc  string
		want  bool
	}{
		{desc: "length of two strings ok", data: makeComparisonData("ab"), value: "abc", want: true},
		{desc: "length of two strings ok", data: makeComparisonData("def"), value: "abc", want: true},
		{desc: "length of two strings nok", data: makeComparisonData("defg"), value: "abc", want: false},
		{desc: "length of many strings ok", data: makeComparisonData("ab", "def"), value: "abc", want: true},
		{desc: "length of many strings nok", data: makeComparisonData("ab", "defg"), value: "abc", want: false},
		{desc: "value of two int ok", data: makeComparisonData(3), value: 4, want: true},
		{desc: "value of two int ok", data: makeComparisonData(4), value: 4, want: true},
		{desc: "value of two int nok", data: makeComparisonData(3), value: 2, want: false},
		{desc: "value of many int ok", data: makeComparisonData(3, 5), value: 5, want: true},
		{desc: "value of many int nok", data: makeComparisonData(3, 5), value: 4, want: false},
		{desc: "value uint overflow", data: nil, value: uint(math.MaxInt64), want: false},
		{desc: "compared value uint overflow", data: makeComparisonData(3, uint(math.MaxInt64)), value: 4, want: false},
		{desc: "value uint64 overflow", data: nil, value: uint64(math.MaxInt64), want: false},
		{desc: "compared value uint64 overflow", data: makeComparisonData(3, uint64(math.MaxInt64)), value: 4, want: false},
		{desc: "value of two float ok", data: makeComparisonData(3.045), value: 3.0451, want: true},
		{desc: "value of two float ok", data: makeComparisonData(3.045), value: 3.045, want: true},
		{desc: "value of two float nok", data: makeComparisonData(3.045), value: 3.04499999, want: false},
		{desc: "value of many float ok", data: makeComparisonData(3.045, 3.0461), value: 3.0461, want: true},
		{desc: "value of many float nok", data: makeComparisonData(3.045, 3.046), value: 3.0451, want: false},
		{desc: "float with int ok", data: makeComparisonData(3), value: 3.01, want: true},
		{desc: "float with int ok", data: makeComparisonData(3), value: 3.0, want: true},
		{desc: "float with int nok", data: makeComparisonData(3), value: 2.99, want: false},
		{desc: "float with many int ok", data: makeComparisonData(3, 4), value: 4.0, want: true},
		{desc: "float with many int nok", data: makeComparisonData(3, 4), value: 3.01, want: false},
		{desc: "int with float ok", data: makeComparisonData(2.99), value: 3, want: true},
		{desc: "int with float ok", data: makeComparisonData(3), value: 3, want: true},
		{desc: "int with float nok", data: makeComparisonData(3.01), value: 3, want: false},
		{desc: "int with many float ok", data: makeComparisonData(2.99, 4), value: 4, want: true},
		{desc: "int with many float nok", data: makeComparisonData(2.99, 3.99), value: 3, want: false},
		{desc: "string with number ok", data: makeComparisonData(2), value: "abc", want: true},
		{desc: "string with number ok", data: makeComparisonData(3), value: "abc", want: true},
		{desc: "string with number nok", data: makeComparisonData(4), value: "abc", want: false},
		{desc: "string with many numbers ok", data: makeComparisonData(1, 3), value: "abc", want: true},
		{desc: "string with many numbers nok", data: makeComparisonData(2, 4), value: "abc", want: false},
		{desc: "number with string ok", data: makeComparisonData("abc"), value: 3.5, want: true},
		{desc: "number with string ok", data: makeComparisonData("abcd"), value: 4.0, want: true},
		{desc: "number with string nok", data: makeComparisonData("abcde"), value: 4.0, want: false},
		{desc: "number with many strings ok", data: makeComparisonData("ab", "abc"), value: 3.0, want: true},
		{desc: "number with many strings nok", data: makeComparisonData("abc", "defg"), value: 3.5, want: false},
		{desc: "number with array ok", data: makeComparisonData([]string{"a", "b", "c"}), value: 3.5, want: true},
		{desc: "number with array ok", data: makeComparisonData([]string{"a", "b", "c"}), value: 3.0, want: true},
		{desc: "number with array nok", data: makeComparisonData([]string{"a", "b", "c"}), value: 2.5, want: false},
		{desc: "number with many arrays ok", data: makeComparisonData([]string{"a", "b", "c"}, []string{"a", "b"}), value: 3.0, want: true},
		{desc: "number with many arrays nok", data: makeComparisonData([]string{"a", "b", "c"}, []string{"a", "b", "c", "d"}), value: 3.5, want: false},
		{desc: "string with array ok", data: makeComparisonData([]string{"a", "b"}), value: "abc", want: true},
		{desc: "string with array ok", data: makeComparisonData([]string{"a", "b", "c"}), value: "abc", want: true},
		{desc: "string with array nok", data: makeComparisonData([]string{"a", "b", "c", "d"}), value: "abc", want: false},
		{desc: "string with many arrays ok", data: makeComparisonData([]string{"a", "b", "c", "d"}, []string{"a", "b"}), value: "abcd", want: true},
		{desc: "string with many arrays nok", data: makeComparisonData([]string{"a", "b", "c", "d"}, []string{"a", "b"}), value: "abc", want: false},
		{desc: "array with number ok", data: makeComparisonData(1.6), value: []string{"a", "b"}, want: true},
		{desc: "array with number ok", data: makeComparisonData(2.0), value: []string{"a", "b"}, want: true},
		{desc: "array with number nok", data: makeComparisonData(2.2), value: []string{"a", "b"}, want: false},
		{desc: "array with many numbers ok", data: makeComparisonData(2.0, 1.6), value: []string{"a", "b"}, want: true},
		{desc: "array with many numbers nok", data: makeComparisonData(1.6, 2.1), value: []string{"a", "b"}, want: false},
		{desc: "array with string ok", data: makeComparisonData("a"), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with string ok", data: makeComparisonData("abc"), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with string nok", data: makeComparisonData("abcd"), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with many strings ok", data: makeComparisonData("ab", "def"), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with many strings nok", data: makeComparisonData("abcd", "de"), value: []string{"a", "b", "c"}, want: false},
		{desc: "number with number of keys in a map ok", data: makeComparisonData(map[string]any{"a": 0}), value: 2, want: true},
		{desc: "number with number of keys in a map ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}), value: 2, want: true},
		{desc: "number with number of keys in a map nok", data: makeComparisonData(map[string]any{"a": 0, "b": 1, "c": 2}), value: 2, want: false},
		{desc: "number with number of keys in many maps ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}, map[string]any{"c": 0}), value: 2, want: true},
		{desc: "number with number of keys in many maps nok", data: makeComparisonData(map[string]any{"a": 0}, map[string]any{"c": 0, "d": 1, "e": 2}), value: 2, want: false},
		{desc: "string with number of keys in a map ok", data: makeComparisonData(map[string]any{"a": 0}), value: "ab", want: true},
		{desc: "string with number of keys in a map ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}), value: "ab", want: true},
		{desc: "string with number of keys in a map nok", data: makeComparisonData(map[string]any{"a": 0, "b": 1, "c": 2}), value: "ab", want: false},
		{desc: "string with number of keys in many maps ok", data: makeComparisonData(map[string]any{"a": 0}, map[string]any{"c": 0, "d": 1}), value: "ab", want: true},
		{desc: "string with number of keys in many maps nok", data: makeComparisonData(map[string]any{"a": 0}, map[string]any{"c": 0, "d": 1, "e": 2}), value: "ab", want: false},
		{desc: "number of keys in a map with number ok", data: makeComparisonData(1.5), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with number ok", data: makeComparisonData(2), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with number nok", data: makeComparisonData(3), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "number of keys in a map with many numbers ok", data: makeComparisonData(1.5, 1.9, 2), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with many numbers nok", data: makeComparisonData(1.5, 2, 3), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "number of keys in a map with string ok", data: makeComparisonData("a"), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with string ok", data: makeComparisonData("ab"), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with string nok", data: makeComparisonData("abc"), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "number of keys in a map with many strings ok", data: makeComparisonData("a", "bc"), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with many strings nok", data: makeComparisonData("a", "bcd"), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with object ok", data: makeComparisonData(map[string]any{"c": 0}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with object ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with object nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with many objects ok", data: makeComparisonData(map[string]any{"c": 0}, map[string]any{"d": 0, "e": 1}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with many objects nok", data: makeComparisonData(map[string]any{"c": 0}, map[string]any{"d": 0, "e": 1, "f": 2}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with array ok", data: makeComparisonData([]string{"a"}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with array ok", data: makeComparisonData([]string{"a", "b"}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with array nok", data: makeComparisonData([]string{"a", "b", "c"}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with many arrays ok", data: makeComparisonData([]string{"a"}, []string{"b", "c"}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with many arrays nok", data: makeComparisonData([]string{"a"}, []string{"b", "c", "d"}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "array with object ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with object ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with object nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2, "f": 3}), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with many objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2}, map[string]any{"f": 2}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with many objects nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}, map[string]any{"e": 2, "f": 3, "g": 4, "h": 5}), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with array ok", data: makeComparisonData([]string{"d", "e"}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with array ok", data: makeComparisonData([]string{"d", "e", "f"}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with array nok", data: makeComparisonData([]string{"d", "e", "f", "g"}), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with many arrays ok", data: makeComparisonData([]string{"d", "e"}, []string{"f", "g", "h"}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with many arrays nok", data: makeComparisonData([]string{"d", "e"}, []string{"f", "g", "h", "i"}), value: []string{"a", "b", "c"}, want: false},
		{desc: "file size with number ok", data: makeComparisonData(3), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with number ok", data: makeComparisonData(4), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with number nok", data: makeComparisonData(4.1), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with many numbers ok", data: makeComparisonData(3, 3.99, 4), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with many numbers nok", data: makeComparisonData(3, 4.1), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with array ok", data: makeComparisonData([]string{"a", "b", "c"}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with array ok", data: makeComparisonData([]string{"a", "b", "c", "d"}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with array nok", data: makeComparisonData([]string{"a", "b", "c", "d", "e"}), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with many arrays ok", data: makeComparisonData([]string{"a", "b"}, []string{"a", "b", "c", "d"}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with many arrays nok", data: makeComparisonData([]string{"a", "b"}, []string{"a", "b", "c", "d", "e"}), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2, "f": 3}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with objects nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2, "f": 3, "g": 4}), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with many objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}, map[string]any{"e": 2, "f": 3, "g": 4, "h": 5}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with many objects nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}, map[string]any{"e": 2, "f": 3, "g": 4, "h": 5, "i": 6}), value: []fsutil.File{largeFile}, want: false},
		{desc: "number with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: 5, want: true},
		{desc: "number with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: 4, want: true},
		{desc: "number with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: 3.99, want: false},
		{desc: "string with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: "abcde", want: true},
		{desc: "string with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: "abcd", want: true},
		{desc: "string with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: "abc", want: false},
		{desc: "array with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: []string{"a", "b", "c", "d", "e"}, want: true},
		{desc: "array with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: []string{"a", "b", "c", "d"}, want: true},
		{desc: "array with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: []string{"a", "b", "c"}, want: false},
		{desc: "object with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: map[string]any{"a": 0, "b": 1, "c": 2, "d": 3, "e": 4}, want: true},
		{desc: "object with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: map[string]any{"a": 0, "b": 1, "c": 2, "d": 3}, want: true},
		{desc: "object with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: map[string]any{"e": 2, "f": 3, "g": 4}, want: false},
		{desc: "empty array", data: makeComparisonData(), value: "abc", want: true},

		{desc: "cannot validate bool", data: makeComparisonData(false), value: true, want: true},
		{desc: "cannot validate time", data: makeComparisonData(time.Now()), value: time.Now(), want: true},
		{desc: "cannot validate nil", data: makeComparisonData(nil), value: nil, want: true},
		{desc: "cannot validate nil", data: makeComparisonData(nil), value: 1, want: true},
		{desc: "cannot validate nil", data: makeComparisonData(1), value: nil, want: true},

		{desc: "not found", data: map[string]any{"object": map[string]any{}}, value: "abc", want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			v := GreaterThanEqual(path)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
				Data:  c.data,
			}))
		})
	}
}

func TestLowerThanValidator(t *testing.T) {
	path := "object.field[]"
	t.Run("Constructor", func(t *testing.T) {
		v := LowerThan(path)
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "lower_than", v.Name())
		assert.False(t, v.IsType())
		assert.True(t, v.IsTypeDependent())
		assert.Equal(t, []string{":other", "field"}, v.MessagePlaceholders(&Context{}))

		assert.Panics(t, func() {
			LowerThan("invalid[path.")
		})
	})

	largeFile := fsutil.File{
		Header: &multipart.FileHeader{
			Size: 4 * 1024,
		},
	}

	cases := []struct {
		value any
		data  map[string]any
		desc  string
		want  bool
	}{
		{desc: "length of two strings ok", data: makeComparisonData("abc"), value: "ab", want: true},
		{desc: "length of two strings nok", data: makeComparisonData("def"), value: "abc", want: false},
		{desc: "length of many strings ok", data: makeComparisonData("abc", "def"), value: "ab", want: true},
		{desc: "length of many strings nok", data: makeComparisonData("ab", "def"), value: "ab", want: false},
		{desc: "value of two int ok", data: makeComparisonData(3), value: 2, want: true},
		{desc: "value of two int nok", data: makeComparisonData(3), value: 3, want: false},
		{desc: "value of many int ok", data: makeComparisonData(3, 4), value: 2, want: true},
		{desc: "value of many int nok", data: makeComparisonData(3, 5), value: 3, want: false},
		{desc: "value uint overflow", data: nil, value: uint(math.MaxInt64), want: false},
		{desc: "compared value uint overflow", data: makeComparisonData(3, uint(math.MaxInt64)), value: 2, want: false},
		{desc: "value uint64 overflow", data: nil, value: uint64(math.MaxInt64), want: false},
		{desc: "compared value uint64 overflow", data: makeComparisonData(3, uint64(math.MaxInt64)), value: 2, want: false},
		{desc: "value of two float ok", data: makeComparisonData(3.045), value: 3.0449, want: true},
		{desc: "value of two float nok", data: makeComparisonData(3.045), value: 3.0451, want: false},
		{desc: "value of many float ok", data: makeComparisonData(3.045, 3.046), value: 3.0444, want: true},
		{desc: "value of many float nok", data: makeComparisonData(3.045, 3.046), value: 3.045, want: false},
		{desc: "float with int ok", data: makeComparisonData(3), value: 2.99, want: true},
		{desc: "float with int nok", data: makeComparisonData(3), value: 3.01, want: false},
		{desc: "float with many int ok", data: makeComparisonData(3, 4), value: 2.99, want: true},
		{desc: "float with many int nok", data: makeComparisonData(3, 4), value: 3.0, want: false},
		{desc: "int with float ok", data: makeComparisonData(3.01), value: 3, want: true},
		{desc: "int with float nok", data: makeComparisonData(2.99), value: 3, want: false},
		{desc: "int with many float ok", data: makeComparisonData(2.99, 3.99), value: 2.98, want: true},
		{desc: "int with many float nok", data: makeComparisonData(2.99, 3.99), value: 2.99, want: false},
		{desc: "string with number ok", data: makeComparisonData(3), value: "ab", want: true},
		{desc: "string with number nok", data: makeComparisonData(3), value: "abc", want: false},
		{desc: "string with many numbers ok", data: makeComparisonData(3, 4), value: "ab", want: true},
		{desc: "string with many numbers nok", data: makeComparisonData(2, 3), value: "ab", want: false},
		{desc: "number with string ok", data: makeComparisonData("abc"), value: 2.5, want: true},
		{desc: "number with string nok", data: makeComparisonData("abcd"), value: 4.0, want: false},
		{desc: "number with many strings ok", data: makeComparisonData("abc", "abcd"), value: 2.5, want: true},
		{desc: "number with many strings nok", data: makeComparisonData("abc", "defg"), value: 3.0, want: false},
		{desc: "number with array ok", data: makeComparisonData([]string{"a", "b", "c"}), value: 2.5, want: true},
		{desc: "number with array nok", data: makeComparisonData([]string{"a", "b", "c"}), value: 3.0, want: false},
		{desc: "number with many arrays ok", data: makeComparisonData([]string{"a", "b", "c"}, []string{"a", "b"}), value: 1.5, want: true},
		{desc: "number with many arrays nok", data: makeComparisonData([]string{"a", "b", "c"}, []string{"a", "b", "c", "d"}), value: 3.0, want: false},
		{desc: "string with array ok", data: makeComparisonData([]string{"a", "b", "c"}), value: "ab", want: true},
		{desc: "string with array nok", data: makeComparisonData([]string{"a", "b", "c"}), value: "abc", want: false},
		{desc: "string with many arrays ok", data: makeComparisonData([]string{"a", "b", "c", "d"}, []string{"a", "b", "c"}), value: "ab", want: true},
		{desc: "string with many arrays nok", data: makeComparisonData([]string{"a", "b", "c"}, []string{"a", "b"}), value: "ab", want: false},
		{desc: "array with number ok", data: makeComparisonData(2.6), value: []string{"a", "b"}, want: true},
		{desc: "array with number nok", data: makeComparisonData(2.0), value: []string{"a", "b"}, want: false},
		{desc: "array with many numbers ok", data: makeComparisonData(2.6, 3), value: []string{"a", "b"}, want: true},
		{desc: "array with many numbers nok", data: makeComparisonData(1.6, 2.1), value: []string{"a", "b"}, want: false},
		{desc: "array with string ok", data: makeComparisonData("abcd"), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with string nok", data: makeComparisonData("abc"), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with many strings ok", data: makeComparisonData("abc", "def"), value: []string{"a", "b"}, want: true},
		{desc: "array with many strings nok", data: makeComparisonData("abc", "defg"), value: []string{"a", "b", "c"}, want: false},
		{desc: "number with number of keys in a map ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}), value: 1, want: true},
		{desc: "number with number of keys in a map nok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}), value: 2, want: false},
		{desc: "number with number of keys in many maps ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}, map[string]any{"c": 0, "d": 1, "e": 2}), value: 1, want: true},
		{desc: "number with number of keys in many maps nok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}, map[string]any{"c": 0, "d": 1, "e": 2}), value: 2, want: false},
		{desc: "string with number of keys in a map ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1, "c": 2}), value: "ab", want: true},
		{desc: "string with number of keys in a map nok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}), value: "ab", want: false},
		{desc: "string with number of keys in many maps ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}, map[string]any{"c": 0, "d": 1, "e": 2}), value: "a", want: true},
		{desc: "string with number of keys in many maps nok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}, map[string]any{"c": 0, "d": 1, "e": 2}), value: "ab", want: false},
		{desc: "number of keys in a map with number ok", data: makeComparisonData(2.5), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with number nok", data: makeComparisonData(2), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "number of keys in a map with many numbers ok", data: makeComparisonData(2.5, 2.9), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with many numbers nok", data: makeComparisonData(2.5, 2), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "number of keys in a map with string ok", data: makeComparisonData("abc"), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with string nok", data: makeComparisonData("ab"), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "number of keys in a map with many strings ok", data: makeComparisonData("abc", "defg"), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with many strings nok", data: makeComparisonData("ab", "cde"), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with object ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: map[string]any{"a": 0}, want: true},
		{desc: "object with object nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with many objects ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}, map[string]any{"c": 0, "d": 1, "e": 2}), value: map[string]any{"a": 0}, want: true},
		{desc: "object with many objects nok", data: makeComparisonData(map[string]any{"a": 0, "b": 1, "c": 2}, map[string]any{"d": 0, "e": 1}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with array ok", data: makeComparisonData([]string{"a", "b"}), value: map[string]any{"a": 0}, want: true},
		{desc: "object with array nok", data: makeComparisonData([]string{"a", "b"}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with many arrays ok", data: makeComparisonData([]string{"a", "b"}, []string{"c", "d", "e"}), value: map[string]any{"a": 0}, want: true},
		{desc: "object with many arrays nok", data: makeComparisonData([]string{"a"}, []string{"b", "c"}), value: map[string]any{"a": 0}, want: false},
		{desc: "array with object ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: []string{"a"}, want: true},
		{desc: "array with object nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: []string{"a", "b"}, want: false},
		{desc: "array with many objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}, map[string]any{"e": 2, "f": 3}), value: []string{"a"}, want: true},
		{desc: "array with many objects nok", data: makeComparisonData(map[string]any{"c": 0}, map[string]any{"e": 2, "f": 3, "g": 4}), value: []string{"a"}, want: false},
		{desc: "array with array ok", data: makeComparisonData([]string{"d", "e"}), value: []string{"a"}, want: true},
		{desc: "array with array nok", data: makeComparisonData([]string{"d", "e", "f"}), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with many arrays ok", data: makeComparisonData([]string{"d", "e"}, []string{"f", "g"}), value: []string{"a"}, want: true},
		{desc: "array with many arrays nok", data: makeComparisonData([]string{"d", "e"}, []string{"f", "g", "h"}), value: []string{"a", "b"}, want: false},
		{desc: "file size with number ok", data: makeComparisonData(5), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with number nok", data: makeComparisonData(4), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with many numbers ok", data: makeComparisonData(5, 4.01), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with many numbers nok", data: makeComparisonData(5, 4.0), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with array ok", data: makeComparisonData([]string{"a", "b", "c", "d", "e"}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with array nok", data: makeComparisonData([]string{"a", "b", "c", "d"}), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with many arrays ok", data: makeComparisonData([]string{"a", "b", "c", "d", "e"}, []string{"a", "b", "c", "d", "e"}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with many arrays nok", data: makeComparisonData([]string{"a", "b", "c", "d"}, []string{"a", "b", "c", "d", "e"}), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2, "f": 3, "g": 4}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with objects nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2, "f": 3}), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with many objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2, "f": 3, "g": 4}, map[string]any{"c": 0, "d": 1, "e": 2, "f": 3, "g": 4}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with many objects nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2, "f": 3}, map[string]any{"e": 2, "f": 3, "g": 4, "h": 5, "i": 6}), value: []fsutil.File{largeFile}, want: false},
		{desc: "number with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: 3, want: true},
		{desc: "number with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: 4, want: false},
		{desc: "string with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: "abc", want: true},
		{desc: "string with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: "abcd", want: false},
		{desc: "array with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: []string{"a", "b", "c", "d"}, want: false},
		{desc: "object with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: map[string]any{"a": 0, "b": 1, "c": 2}, want: true},
		{desc: "object with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: map[string]any{"e": 2, "f": 3, "g": 4, "h": 5}, want: false},
		{desc: "empty array", data: makeComparisonData(), value: "abc", want: true},

		{desc: "cannot validate bool", data: makeComparisonData(false), value: true, want: true},
		{desc: "cannot validate time", data: makeComparisonData(time.Now()), value: time.Now(), want: true},
		{desc: "cannot validate nil", data: makeComparisonData(nil), value: nil, want: true},
		{desc: "cannot validate nil", data: makeComparisonData(nil), value: 1, want: true},
		{desc: "cannot validate nil", data: makeComparisonData(1), value: nil, want: true},

		{desc: "not found", data: map[string]any{"object": map[string]any{}}, value: "abc", want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			v := LowerThan(path)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
				Data:  c.data,
			}))
		})
	}
}

func TestLowerThanEqualValidator(t *testing.T) {
	path := "object.field[]"
	t.Run("Constructor", func(t *testing.T) {
		v := LowerThanEqual(path)
		v.lang = &lang.Language{}
		assert.NotNil(t, v)
		assert.Equal(t, "lower_than_equal", v.Name())
		assert.False(t, v.IsType())
		assert.True(t, v.IsTypeDependent())
		assert.Equal(t, []string{":other", "field"}, v.MessagePlaceholders(&Context{}))

		assert.Panics(t, func() {
			LowerThanEqual("invalid[path.")
		})
	})

	largeFile := fsutil.File{
		Header: &multipart.FileHeader{
			Size: 4 * 1024,
		},
	}

	cases := []struct {
		value any
		data  map[string]any
		desc  string
		want  bool
	}{
		{desc: "length of two strings ok", data: makeComparisonData("ab"), value: "a", want: true},
		{desc: "length of two strings ok", data: makeComparisonData("def"), value: "abc", want: true},
		{desc: "length of two strings nok", data: makeComparisonData("efg"), value: "abcd", want: false},
		{desc: "length of many strings ok", data: makeComparisonData("ab", "c"), value: "a", want: true},
		{desc: "length of many strings nok", data: makeComparisonData("ab", "efg"), value: "abc", want: false},
		{desc: "value of two int ok", data: makeComparisonData(3), value: 2, want: true},
		{desc: "value of two int ok", data: makeComparisonData(4), value: 4, want: true},
		{desc: "value of two int nok", data: makeComparisonData(3), value: 4, want: false},
		{desc: "value of many int ok", data: makeComparisonData(3, 5), value: 3, want: true},
		{desc: "value of many int nok", data: makeComparisonData(3, 5), value: 4, want: false},
		{desc: "value uint overflow", data: nil, value: uint(math.MaxInt64), want: false},
		{desc: "compared value uint overflow", data: makeComparisonData(3, uint(math.MaxInt64)), value: 3, want: false},
		{desc: "value uint64 overflow", data: nil, value: uint64(math.MaxInt64), want: false},
		{desc: "compared value uint64 overflow", data: makeComparisonData(3, uint64(math.MaxInt64)), value: 3, want: false},
		{desc: "value of two float ok", data: makeComparisonData(3.045), value: 3.0444, want: true},
		{desc: "value of two float ok", data: makeComparisonData(3.045), value: 3.045, want: true},
		{desc: "value of two float nok", data: makeComparisonData(3.045), value: 3.0451, want: false},
		{desc: "value of many float ok", data: makeComparisonData(3.045, 3.0461), value: 3.045, want: true},
		{desc: "value of many float nok", data: makeComparisonData(3.045, 3.046), value: 3.046, want: false},
		{desc: "float with int ok", data: makeComparisonData(3), value: 2.99, want: true},
		{desc: "float with int ok", data: makeComparisonData(3), value: 3.0, want: true},
		{desc: "float with int nok", data: makeComparisonData(3), value: 3.01, want: false},
		{desc: "float with many int ok", data: makeComparisonData(3, 4), value: 3.0, want: true},
		{desc: "float with many int nok", data: makeComparisonData(3, 4), value: 3.01, want: false},
		{desc: "int with float ok", data: makeComparisonData(3.01), value: 3, want: true},
		{desc: "int with float ok", data: makeComparisonData(3), value: 3, want: true},
		{desc: "int with float nok", data: makeComparisonData(2.99), value: 3, want: false},
		{desc: "int with many float ok", data: makeComparisonData(2.99, 2), value: 2, want: true},
		{desc: "int with many float nok", data: makeComparisonData(2.99, 3.99), value: 3, want: false},
		{desc: "string with number ok", data: makeComparisonData(3), value: "ab", want: true},
		{desc: "string with number ok", data: makeComparisonData(3), value: "abc", want: true},
		{desc: "string with number nok", data: makeComparisonData(2), value: "abc", want: false},
		{desc: "string with many numbers ok", data: makeComparisonData(2, 3), value: "ab", want: true},
		{desc: "string with many numbers nok", data: makeComparisonData(2, 4), value: "abc", want: false},
		{desc: "number with string ok", data: makeComparisonData("abc"), value: 2.5, want: true},
		{desc: "number with string ok", data: makeComparisonData("abcd"), value: 4.0, want: true},
		{desc: "number with string nok", data: makeComparisonData("abc"), value: 4.0, want: false},
		{desc: "number with many strings ok", data: makeComparisonData("abc", "de"), value: 2.0, want: true},
		{desc: "number with many strings nok", data: makeComparisonData("abc", "de"), value: 2.5, want: false},
		{desc: "number with array ok", data: makeComparisonData([]string{"a", "b", "c"}), value: 2.5, want: true},
		{desc: "number with array ok", data: makeComparisonData([]string{"a", "b", "c"}), value: 3.0, want: true},
		{desc: "number with array nok", data: makeComparisonData([]string{"a", "b", "c"}), value: 3.5, want: false},
		{desc: "number with many arrays ok", data: makeComparisonData([]string{"a", "b", "c"}, []string{"a", "b"}), value: 2.0, want: true},
		{desc: "number with many arrays nok", data: makeComparisonData([]string{"a", "b", "c"}, []string{"a", "b", "c", "d"}), value: 3.5, want: false},
		{desc: "string with array ok", data: makeComparisonData([]string{"a", "b"}), value: "a", want: true},
		{desc: "string with array ok", data: makeComparisonData([]string{"a", "b", "c"}), value: "abc", want: true},
		{desc: "string with array nok", data: makeComparisonData([]string{"a", "b"}), value: "abc", want: false},
		{desc: "string with many arrays ok", data: makeComparisonData([]string{"a", "b", "c", "d"}, []string{"a", "b"}), value: "ab", want: true},
		{desc: "string with many arrays nok", data: makeComparisonData([]string{"a", "b", "c", "d"}, []string{"a", "b"}), value: "abc", want: false},
		{desc: "array with number ok", data: makeComparisonData(2.6), value: []string{"a", "b"}, want: true},
		{desc: "array with number ok", data: makeComparisonData(2.0), value: []string{"a", "b"}, want: true},
		{desc: "array with number nok", data: makeComparisonData(1.2), value: []string{"a", "b"}, want: false},
		{desc: "array with many numbers ok", data: makeComparisonData(2.0, 2.6), value: []string{"a", "b"}, want: true},
		{desc: "array with many numbers nok", data: makeComparisonData(1.6, 2.1), value: []string{"a", "b"}, want: false},
		{desc: "array with string ok", data: makeComparisonData("ab"), value: []string{"a"}, want: true},
		{desc: "array with string ok", data: makeComparisonData("abc"), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with string nok", data: makeComparisonData("ab"), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with many strings ok", data: makeComparisonData("a", "bc"), value: []string{"a"}, want: true},
		{desc: "array with many strings nok", data: makeComparisonData("abc", "de"), value: []string{"a", "b", "c"}, want: false},
		{desc: "number with number of keys in a map ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}), value: 1, want: true},
		{desc: "number with number of keys in a map ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}), value: 2, want: true},
		{desc: "number with number of keys in a map nok", data: makeComparisonData(map[string]any{"a": 0, "b": 1, "c": 2}), value: 4, want: false},
		{desc: "number with number of keys in many maps ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}, map[string]any{"c": 0}), value: 1, want: true},
		{desc: "number with number of keys in many maps nok", data: makeComparisonData(map[string]any{"a": 0}, map[string]any{"c": 0, "d": 1, "e": 2}), value: 3, want: false},
		{desc: "string with number of keys in a map ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}), value: "a", want: true},
		{desc: "string with number of keys in a map ok", data: makeComparisonData(map[string]any{"a": 0, "b": 1}), value: "ab", want: true},
		{desc: "string with number of keys in a map nok", data: makeComparisonData(map[string]any{"a": 0, "b": 1, "c": 2}), value: "abcd", want: false},
		{desc: "string with number of keys in many maps ok", data: makeComparisonData(map[string]any{"a": 0}, map[string]any{"c": 0, "d": 1}), value: "a", want: true},
		{desc: "string with number of keys in many maps nok", data: makeComparisonData(map[string]any{"a": 0}, map[string]any{"c": 0, "d": 1, "e": 2}), value: "ab", want: false},
		{desc: "number of keys in a map with number ok", data: makeComparisonData(2.5), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with number ok", data: makeComparisonData(2), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with number nok", data: makeComparisonData(1), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "number of keys in a map with many numbers ok", data: makeComparisonData(2.5, 2.01, 2), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with many numbers nok", data: makeComparisonData(1.5, 2, 3), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "number of keys in a map with string ok", data: makeComparisonData("abc"), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with string ok", data: makeComparisonData("ab"), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "number of keys in a map with string nok", data: makeComparisonData("a"), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "number of keys in a map with many strings ok", data: makeComparisonData("a", "bc"), value: map[string]any{"a": 0}, want: true},
		{desc: "number of keys in a map with many strings nok", data: makeComparisonData("a", "bcd"), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with object ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: map[string]any{"a": 0}, want: true},
		{desc: "object with object ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with object nok", data: makeComparisonData(map[string]any{"c": 0}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with many objects ok", data: makeComparisonData(map[string]any{"c": 0}, map[string]any{"d": 0, "e": 1}), value: map[string]any{"a": 0}, want: true},
		{desc: "object with many objects nok", data: makeComparisonData(map[string]any{"c": 0}, map[string]any{"d": 0, "e": 1, "f": 2}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with array ok", data: makeComparisonData([]string{"a", "b", "c"}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with array ok", data: makeComparisonData([]string{"a", "b"}), value: map[string]any{"a": 0, "b": 1}, want: true},
		{desc: "object with array nok", data: makeComparisonData([]string{"a"}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "object with many arrays ok", data: makeComparisonData([]string{"a"}, []string{"b", "c"}), value: map[string]any{"a": 0}, want: true},
		{desc: "object with many arrays nok", data: makeComparisonData([]string{"a"}, []string{"b", "c", "d"}), value: map[string]any{"a": 0, "b": 1}, want: false},
		{desc: "array with object ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}), value: []string{"a"}, want: true},
		{desc: "array with object ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with object nok", data: makeComparisonData(map[string]any{"c": 0}), value: []string{"a", "b"}, want: false},
		{desc: "array with many objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2}, map[string]any{"f": 2, "g": 3}), value: []string{"a", "b"}, want: true},
		{desc: "array with many objects nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}, map[string]any{"e": 2, "f": 3, "g": 4, "h": 5}), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with array ok", data: makeComparisonData([]string{"d", "e"}), value: []string{"a"}, want: true},
		{desc: "array with array ok", data: makeComparisonData([]string{"d", "e", "f"}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with array nok", data: makeComparisonData([]string{"d", "e"}), value: []string{"a", "b", "c"}, want: false},
		{desc: "array with many arrays ok", data: makeComparisonData([]string{"d", "e"}, []string{"f", "g", "h"}), value: []string{"a", "b"}, want: true},
		{desc: "array with many arrays nok", data: makeComparisonData([]string{"d", "e"}, []string{"f", "g", "h", "i"}), value: []string{"a", "b", "c"}, want: false},
		{desc: "file size with number ok", data: makeComparisonData(5), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with number ok", data: makeComparisonData(4), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with number nok", data: makeComparisonData(3.99), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with many numbers ok", data: makeComparisonData(4, 4.99, 5), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with many numbers nok", data: makeComparisonData(3, 4.1), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with array ok", data: makeComparisonData([]string{"a", "b", "c", "d", "e"}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with array ok", data: makeComparisonData([]string{"a", "b", "c", "d"}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with array nok", data: makeComparisonData([]string{"a", "b", "c"}), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with many arrays ok", data: makeComparisonData([]string{"a", "b", "c", "d"}, []string{"a", "b", "c", "d", "e"}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with many arrays nok", data: makeComparisonData([]string{"a", "b"}, []string{"a", "b", "c", "d", "e"}), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2, "f": 3, "g": 4}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2, "f": 3}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with objects nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2}), value: []fsutil.File{largeFile}, want: false},
		{desc: "file size with many objects ok", data: makeComparisonData(map[string]any{"c": 0, "d": 1, "e": 2, "f": 3, "g": 4}, map[string]any{"e": 2, "f": 3, "g": 4, "h": 5}), value: []fsutil.File{largeFile}, want: true},
		{desc: "file size with many objects nok", data: makeComparisonData(map[string]any{"c": 0, "d": 1}, map[string]any{"e": 2, "f": 3, "g": 4, "h": 5, "i": 6}), value: []fsutil.File{largeFile}, want: false},
		{desc: "number with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: 3, want: true},
		{desc: "number with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: 4, want: true},
		{desc: "number with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: 4.01, want: false},
		{desc: "string with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: "abc", want: true},
		{desc: "string with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: "abcd", want: true},
		{desc: "string with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: "abcde", want: false},
		{desc: "array with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: []string{"a", "b", "c"}, want: true},
		{desc: "array with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: []string{"a", "b", "c", "d"}, want: true},
		{desc: "array with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: []string{"a", "b", "c", "d", "e"}, want: false},
		{desc: "object with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: map[string]any{"a": 0, "b": 1, "c": 2}, want: true},
		{desc: "object with file size ok", data: makeComparisonData([]fsutil.File{largeFile}), value: map[string]any{"a": 0, "b": 1, "c": 2, "d": 3}, want: true},
		{desc: "object with file size nok", data: makeComparisonData([]fsutil.File{largeFile}), value: map[string]any{"c": 0, "d": 1, "e": 2, "f": 3, "g": 4}, want: false},
		{desc: "empty array", data: makeComparisonData(), value: "abc", want: true},

		{desc: "cannot validate bool", data: makeComparisonData(false), value: true, want: true},
		{desc: "cannot validate time", data: makeComparisonData(time.Now()), value: time.Now(), want: true},
		{desc: "cannot validate nil", data: makeComparisonData(nil), value: nil, want: true},
		{desc: "cannot validate nil", data: makeComparisonData(nil), value: 1, want: true},
		{desc: "cannot validate nil", data: makeComparisonData(1), value: nil, want: true},

		{desc: "not found", data: map[string]any{"object": map[string]any{}}, value: "abc", want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			v := LowerThanEqual(path)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
				Data:  c.data,
			}))
		})
	}
}
