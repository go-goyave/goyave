package fsutil

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"mime/multipart"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
	"goyave.dev/goyave/v5/util/typeutil"
)

func deleteFile(path string) {
	if err := os.Remove(path); err != nil {
		panic(err)
	}
}

func addFileToRequest(writer *multipart.Writer, path, name, fileName string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
	}()
	part, err := writer.CreateFormFile(name, fileName)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		panic(err)
	}
}

func createTestForm(files ...string) *multipart.Form {
	_, filename, _, _ := runtime.Caller(1)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, p := range files {
		fp := path.Dir(filename) + "/../../" + p
		addFileToRequest(writer, fp, "file", filepath.Base(fp))
	}
	err := writer.Close()
	if err != nil {
		panic(err)
	}

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(math.MaxInt64 - 1)
	if err != nil {
		panic(err)
	}
	return form
}

func createTestFiles(files ...string) []File {
	form := createTestForm(files...)
	f, err := ParseMultipartFiles(form.File["file"])
	if err != nil {
		panic(err)
	}
	return f
}

func toAbsolutePath(relativePath string) string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename) + "/../../" + relativePath
}

func TestGetFileExtension(t *testing.T) {
	assert.Equal(t, "png", GetFileExtension("test.png"))
	assert.Equal(t, "gz", GetFileExtension("test.tar.gz"))
	assert.Empty(t, GetFileExtension("test"))
}

func TestGetMIMEType(t *testing.T) {
	cssPath := toAbsolutePath("util/fsutil/test.css")
	err := os.WriteFile(cssPath, []byte("body{ margin:0; }"), 0644)
	require.NoError(t, err)
	t.Cleanup(func() {
		deleteFile(cssPath)
	})

	cases := []struct {
		wantSize *int64
		path     string
		wantMIME string
		wantErr  bool
	}{
		{
			path:     "resources/img/logo/goyave_16.png",
			wantMIME: "image/png",
			wantSize: new(int64(630)),
			wantErr:  false,
		},
		{
			path:     "resources/test_script.sh",
			wantMIME: "application/x-sh; charset=utf-8",
			wantErr:  false,
		},
		{
			path:     "resources/empty.txt",
			wantMIME: "text/plain",
			wantErr:  false,
		},
		{
			path:     "resources/test_file.txt",
			wantMIME: "text/plain; charset=utf-8",
			wantErr:  false,
		},
		{
			path:     ".gitignore",
			wantMIME: "application/octet-stream",
			wantErr:  false,
		},
		{
			path:     "config/config.test.json",
			wantMIME: "application/json",
			wantErr:  false,
		},
		{
			path:     "resources/test_script.js",
			wantMIME: "text/javascript; charset=utf-8",
			wantErr:  false,
		},
		{
			path:     "util/fsutil/test.css",
			wantMIME: "text/css",
			wantErr:  false,
		},
		{
			path:     "doesn't exist",
			wantMIME: "",
			wantSize: new(int64(0)),
			wantErr:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			mime, size, err := GetMIMEType(&osfs.FS{}, toAbsolutePath(c.path))
			assert.Equal(t, c.wantMIME, mime)
			if c.wantSize != nil {
				assert.Equal(t, *c.wantSize, size)
			}
			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	t.Run("empty_file", func(t *testing.T) {
		filename := "empty_GetMIMEType.json"
		if err := os.WriteFile(filename, []byte{}, 0644); err != nil {
			panic(err)
		}

		t.Cleanup(func() {
			deleteFile(filename)
		})

		mime, size, err := GetMIMEType(&osfs.FS{}, filename)

		assert.Equal(t, "application/json", mime)
		assert.Equal(t, int64(0), size)
		require.NoError(t, err)
	})
}

type testBuffer struct {
	buf []byte
	off int
}

func (b *testBuffer) Seek(_ int64, _ int) (int64, error) {
	b.off = 0
	return 0, nil
}

func (b *testBuffer) Read(p []byte) (int, error) {
	if b.off >= len(b.buf) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, b.buf[b.off:])
	b.off += n
	return n, nil
}

func TestDetectContentType(t *testing.T) {
	cases := []struct {
		fileName string
		wantMIME string
		buf      []byte
		wantErr  bool
	}{
		{
			fileName: "utf8.txt",
			buf:      append([]byte{0xef, 0xbb, 0xbf}, []byte("utf-8 with BOM content")...),
			wantMIME: "text/plain; charset=utf-8",
		},
		{
			fileName: "utf8.json",
			buf:      append([]byte{0xef, 0xbb, 0xbf}, []byte("utf-8 with BOM content")...),
			wantMIME: "application/json; charset=utf-8",
		},
		{
			fileName: "file.json",
			buf:      []byte("non-utf-8 content"),
			wantMIME: "application/json",
		},
		{
			fileName: "octet-stream",
			buf:      []byte{1, 2, 3},
			wantMIME: "application/octet-stream",
		},
		{
			fileName: "octet-stream.js",
			buf:      []byte{1, 2, 3},
			wantMIME: "text/javascript",
		},
		{
			fileName: "image.svg",
			buf:      []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?><svg xmlns=\"http://www.w3.org/2000/svg\"></svg>"),
			wantMIME: "image/svg+xml",
		},
		{
			fileName: "script.js",
			buf:      []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?><root></root>"),
			wantMIME: "text/xml; charset=utf-8",
		},
		{
			fileName: "eof",
			buf:      []byte{},
			wantErr:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.fileName, func(t *testing.T) {
			b := &testBuffer{
				buf: c.buf,
			}
			mime, err := DetectContentType(b, c.fileName)
			assert.Equal(t, c.wantMIME, mime)
			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, 0, b.off) // The reader has been reset
		})
	}
}

