package goyave

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/System-Glitch/goyave/v2/cors"
	"github.com/System-Glitch/goyave/v2/helper/filesystem"
	"github.com/System-Glitch/goyave/v2/lang"
	"github.com/System-Glitch/goyave/v2/validation"
	"github.com/stretchr/testify/suite"
)

type MiddlewareTestSuite struct {
	suite.Suite
}

func (suite *MiddlewareTestSuite) SetupSuite() {
	config.Load()
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
	_, filename, _, _ := runtime.Caller(1)

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

func testMiddleware(middleware Middleware, rawRequest *http.Request, data map[string]interface{}, rules validation.RuleSet, corsOptions *cors.Options, handler func(*Response, *Request)) *http.Response {
	request := &Request{
		httpRequest: rawRequest,
		corsOptions: corsOptions,
		Data:        data,
		Rules:       rules.AsRules(),
		Lang:        "en-US",
		Params:      map[string]string{},
	}
	response := newResponse(httptest.NewRecorder(), nil)
	middleware(handler)(response, request)

	return response.responseWriter.(*httptest.ResponseRecorder).Result()
}

func (suite *MiddlewareTestSuite) TestRecoveryMiddlewarePanic() {
	response := newResponse(httptest.NewRecorder(), nil)
	err := fmt.Errorf("error message")
	recoveryMiddleware(func(response *Response, r *Request) {
		panic(err)
	})(response, &Request{})
	suite.Equal(err, response.GetError())
	suite.Equal(500, response.status)
}

func (suite *MiddlewareTestSuite) TestRecoveryMiddlewareNoPanic() {
	response := newResponse(httptest.NewRecorder(), nil)
	recoveryMiddleware(func(response *Response, r *Request) {
		response.String(200, "message")
	})(response, &Request{})

	resp := response.responseWriter.(*httptest.ResponseRecorder).Result()
	suite.Nil(response.GetError())
	suite.Equal(200, response.status)
	suite.Equal(200, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	suite.Equal("message", string(body))
}

func (suite *MiddlewareTestSuite) TestRecoveryMiddlewareNilPanic() {
	response := newResponse(httptest.NewRecorder(), nil)
	recoveryMiddleware(func(response *Response, r *Request) {
		panic(nil)
	})(response, &Request{})
	suite.Nil(response.GetError())
	suite.Equal(500, response.status)
}

func (suite *MiddlewareTestSuite) TestLanguageMiddleware() {
	executed := false
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	rawRequest.Header.Set("Accept-Language", "en-US")
	testMiddleware(languageMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal("en-US", r.Lang)
		executed = true
	})
	suite.True(executed)

	rawRequest = httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	testMiddleware(languageMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal("en-US", r.Lang)
		executed = true
	})
	suite.True(executed)
}

func (suite *MiddlewareTestSuite) TestParsePostRequestMiddleware() {
	executed := false
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal("hello world", r.Data["string"])
		suite.Equal("42", r.Data["number"])
		executed = true
	})
	suite.True(executed)
}

func (suite *MiddlewareTestSuite) TestParseGetRequestMiddleware() {
	executed := false
	rawRequest := httptest.NewRequest("GET", "/test-route?string=hello%20world&number=42", nil)
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal("hello world", r.Data["string"])
		suite.Equal("42", r.Data["number"])
		executed = true
	})
	suite.True(executed)

	executed = false
	rawRequest = httptest.NewRequest("GET", "/test-route?%9", nil)
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Nil(r.Data)
		executed = true
	})
	suite.True(executed)
}

func (suite *MiddlewareTestSuite) TestParseJsonRequestMiddleware() {
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("{\"string\":\"hello world\", \"number\":42, \"array\":[\"val1\",\"val2\"]}"))
	rawRequest.Header.Set("Content-Type", "application/json")
	executed := false
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal("hello world", r.Data["string"])
		suite.Equal(42.0, r.Data["number"])
		slice, ok := r.Data["array"].([]interface{})
		suite.True(ok)
		suite.Equal(2, len(slice))
		suite.Equal("val1", slice[0])
		suite.Equal("val2", slice[1])
		executed = true
	})
	suite.True(executed)

	executed = false
	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("{\"string\":\"hello world\", \"number\":42, \"array\":[\"val1\",\"val2\"]")) // Missing closing braces
	rawRequest.Header.Set("Content-Type", "application/json")
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Nil(r.Data)
		executed = true
	})
	suite.True(executed)

	// Test with query parameters
	executed = false
	rawRequest = httptest.NewRequest("POST", "/test-route?query=param", strings.NewReader("{\"string\":\"hello world\", \"number\":42, \"array\":[\"val1\",\"val2\"]}"))
	rawRequest.Header.Set("Content-Type", "application/json")
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.NotNil(r.Data)
		suite.Equal("param", r.Data["query"])
		executed = true
	})
	suite.True(executed)

	// Test with charset (#101)
	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("{\"string\":\"hello world\", \"number\":42, \"array\":[\"val1\",\"val2\"]}"))
	rawRequest.Header.Set("Content-Type", "application/json; charset=utf-8")
	executed = false
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal("hello world", r.Data["string"])
		suite.Equal(42.0, r.Data["number"])
		slice, ok := r.Data["array"].([]interface{})
		suite.True(ok)
		suite.Equal(2, len(slice))
		suite.Equal("val1", slice[0])
		suite.Equal("val2", slice[1])
		executed = true
	})
	suite.True(executed)

}

