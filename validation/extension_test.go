package validation

import (
	"fmt"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/util/fsutil"
)

func TestExtensionValidator(t *testing.T) {
	t.Run("Constructor", func(t *testing.T) {
		v := Extension("pdf", "png", "jpeg")
		assert.NotNil(t, v)
		assert.Equal(t, "extension", v.Name())
		assert.False(t, v.IsType())
		assert.False(t, v.IsTypeDependent())
		assert.Equal(t, []string{":values", "pdf, png, jpeg"}, v.MessagePlaceholders(&Context{}))
	})

	cases := []struct {
		value   any
		allowed []string
		want    bool
	}{
		{value: makeExtTestFiles("test.txt"), allowed: []string{"docx", "txt"}, want: true},
		{value: makeExtTestFiles("test.txt", "doc.docx"), allowed: []string{"docx", "txt"}, want: true},
		{value: makeExtTestFiles("test.txt", "doc.docx", "doc.pdf"), allowed: []string{"docx", "txt"}, want: false},
		{value: makeExtTestFiles("test.pdf"), allowed: []string{"docx", "txt"}, want: false},
		{value: makeExtTestFiles("archive.tar.gz"), allowed: []string{"tar.gz", "zip"}, want: true},
		{value: makeExtTestFiles("archive.atar.gz"), allowed: []string{"tar.gz", "zip"}, want: false},
		{value: makeExtTestFiles("noext"), allowed: []string{"tar.gz", "zip", "noext"}, want: false},
		{value: makeExtTestFiles("noext."), allowed: []string{"tar.gz", "zip", "noext"}, want: false},
		{value: "string", want: false},
		{value: "", want: false},
		{value: 'a', want: false},
		{value: 2, want: false},
		{value: 2.5, want: false},
		{value: []string{"string"}, want: false},
		{value: map[string]any{"a": 1}, want: false},
		{value: true, want: false},
		{value: nil, want: false},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("Validate_%s_%t", extTestFilesNames(c.value), c.want), func(t *testing.T) {
			v := Extension(c.allowed...)
			assert.Equal(t, c.want, v.Validate(&Context{
				Value: c.value,
			}))
		})
	}
}

func makeExtTestFiles(filename ...string) []fsutil.File {
	return lo.Map(filename, func(f string, _ int) fsutil.File {
		return fsutil.File{
			Header: &multipart.FileHeader{
				Filename: f,
			},
		}
	})
}

func extTestFilesNames(value any) string {
	if files, ok := value.([]fsutil.File); ok {
		return "[" + strings.Join(lo.Map(files, func(f fsutil.File, _ int) string { return f.Header.Filename }), "_") + "]"
	}
	return fmt.Sprintf("%v", value)

}