func TestDetectContentTypeByExtension(t *testing.T) {
	cases := []struct {
		desc        string
		want        string
		fileName    string
		contentType string
	}{
		{
			desc:        "unknown",
			want:        "unknown",
			contentType: "unknown",
			fileName:    "test.xyz",
		},
		{
			desc:        "empty_filename",
			want:        "unknown",
			contentType: "unknown",
			fileName:    "",
		},
		{
			desc:        "unknown_with_params",
			want:        "unknown; charset=utf-8",
			contentType: "unknown; charset=utf-8",
			fileName:    "test.xyz",
		},
		{
			desc:        "webp",
			want:        "image/webp",
			contentType: "application/octet-stream",
			fileName:    "picture.webp",
		},
		{
			desc:        "utf8_text",
			want:        "text/plain; charset=utf-8",
			contentType: "text/plain; charset=utf-8",
			fileName:    "test.txt",
		},
		{
			desc:        "utf8_css",
			want:        "text/css; charset=utf-8",
			contentType: "text/plain; charset=utf-8",
			fileName:    "test.css",
		},
		{
			desc:        "utf8_svg",
			want:        "image/svg+xml; charset=utf-8",
			contentType: "text/xml; charset=utf-8",
			fileName:    "image.svg",
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			assert.Equal(t, c.want, detectContentTypeByExtension(c.fileName, c.contentType))
		})
	}
}

func TestHasSVGSignature(t *testing.T) {
	cases := []struct {
		desc string
		buf  []byte
		want bool
	}{
		{
			desc: "direct_svg",
			buf:  []byte("<svg xmlns=\"http://www.w3.org/2000/svg\"></svg>"),
			want: true,
		},
		{
			desc: "svg_with_prolog",
			buf:  []byte("<?xml version=\"1.0\"?><svg></svg>"),
			want: true,
		},
		{
			desc: "svg_with_comment",
			buf:  []byte("<?xml version=\"1.0\"?><!--comment--><svg></svg>"),
			want: true,
		},
		{
			desc: "xml_not_svg",
			buf:  []byte("<?xml version=\"1.0\"?><root></root>"),
			want: false,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			assert.Equal(t, c.want, hasSVGSignature(c.buf))
		})
	}
}

