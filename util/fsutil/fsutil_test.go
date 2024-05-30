package fsutil

import (
	"bytes"
	"embed"
	"encoding/json"
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

	_ "embed"

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
	assert.Equal(t, "", GetFileExtension("test"))
}

func TestGetMIMEType(t *testing.T) {
	mime, size, err := GetMIMEType(&osfs.FS{}, toAbsolutePath("resources/img/logo/goyave_16.png"))
	assert.Equal(t, "image/png", mime)
	assert.Equal(t, int64(630), size)
	require.NoError(t, err)

	mime, _, err = GetMIMEType(&osfs.FS{}, toAbsolutePath("resources/test_script.sh"))
	require.NoError(t, err)
	assert.Equal(t, "text/plain; charset=utf-8", mime)

	mime, _, err = GetMIMEType(&osfs.FS{}, toAbsolutePath(".gitignore"))
	require.NoError(t, err)
	assert.Equal(t, "application/octet-stream", mime)

	mime, _, err = GetMIMEType(&osfs.FS{}, toAbsolutePath("config/config.test.json"))
	require.NoError(t, err)
	assert.Equal(t, "application/json", mime)

	mime, _, err = GetMIMEType(&osfs.FS{}, toAbsolutePath("resources/test_script.js"))
	require.NoError(t, err)
	assert.Equal(t, "text/javascript; charset=utf-8", mime)

	cssPath := toAbsolutePath("util/fsutil/test.css")
	err = os.WriteFile(cssPath, []byte("body{ margin:0; }"), 0644)
	require.NoError(t, err)
	mime, _, err = GetMIMEType(&osfs.FS{}, cssPath)
	assert.Equal(t, "text/css", mime)
	require.NoError(t, err)
	deleteFile(cssPath)

	_, _, err = GetMIMEType(&osfs.FS{}, toAbsolutePath("doesn't exist"))
	require.Error(t, err)

	t.Run("empty_file", func(t *testing.T) {
		filename := "empty_GetMIMEType.json"
		if err := os.WriteFile(filename, []byte{}, 0644); err != nil {
			panic(err)
		}

		t.Cleanup(func() {
			deleteFile(filename)
		})

		mime, size, err = GetMIMEType(&osfs.FS{}, filename)

		assert.Equal(t, "application/json", mime)
		assert.Equal(t, int64(0), size)
		require.NoError(t, err)
	})
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
		assert.Contains(t, err.Error(), "cannot unmarshal number into Go struct field testDTO.files of type []fsutil.File")
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
