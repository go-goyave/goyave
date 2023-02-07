package fsutil

import (
	"bytes"
	"io"
	"math"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func addFileToRequest(writer *multipart.Writer, path, name, fileName string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	part, err := writer.CreateFormFile(name, fileName)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		panic(err)
	}
}

func createTestFiles(files ...string) []File {
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
	mime, size, err := GetMIMEType(toAbsolutePath("resources/img/logo/goyave_16.png"))
	assert.Equal(t, "image/png", mime)
	assert.Equal(t, int64(716), size)
	assert.NoError(t, err)

	mime, _, err = GetMIMEType(toAbsolutePath("resources/test_script.sh"))
	assert.NoError(t, err)
	assert.Equal(t, "text/plain; charset=utf-8", mime)

	mime, _, err = GetMIMEType(toAbsolutePath(".gitignore"))
	assert.NoError(t, err)
	assert.Equal(t, "application/octet-stream", mime)

	mime, _, err = GetMIMEType(toAbsolutePath("config/config.test.json"))
	assert.NoError(t, err)
	assert.Equal(t, "application/json", mime)

	mime, _, err = GetMIMEType(toAbsolutePath("resources/test_script.js"))
	assert.NoError(t, err)
	assert.Equal(t, "text/javascript; charset=utf-8", mime)

	cssPath := toAbsolutePath("util/fsutil/test.css")
	err = os.WriteFile(cssPath, []byte("body{ margin:0; }"), 0644)
	if err != nil {
		panic(err)
	}
	mime, _, err = GetMIMEType(cssPath)
	assert.Equal(t, "text/css", mime)
	assert.NoError(t, err)
	Delete(cssPath)

	_, _, err = GetMIMEType(toAbsolutePath("doesn't exist"))
	assert.Error(t, err)

	t.Run("empty_file", func(t *testing.T) {
		filename := "empty_GetMIMEType.json"
		if err := os.WriteFile(filename, []byte{}, 0644); err != nil {
			panic(err)
		}

		t.Cleanup(func() {
			Delete(filename)
		})

		mime, size, err = GetMIMEType(filename)

		assert.Equal(t, "application/json", mime)
		assert.Equal(t, int64(0), size)
		assert.NoError(t, err)
	})
}

func TestFileExists(t *testing.T) {
	assert.True(t, FileExists(toAbsolutePath("resources/img/logo/goyave_16.png")))
	assert.False(t, FileExists(toAbsolutePath("doesn't exist")))
}

func TestIsDirectory(t *testing.T) {
	assert.True(t, IsDirectory(toAbsolutePath("resources/img/logo")))
	assert.False(t, IsDirectory(toAbsolutePath("resources/img/logo/goyave_16.png")))
	assert.False(t, IsDirectory(toAbsolutePath("doesn't exist")))
}

func TestSaveDelete(t *testing.T) {
	file := createTestFiles("resources/img/logo/goyave_16.png")[0]
	actualName, err := file.Save(toAbsolutePath("."), "saved.png")
	actualPath := toAbsolutePath(actualName)
	assert.True(t, FileExists(actualPath))
	assert.NoError(t, err)

	Delete(actualPath)
	assert.False(t, FileExists(actualPath))

	file = createTestFiles("resources/img/logo/goyave_16.png")[0]
	actualName, err = file.Save(toAbsolutePath("."), "saved")
	actualPath = toAbsolutePath(actualName)
	assert.Equal(t, -1, strings.Index(actualName, "."))
	assert.True(t, FileExists(actualPath))
	assert.NoError(t, err)

	Delete(actualPath)
	assert.False(t, FileExists(actualPath))

	assert.Panics(t, func() {
		Delete(actualPath)
	})

	file = createTestFiles("resources/img/logo/goyave_16.png")[0]
	path := toAbsolutePath("./subdir")
	actualName, err = file.Save(path, "saved")
	actualPath = toAbsolutePath("./subdir/" + actualName)
	assert.True(t, FileExists(actualPath))
	assert.NoError(t, err)

	os.RemoveAll(path)
	assert.False(t, FileExists(actualPath))

	file = createTestFiles("resources/img/logo/goyave_16.png")[0]
	_, err = file.Save(toAbsolutePath("./go.mod"), "saved")
	assert.Error(t, err)
}

func TestOpenFileError(t *testing.T) {
	dir := "./forbidden_directory"
	assert.Panics(t, func() {
		os.Mkdir(dir, 0500)
		defer os.RemoveAll(dir)
		file := createTestFiles("resources/img/logo/goyave_16.png")[0]
		file.Save(dir, "saved.png")
	})
}