func TestFileExists(t *testing.T) {
	assert.True(t, FileExists(&osfs.FS{}, toAbsolutePath("resources/img/logo/goyave_16.png")))
	assert.False(t, FileExists(&osfs.FS{}, toAbsolutePath("doesn't exist")))
}

func TestIsDirectory(t *testing.T) {
	assert.True(t, IsDirectory(&osfs.FS{}, toAbsolutePath("resources/img/logo")))
	assert.False(t, IsDirectory(&osfs.FS{}, toAbsolutePath("resources/img/logo/goyave_16.png")))
	assert.False(t, IsDirectory(&osfs.FS{}, toAbsolutePath("doesn't exist")))
}

func TestSave(t *testing.T) {
	fs := &osfs.FS{}
	file := createTestFiles("resources/img/logo/goyave_16.png")[0]
	actualName, err := file.Save(fs, toAbsolutePath("."), "saved.png")
	actualPath := toAbsolutePath(actualName)
	assert.True(t, FileExists(fs, actualPath))
	assert.NoError(t, err)

	deleteFile(actualPath)
	assert.False(t, FileExists(fs, actualPath))

	file = createTestFiles("resources/img/logo/goyave_16.png")[0]
	actualName, err = file.Save(fs, toAbsolutePath("."), "saved")
	actualPath = toAbsolutePath(actualName)
	assert.Equal(t, -1, strings.Index(actualName, "."))
	assert.True(t, FileExists(fs, actualPath))
	assert.NoError(t, err)

	deleteFile(actualPath)
	assert.False(t, FileExists(fs, actualPath))

	assert.Panics(t, func() {
		deleteFile(actualPath)
	})

	file = createTestFiles("resources/img/logo/goyave_16.png")[0]
	path := toAbsolutePath("./subdir")
	actualName, err = file.Save(fs, path, "saved")
	actualPath = toAbsolutePath("./subdir/" + actualName)
	assert.True(t, FileExists(fs, actualPath))
	assert.NoError(t, err)

	assert.NoError(t, os.RemoveAll(path))
	assert.False(t, FileExists(fs, actualPath))

	file = createTestFiles("resources/img/logo/goyave_16.png")[0]
	_, err = file.Save(fs, toAbsolutePath("./go.mod"), "saved")
	assert.Error(t, err)
}