func (suite *MiddlewareTestSuite) TestParseMultipartRequestMiddleware() {
	executed := false
	rawRequest := createTestFileRequest("/test-route?test=hello", "resources/img/logo/goyave_16.png")
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal(3, len(r.Data))
		suite.Equal("hello", r.Data["test"])
		suite.Equal("world", r.Data["field"])
		files, ok := r.Data["file"].([]filesystem.File)
		suite.True(ok)
		suite.Equal(1, len(files))
		executed = true
	})
	suite.True(executed)

	// Test payload too large
	prev := config.Get("maxUploadSize")
	config.Set("maxUploadSize", -10.0)
	rawRequest = createTestFileRequest("/test-route?test=hello", "resources/img/logo/goyave_16.png")

	request := createTestRequest(rawRequest)
	response := newResponse(httptest.NewRecorder(), nil)
	parseRequestMiddleware(nil)(response, request)
	suite.Equal(http.StatusRequestEntityTooLarge, response.GetStatus())
	config.Set("maxUploadSize", prev)
}

func (suite *MiddlewareTestSuite) TestParseMultipartOverrideMiddleware() {
	executed := false
	rawRequest := createTestFileRequest("/test-route?field=hello", "resources/img/logo/goyave_16.png")
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal(2, len(r.Data))
		suite.Equal("world", r.Data["field"])
		files, ok := r.Data["file"].([]filesystem.File)
		suite.True(ok)
		suite.Equal(1, len(files))
		executed = true
	})
	suite.True(executed)
}

func (suite *MiddlewareTestSuite) TestParseMiddlewareWithArray() {
	executed := false
	rawRequest := httptest.NewRequest("GET", "/test-route?arr=hello&arr=world", nil)
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		arr, ok := r.Data["arr"].([]string)
		suite.True(ok)
		if ok {
			suite.Equal(2, len(arr))
			suite.Equal("hello", arr[0])
			suite.Equal("world", arr[1])
		}
		executed = true
	})
	suite.True(executed)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	field, err := writer.CreateFormField("field")
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(field, strings.NewReader("hello"))
	if err != nil {
		panic(err)
	}

	field, err = writer.CreateFormField("field")
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

	executed = false
	rawRequest, err = http.NewRequest("POST", "/test-route", body)
	if err != nil {
		panic(err)
	}
	rawRequest.Header.Set("Content-Type", writer.FormDataContentType())
	testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal(1, len(r.Data))
		arr, ok := r.Data["field"].([]string)
		suite.True(ok)
		if ok {
			suite.Equal(2, len(arr))
			suite.Equal("hello", arr[0])
			suite.Equal("world", arr[1])
		}
		executed = true
	})
	suite.True(executed)
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
	result := testMiddleware(validateRequestMiddleware, rawRequest, data, rules, nil, func(response *Response, r *Request) {})
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
	result = testMiddleware(validateRequestMiddleware, rawRequest, data, rules, nil, func(response *Response, r *Request) {})
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(422, result.StatusCode)
	suite.Equal("{\"validationError\":{\"number\":[\"The number must be at least 50.\"]}}\n", string(body))

	rawRequest = httptest.NewRequest("POST", "/test-route", nil)
	rawRequest.Header.Set("Content-Type", "application/json")
	result = testMiddleware(validateRequestMiddleware, rawRequest, nil, rules, nil, func(response *Response, r *Request) {})
	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(400, result.StatusCode)
	suite.Equal("{\"validationError\":{\"error\":[\"Malformed JSON\"]}}\n", string(body))
}

func (suite *MiddlewareTestSuite) TestCORSMiddleware() {
	// No CORS options
	rawRequest := httptest.NewRequest("GET", "/test-route", nil)
	result := testMiddleware(corsMiddleware, rawRequest, nil, nil, nil, func(response *Response, r *Request) {})
	suite.Equal(200, result.StatusCode)

	// Preflight
	options := cors.Default()
	rawRequest = httptest.NewRequest("OPTIONS", "/test-route", nil)
	rawRequest.Header.Set("Origin", "https://google.com")
	rawRequest.Header.Set("Access-Control-Request-Method", "GET")
	result = testMiddleware(corsMiddleware, rawRequest, nil, nil, options, func(response *Response, r *Request) {
		response.String(200, "Hi!")
	})
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(204, result.StatusCode)
	suite.Empty(body)

	// Preflight passthrough
	options = cors.Default()
	options.OptionsPassthrough = true
	result = testMiddleware(corsMiddleware, rawRequest, nil, nil, options, func(response *Response, r *Request) {
		response.String(200, "Passthrough")
	})
	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(200, result.StatusCode)
	suite.Equal("Passthrough", string(body))

	// Preflight without Access-Control-Request-Method
	rawRequest = httptest.NewRequest("OPTIONS", "/test-route", nil)
	result = testMiddleware(corsMiddleware, rawRequest, nil, nil, options, func(response *Response, r *Request) {
		response.String(200, "Hi!")
	})
	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal(200, result.StatusCode)
	suite.Equal("Hi!", string(body))

	// Actual request
	options = cors.Default()
	options.AllowedOrigins = []string{"https://google.com", "https://images.google.com"}
	rawRequest = httptest.NewRequest("GET", "/test-route", nil)
	rawRequest.Header.Set("Origin", "https://images.google.com")
	result = testMiddleware(corsMiddleware, rawRequest, nil, nil, options, func(response *Response, r *Request) {
		response.String(200, "Hi!")
	})
	body, err = ioutil.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	suite.Equal("Hi!", string(body))
	suite.Equal("https://images.google.com", result.Header.Get("Access-Control-Allow-Origin"))
	suite.Equal("Origin", result.Header.Get("Vary"))
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}
