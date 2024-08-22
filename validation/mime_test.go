package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/util/fsutil"
)

func TestMIMEValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := MIME("application/json", "image/png")
		assert.NotNil(t, v)
		assert.Equal(t, "mime", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":values", "application/json, image/png"}, v.MessagePlaceholders(&Context{}))
		assert.Equal(t, []string{"application/json", "image/png"}, v.MIMETypes)
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: []fsutil.File{{MIMEType: "image/png"}, {MIMEType: "image/jpeg"}}, want: true},
		{value: []fsutil.File{{MIMEType: "image/png; setting=value"}, {MIMEType: "image/jpeg"}}, want: true},
		{value: []fsutil.File{{MIMEType: "text/csv"}, {MIMEType: "image/png"}}, want: false},
		{value: []fsutil.File{{MIMEType: "text/csv"}}, want: false},
		{value: fsutil.File{}, want: false},
		{value: "string", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: struct{}{}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := MIME("image/jpeg", "image/png")
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}

func TestImageValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Image()
		assert.NotNil(t, v)
		assert.Equal(t, "image", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":values", "image/jpeg, image/png, image/gif, image/bmp, image/svg+xml, image/webp"}, v.MessagePlaceholders(&Context{}))
		assert.Equal(t, []string{"image/jpeg", "image/png", "image/gif", "image/bmp", "image/svg+xml", "image/webp"}, v.MIMETypes)
	})

	cases := []struct {
		value any
		want  bool
	}{
		{value: []fsutil.File{{MIMEType: "image/png"}, {MIMEType: "image/jpeg"}, {MIMEType: "image/gif"}, {MIMEType: "image/bmp"}, {MIMEType: "image/svg+xml"}, {MIMEType: "image/webp"}}, want: true},
		{value: []fsutil.File{{MIMEType: "image/png; setting=value"}, {MIMEType: "image/jpeg"}}, want: true},
		{value: []fsutil.File{{MIMEType: "text/csv"}, {MIMEType: "image/png"}}, want: false},
		{value: []fsutil.File{{MIMEType: "text/csv"}}, want: false},
		{value: fsutil.File{}, want: false},
		{value: "string", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: struct{}{}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Validate_%v_%t", c.value, c.want), func(t *testing.T) {
			v := Image()
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}