func TestMarshalFile(t *testing.T) {
	type testDTO struct {
		Files []File `json:"files"`
	}

	t.Run("success", func(t *testing.T) {
		files := createTestFiles("resources/img/logo/goyave_16.png")
		data := map[string]any{"files": files}

		dto, err := typeutil.Convert[*testDTO](data)
		require.NoError(t, err)

		assert.Equal(t, files, dto.Files)
		for i, f := range files {
			assert.Same(t, f.Header, dto.Files[i].Header)
		}

		// Cache should be emptied.
		cacheMu.RLock()
		assert.Empty(t, marshalCache)
		cacheMu.RUnlock()
	})

	t.Run("unmarshal_err", func(t *testing.T) {
		data := map[string]any{"files": 123}

		_, err := typeutil.Convert[*testDTO](data)
		require.Error(t, err)
		// The error message format depends on the Go version: it names the
		// struct field either "testDTO.files" or ".files". Only assert on
		// the stable parts.
		assert.Contains(t, err.Error(), "cannot unmarshal number into Go struct field")
		assert.Contains(t, err.Error(), "[]fsutil.File")
	})

	t.Run("unmarshal_nocache", func(t *testing.T) {
		err := json.Unmarshal([]byte(`{"files": [{"Header":"uuid"}]}`), &testDTO{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot unmarshal fsutil.File: multipart header not found in cache")
	})
}

func TestOpenFileError(t *testing.T) {
	dir := "./forbidden_directory"
	assert.NoError(t, os.Mkdir(dir, 0500))
	defer func() {
		assert.NoError(t, os.RemoveAll(dir))
	}()
	file := createTestFiles("resources/img/logo/goyave_16.png")[0]
	filename, err := file.Save(&osfs.FS{}, dir, "saved.png")
	assert.Error(t, err)
	assert.NotEmpty(t, filename)
}

func TestParseMultipartFiles(t *testing.T) {
	t.Run("png", func(t *testing.T) {
		form := createTestForm("resources/img/logo/goyave_16.png")
		files, err := ParseMultipartFiles(form.File["file"])

		expected := []File{
			{
				Header:   form.File["file"][0],
				MIMEType: "image/png",
			},
		}
		assert.Equal(t, expected, files)
		assert.NoError(t, err)
	})

	t.Run("svg", func(t *testing.T) {
		svgPath := toAbsolutePath("util/fsutil/test.svg")
		err := os.WriteFile(svgPath, []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?><svg xmlns=\"http://www.w3.org/2000/svg\"></svg>"), 0644)
		require.NoError(t, err)
		t.Cleanup(func() {
			deleteFile(svgPath)
		})

		form := createTestForm("util/fsutil/test.svg")
		files, err := ParseMultipartFiles(form.File["file"])

		expected := []File{
			{
				Header:   form.File["file"][0],
				MIMEType: "image/svg+xml",
			},
		}
		assert.Equal(t, expected, files)
		assert.NoError(t, err)
	})

	t.Run("xml_content_with_js_extension", func(t *testing.T) {
		filePath := toAbsolutePath("util/fsutil/test.js")
		err := os.WriteFile(filePath, []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?><root></root>"), 0644)
		require.NoError(t, err)
		t.Cleanup(func() {
			deleteFile(filePath)
		})

		form := createTestForm("util/fsutil/test.js")
		files, err := ParseMultipartFiles(form.File["file"])

		require.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, "text/xml; charset=utf-8", files[0].MIMEType)
	})

	t.Run("declared_content_type_separate_from_detected_mime", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		headers := textproto.MIMEHeader{}
		headers.Set("Content-Disposition", `form-data; name="file"; filename="custom.js"`)
		headers.Set("Content-Type", "application/javascript")

		part, err := writer.CreatePart(headers)
		require.NoError(t, err)
		_, err = part.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?><root></root>"))
		require.NoError(t, err)
		require.NoError(t, writer.Close())

		reader := multipart.NewReader(body, writer.Boundary())
		form, err := reader.ReadForm(math.MaxInt64 - 1)
		require.NoError(t, err)

		files, err := ParseMultipartFiles(form.File["file"])
		require.NoError(t, err)
		require.Len(t, files, 1)

		assert.Equal(t, "application/javascript", files[0].Header.Header.Get("Content-Type"))
		assert.Equal(t, "text/xml; charset=utf-8", files[0].MIMEType)
	})

	t.Run("empty_file", func(t *testing.T) {
		headers := []*multipart.FileHeader{
			{
				Filename: "empty_ParseMultipartFiles.json",
				Size:     0,
				Header:   textproto.MIMEHeader{},
			},
		}
		files, err := ParseMultipartFiles(headers)

		expected := []File{
			{
				Header:   headers[0],
				MIMEType: "application/octet-stream",
			},
		}
		assert.Equal(t, expected, files)
		assert.NoError(t, err)
	})
}

//go:embed osfs
var resources embed.FS

type testStatFS struct {
	embed.FS
}

type mockFileInfo struct{}

func (fs *mockFileInfo) Name() string       { return "" }
func (fs *mockFileInfo) Size() int64        { return 0 }
func (fs *mockFileInfo) Mode() fs.FileMode  { return 0 }
func (fs *mockFileInfo) ModTime() time.Time { return time.Now() }
func (fs *mockFileInfo) Sys() any           { return nil }
func (fs *mockFileInfo) IsDir() bool        { return false }

func (t testStatFS) Stat(_ string) (fileinfo fs.FileInfo, err error) {
	return &mockFileInfo{}, nil
}

