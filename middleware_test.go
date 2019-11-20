package goyave

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/System-Glitch/goyave/config"
	"github.com/System-Glitch/goyave/helpers/filesystem"
	"github.com/System-Glitch/goyave/lang"
	"github.com/System-Glitch/goyave/validation"
	"github.com/stretchr/testify/suite"
)

type MiddlewareTestSuite struct {
	suite.Suite
}

func (suite *MiddlewareTestSuite) SetupSuite() {
	config.LoadConfig()
	lang.LoadDefault()
}

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

func createTestFileRequest(route string, files ...string) *http.Request {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		log.Panicf("Runtime caller error")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, p := range files {
		fp := path.Dir(filename) + "/" + p
		addFileToRequest(writer, fp, "file", filepath.Base(fp))
	}
	field, err := writer.CreateFormField("field")
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(field, strings.NewReader("world"))
	if err != nil {
		panic(err)
	}

	err = writer.Close()
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", route, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func testMiddleware(middleware Middleware, rawRequest *http.Request, data map[string]interface{}, rules validation.RuleSet, handler func(*Response, *Request)) *http.Response {
	request := &Request{
		httpRequest: rawRequest,
		Data:        data,
		Rules:       rules,
		Params:      map[string]string{},
	}
	response := &Response{
		writer: httptest.NewRecorder(),
		empty:  true,
	}
	middleware(handler)(response, request)

	return response.writer.(*httptest.ResponseRecorder).Result()
}

func (suite *MiddlewareTestSuite) TestRecoveryMiddlewarePanicDebug() {
	config.Set("debug", true)
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	resp := testMiddleware(recoveryMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		panic(fmt.Errorf("error message"))
	})
	body, _ := ioutil.ReadAll(resp.Body)
	suite.Equal(500, resp.StatusCode)
	suite.Equal("{\"error\":\"error message\"}\n", string(body))
}

func (suite *MiddlewareTestSuite) TestRecoveryMiddlewarePanicNoDebug() {
	config.Set("debug", false)
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	resp := testMiddleware(recoveryMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		panic(fmt.Errorf("error message"))
	})
	body, _ := ioutil.ReadAll(resp.Body)
	suite.Equal(500, resp.StatusCode)
	suite.Equal("", string(body))
	config.Set("debug", true)
}

func (suite *MiddlewareTestSuite) TestRecoveryMiddlewareNoPanic() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	resp := testMiddleware(recoveryMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		response.String(200, "message")
	})
	body, _ := ioutil.ReadAll(resp.Body)
	suite.Equal(200, resp.StatusCode)
	suite.Equal("message", string(body))
}

func (suite *MiddlewareTestSuite) TestLanguageMiddleware() {
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	rawRequest.Header.Set("Accept-Language", "en-US")
	testMiddleware(languageMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		suite.Equal("en-US", r.Lang)
	})

	rawRequest = httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	testMiddleware(languageMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		suite.Equal("en-US", r.Lang)
	})
}

func (suite *MiddlewareTestSuite) TestParsePostRequestMiddleware() {
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		suite.Equal("hello world", r.Data["string"])
		suite.Equal("42", r.Data["number"])
	})
}

func (suite *MiddlewareTestSuite) TestParseGetRequestMiddleware() {
	rawRequest := httptest.NewRequest("GET", "/test-route?string=hello%20world&number=42", nil)
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		suite.Equal("hello world", r.Data["string"])
		suite.Equal("42", r.Data["number"])
	})

	rawRequest = httptest.NewRequest("GET", "/test-route?%9", nil)
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		suite.Nil(r.Data)
	})
}

func (suite *MiddlewareTestSuite) TestParseJsonRequestMiddleware() {
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("{\"string\":\"hello world\", \"number\":42, \"array\":[\"val1\",\"val2\"]}"))
	rawRequest.Header.Set("Content-Type", "application/json")
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		suite.Equal("hello world", r.Data["string"])
		suite.Equal(42.0, r.Data["number"])
		slice, ok := r.Data["array"].([]interface{})
		suite.True(ok)
		suite.Equal(2, len(slice))
		suite.Equal("val1", slice[0])
		suite.Equal("val2", slice[1])
	})

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("{\"string\":\"hello world\", \"number\":42, \"array\":[\"val1\",\"val2\"]")) // Missing closing braces
	rawRequest.Header.Set("Content-Type", "application/json")
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		suite.Nil(r.Data)
	})
}

func (suite *MiddlewareTestSuite) TestParseMultipartRequestMiddleware() {
	rawRequest := createTestFileRequest("/test-route?test=hello", "resources/img/logo/goyave_16.png")
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		suite.Equal(3, len(r.Data))
		suite.Equal("hello", r.Data["test"])
		suite.Equal("world", r.Data["field"])
		files, ok := r.Data["file"].([]filesystem.File)
		suite.True(ok)
		suite.Equal(1, len(files))
	})
}

func (suite *MiddlewareTestSuite) TestParseMultipartOverrideMiddleware() {
	rawRequest := createTestFileRequest("/test-route?field=hello", "resources/img/logo/goyave_16.png")
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		suite.Equal(2, len(r.Data))
		suite.Equal("world", r.Data["field"])
		files, ok := r.Data["file"].([]filesystem.File)
		suite.True(ok)
		suite.Equal(1, len(files))
	})
}

func (suite *MiddlewareTestSuite) TestParseMiddlewareWithArray() {
	rawRequest := httptest.NewRequest("GET", "/test-route?arr=hello&arr=world", nil)
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, func(response *Response, r *Request) {
		arr, ok := r.Data["arr"].([]string)
		suite.True(ok)
		if ok {
			suite.Equal(2, len(arr))
			suite.Equal("hello", arr[0])
			suite.Equal("world", arr[1])
		}
	})
}

func (suite *MiddlewareTestSuite) TestValidateMiddleware() {
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	data := map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}
	rules := validation.RuleSet{
		"string": {"required", "string"},
		"number": {"required", "numeric", "min:10"},
	}
	result := testMiddleware(validateRequestMiddleware, rawRequest, data, rules, func(response *Response, r *Request) {})
	suite.Equal(200, result.StatusCode)

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	data = map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}
	rules = validation.RuleSet{
		"string": {"required", "string"},
		"number": {"required", "numeric", "min:50"},
	}
	result = testMiddleware(validateRequestMiddleware, rawRequest, data, rules, func(response *Response, r *Request) {})
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(422, result.StatusCode)
	suite.Equal("{\"validationError\":{\"number\":[\"validation.rules.min.numeric\"]}}\n", string(body))

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	result = testMiddleware(validateRequestMiddleware, rawRequest, nil, rules, func(response *Response, r *Request) {})
	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(400, result.StatusCode)
	suite.Equal("{\"validationError\":{\"error\":[\"Malformed JSON\"]}}\n", string(body))
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}
