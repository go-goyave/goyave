package filesystem

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
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

	req, err := http.NewRequest("POST", "/test-route", body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.ParseMultipartForm(10 << 20)
	return ParseMultipartFiles(req, "file")
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
	mime, size := GetMIMEType(toAbsolutePath("resources/img/logo/goyave_16.png"))
	assert.Equal(t, "image/png", mime)
	assert.Equal(t, int64(716), size)

	mime, _ = GetMIMEType(toAbsolutePath("install.sh"))
	assert.Equal(t, "text/plain; charset=utf-8", mime)

	mime, _ = GetMIMEType(toAbsolutePath(".gitignore"))
	assert.Equal(t, "application/octet-stream", mime)

	mime, _ = GetMIMEType(toAbsolutePath("config/config.test.json"))
	assert.Equal(t, "application/json", mime)

	mime, _ = GetMIMEType(toAbsolutePath("docs_src/src/.vuepress/config.js"))
	assert.Equal(t, "application/javascript; charset=utf-8", mime)

	cssPath := toAbsolutePath("helper/filesystem/test.css")
	err := ioutil.WriteFile(cssPath, []byte("body{ margin:0; }"), 0644)
	if err != nil {
		panic(err)
	}
	mime, _ = GetMIMEType(cssPath)
	assert.Equal(t, "text/css", mime)
	Delete(cssPath)

	assert.Panics(t, func() {
		GetMIMEType(toAbsolutePath("doesn't exist"))
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
	actualName := file.Save(toAbsolutePath("."), "saved.png")
	actualPath := toAbsolutePath(actualName)
	assert.True(t, FileExists(actualPath))

	Delete(actualPath)
	assert.False(t, FileExists(actualPath))

	file = createTestFiles("resources/img/logo/goyave_16.png")[0]
	actualName = file.Save(toAbsolutePath("."), "saved")
	actualPath = toAbsolutePath(actualName)
	assert.Equal(t, -1, strings.Index(actualName, "."))
	assert.True(t, FileExists(actualPath))

	Delete(actualPath)
	assert.False(t, FileExists(actualPath))

	assert.Panics(t, func() {
		Delete(actualPath)
	})
}