type mockFile struct {
	name string
}

func (f *mockFile) Stat() (fs.FileInfo, error) { return nil, nil }
func (f *mockFile) Read(_ []byte) (int, error) { return 0, nil }
func (f *mockFile) Close() error               { return nil }

type mockDirEntry struct{}

func (f *mockDirEntry) Name() string               { return "" }
func (f *mockDirEntry) IsDir() bool                { return false }
func (f *mockDirEntry) Type() fs.FileMode          { return 0 }
func (f *mockDirEntry) Info() (fs.FileInfo, error) { return &mockFileInfo{}, nil }

type mockFS struct{}

func (e mockFS) Open(name string) (fs.File, error) {
	return &mockFile{
		name: name,
	}, nil
}

func (e mockFS) ReadDir(_ string) ([]fs.DirEntry, error) {
	return []fs.DirEntry{&mockDirEntry{}}, nil
}

func TestEmbed(t *testing.T) {
	e := NewEmbed(resources)

	stat, err := e.Stat("osfs/osfs.go")
	require.NoError(t, err)
	assert.False(t, stat.IsDir())
	assert.Equal(t, "osfs.go", stat.Name())

	stat, err = e.Stat("notadir/osfs.go")
	assert.Nil(t, stat)
	if assert.Error(t, err) {
		e, ok := err.(*errors.Error)
		if assert.True(t, ok) {
			var fsErr *fs.PathError
			if assert.ErrorAs(t, e, &fsErr) {
				assert.Equal(t, "open", fsErr.Op)
				assert.Equal(t, "notadir/osfs.go", fsErr.Path)
			}
		}
	}

	// Make it so the underlying FS implements
	e.FS = testStatFS{resources}
	stat, err = e.Stat("osfs/osfs.go")
	require.NoError(t, err)
	_, ok := stat.(*mockFileInfo)
	assert.True(t, ok)

	t.Run("Open", func(t *testing.T) {
		e := NewEmbed(&mockFS{})

		f, err := e.Open("")
		require.NoError(t, err)
		_, ok := f.(*mockFile)
		assert.True(t, ok)
	})
	t.Run("ReadDir", func(t *testing.T) {
		e := NewEmbed(&mockFS{})

		f, err := e.ReadDir("")
		require.NoError(t, err)
		require.Len(t, f, 1)
		_, ok := f[0].(*mockDirEntry)
		assert.True(t, ok)
	})
}

func TestEmbedSub(t *testing.T) {
	t.Run("err", func(t *testing.T) {
		e := NewEmbed(resources)
		sub, err := e.Sub("..")
		assert.Equal(t, Embed{}, sub)
		assert.Error(t, err)
	})

	t.Run("Valid", func(t *testing.T) {
		e := NewEmbed(resources)
		sub, err := e.Sub("osfs.go") // It is valid to do this
		assert.NotNil(t, sub.FS)
		assert.NoError(t, err)
	})
}

func TestAddExtensionType(t *testing.T) {
	cases := []struct {
		wantErr error
		desc    string
		ext     string
		mime    string
	}{
		{
			desc: "OK",
			ext:  ".xyz",
			mime: "text/custom",
		},
		{
			desc:    "no_dot",
			ext:     "xyz",
			mime:    "text/custom",
			wantErr: fmt.Errorf("fsutil: extension \"xyz\" missing leading dot"),
		},
		{
			desc:    "no_dot",
			ext:     ".abc",
			mime:    "text/custom; charset=utf-8",
			wantErr: fmt.Errorf("fsutil: MIME type \"text/custom; charset=utf-8\" contains a parameter"),
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			err := AddExtensionType(c.ext, c.mime)
			if c.wantErr == nil {
				assert.Contains(t, contentTypeByExtension, c.ext)
			} else {
				assert.ErrorContains(t, err, c.wantErr.Error())
				assert.NotContains(t, contentTypeByExtension, c.ext)
			}
		})
	}
}
