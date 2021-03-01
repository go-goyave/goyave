package validation

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v3/helper/filesystem"
)

const (
	logoPath       string = "resources/img/logo/goyave_16.png"
	mediumLogoPath string = "resources/img/logo/goyave_128.png"
	largeLogoPath  string = "resources/img/logo/goyave_512.png"
	configPath     string = "config/config.test.json"
	utf8BOMPath    string = "resources/test_file.txt"
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

func createTestFiles(files ...string) []filesystem.File {
	_, filename, _, _ := runtime.Caller(1)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, p := range files {
		fp := path.Dir(filename) + "/../" + p
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
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		panic(err)
	}
	return filesystem.ParseMultipartFiles(req, "file")
}

func createTestFileWithNoExtension() []filesystem.File {
	_, filename, _, _ := runtime.Caller(1)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	addFileToRequest(writer, path.Dir(filename)+"/../"+logoPath, "file", "noextension")
	err := writer.Close()
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "/test-route", body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		panic(err)
	}
	return filesystem.ParseMultipartFiles(req, "file")
}

func TestValidateFile(t *testing.T) {
	assert.True(t, validateFile("file", createTestFiles(logoPath), []string{}, map[string]interface{}{}))
	assert.True(t, validateFile("file", createTestFiles(logoPath, configPath), []string{}, map[string]interface{}{}))
	assert.False(t, validateFile("file", "test", []string{}, map[string]interface{}{}))
	assert.False(t, validateFile("file", 1, []string{}, map[string]interface{}{}))
	assert.False(t, validateFile("file", 1.2, []string{}, map[string]interface{}{}))
	assert.False(t, validateFile("file", true, []string{}, map[string]interface{}{}))
	assert.False(t, validateFile("file", []string{}, []string{}, map[string]interface{}{}))
}

func TestValidateMIME(t *testing.T) {
	assert.True(t, validateMIME("file", createTestFiles(logoPath), []string{"image/png"}, map[string]interface{}{}))
	assert.True(t, validateMIME("file", createTestFiles(logoPath), []string{"text/plain", "image/png"}, map[string]interface{}{}))
	assert.True(t, validateMIME("file", createTestFiles(utf8BOMPath), []string{"text/plain", "image/jpeg"}, map[string]interface{}{}))
	assert.True(t, validateMIME("file", createTestFiles(utf8BOMPath, logoPath), []string{"text/plain", "image/png"}, map[string]interface{}{}))
	assert.False(t, validateMIME("file", createTestFiles(utf8BOMPath, logoPath), []string{"text/plain"}, map[string]interface{}{}))
	assert.False(t, validateMIME("file", createTestFiles(logoPath), []string{"text/plain", "image/jpeg"}, map[string]interface{}{}))
	assert.False(t, validateMIME("file", "test", []string{"text/plain", "image/jpeg"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "mime"},
			},
		}
		field.check()
	})
}

func TestValidateImage(t *testing.T) {
	assert.True(t, validateImage("file", createTestFiles(logoPath), []string{}, map[string]interface{}{}))
	assert.False(t, validateImage("file", createTestFiles(configPath), []string{}, map[string]interface{}{}))
	assert.False(t, validateImage("file", createTestFiles(logoPath, configPath), []string{}, map[string]interface{}{}))
}

func TestValidateExtension(t *testing.T) {
	assert.True(t, validateExtension("file", createTestFiles(logoPath), []string{"png", "jpeg"}, map[string]interface{}{}))
	assert.True(t, validateExtension("file", createTestFiles(logoPath, configPath), []string{"png", "json"}, map[string]interface{}{}))
	assert.False(t, validateExtension("file", createTestFiles(logoPath), []string{"jpeg"}, map[string]interface{}{}))
	assert.False(t, validateExtension("file", createTestFiles(logoPath, configPath), []string{"png"}, map[string]interface{}{}))
	assert.False(t, validateExtension("file", createTestFileWithNoExtension(), []string{"png"}, map[string]interface{}{}))
	assert.False(t, validateExtension("file", "test", []string{"png"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "extension"},
			},
		}
		field.check()
	})
}

func TestValidateCount(t *testing.T) {
	assert.True(t, validateCount("file", createTestFiles(logoPath, configPath), []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateCount("file", createTestFiles(logoPath, configPath), []string{"3"}, map[string]interface{}{}))

	assert.False(t, validateCount("file", "test", []string{"3"}, map[string]interface{}{}))
	assert.Panics(t, func() { validateCount("file", true, []string{"test"}, map[string]interface{}{}) })

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "count"},
			},
		}
		field.check()
	})
}

func TestValidateCountMin(t *testing.T) {
	assert.True(t, validateCountMin("file", createTestFiles(logoPath, configPath), []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateCountMin("file", createTestFiles(logoPath, configPath), []string{"3"}, map[string]interface{}{}))

	assert.False(t, validateCountMin("file", "test", []string{"3"}, map[string]interface{}{}))
	assert.Panics(t, func() { validateCountMin("file", true, []string{"test"}, map[string]interface{}{}) })

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "count_min"},
			},
		}
		field.check()
	})
}

func TestValidateCountMax(t *testing.T) {
	assert.True(t, validateCountMax("file", createTestFiles(logoPath, configPath), []string{"2"}, map[string]interface{}{}))
	assert.False(t, validateCountMax("file", createTestFiles(logoPath, configPath), []string{"1"}, map[string]interface{}{}))

	assert.False(t, validateCountMax("file", "test", []string{"3"}, map[string]interface{}{}))
	assert.Panics(t, func() { validateCountMax("file", true, []string{"test"}, map[string]interface{}{}) })

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "count_max"},
			},
		}
		field.check()
	})
}

func TestValidateCountBetween(t *testing.T) {
	assert.True(t, validateCountBetween("file", createTestFiles(logoPath, configPath), []string{"1", "5"}, map[string]interface{}{}))
	assert.False(t, validateCountBetween("file", createTestFiles(logoPath, largeLogoPath, configPath), []string{"1", "2"}, map[string]interface{}{}))
	assert.False(t, validateCountBetween("file", createTestFiles(logoPath, configPath), []string{"3", "5"}, map[string]interface{}{}))

	assert.False(t, validateCountBetween("file", "test", []string{"3", "4"}, map[string]interface{}{}))
	assert.Panics(t, func() { validateCountBetween("file", true, []string{"test", "2"}, map[string]interface{}{}) })
	assert.Panics(t, func() { validateCountBetween("file", true, []string{"2", "test"}, map[string]interface{}{}) })
	assert.Panics(t, func() { validateCountBetween("file", true, []string{"test", "test"}, map[string]interface{}{}) })

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "count_between"},
			},
		}
		field.check()
	})

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "count_between", Params: []string{"2"}},
			},
		}
		field.check()
	})
}
