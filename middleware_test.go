package goyave

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/cors"
	"goyave.dev/goyave/v4/lang"
	"goyave.dev/goyave/v4/util/fsutil"
	"goyave.dev/goyave/v4/validation"
)

type MiddlewareTestSuite struct {
	TestSuite
}

func (suite *MiddlewareTestSuite) SetupSuite() {
	lang.LoadDefault()
	maxPayloadSize = int64(config.GetFloat("server.maxUploadSize") * 1024 * 1024)
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
	suite.Empty(response.GetStacktrace())
	suite.Equal(500, response.status)

	prev := config.GetBool("app.debug")
	config.Set("app.debug", true)
	defer config.Set("app.debug", prev)

	response = newResponse(httptest.NewRecorder(), nil)
	err = fmt.Errorf("error message")
	recoveryMiddleware(func(response *Response, r *Request) {
		panic(err)
	})(response, &Request{})
	suite.Equal(err, response.GetError())
	suite.NotEmpty(response.GetStacktrace())
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

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
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
	defaultLanguage = config.GetString("app.defaultLanguage")
	executed := false
	rawRequest := httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	rawRequest.Header.Set("Accept-Language", "en-US")
	res := testMiddleware(languageMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal("en-US", r.Lang)
		executed = true
	})
	res.Body.Close()
	suite.True(executed)

	rawRequest = httptest.NewRequest("GET", "/test-route", strings.NewReader("body"))
	res = testMiddleware(languageMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal("en-US", r.Lang)
		executed = true
	})
	res.Body.Close()

	suite.True(executed)
}

func (suite *MiddlewareTestSuite) TestParsePostRequestMiddleware() {
	executed := false
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	res := testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal("hello world", r.Data["string"])
		suite.Equal("42", r.Data["number"])
		executed = true
	})
	suite.True(executed)
	res.Body.Close()
}

func (suite *MiddlewareTestSuite) TestParseGetRequestMiddleware() {
	executed := false
	rawRequest := httptest.NewRequest("GET", "/test-route?string=hello%20world&number=42", nil)
	res := testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal("hello world", r.Data["string"])
		suite.Equal("42", r.Data["number"])
		executed = true
	})
	suite.True(executed)
	res.Body.Close()

	executed = false
	rawRequest = httptest.NewRequest("GET", "/test-route?%9", nil)
	res = testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Nil(r.Data)
		executed = true
	})
	suite.True(executed)
	res.Body.Close()
}

func (suite *MiddlewareTestSuite) TestParseJsonRequestMiddleware() {
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("{\"string\":\"hello world\", \"number\":42, \"array\":[\"val1\",\"val2\"]}"))
	rawRequest.Header.Set("Content-Type", "application/json")
	executed := false
	res := testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
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
	res.Body.Close()

	executed = false
	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("{\"string\":\"hello world\", \"number\":42, \"array\":[\"val1\",\"val2\"]")) // Missing closing braces
	rawRequest.Header.Set("Content-Type", "application/json")
	res = testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Nil(r.Data)
		executed = true
	})
	suite.True(executed)
	res.Body.Close()

	// Test with query parameters
	executed = false
	rawRequest = httptest.NewRequest("POST", "/test-route?query=param", strings.NewReader("{\"string\":\"hello world\", \"number\":42, \"array\":[\"val1\",\"val2\"]}"))
	rawRequest.Header.Set("Content-Type", "application/json")
	res = testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.NotNil(r.Data)
		suite.Equal("param", r.Data["query"])
		executed = true
	})
	suite.True(executed)
	res.Body.Close()

	executed = false
	rawRequest = httptest.NewRequest("POST", "/test-route?%9", strings.NewReader("{\"string\":\"hello world\", \"number\":42, \"array\":[\"val1\",\"val2\"]}")) // Invalid query param
	rawRequest.Header.Set("Content-Type", "application/json")
	res = testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Nil(r.Data)
		executed = true
	})
	suite.True(executed)
	res.Body.Close()

	// Test with charset (#101)
	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("{\"string\":\"hello world\", \"number\":42, \"array\":[\"val1\",\"val2\"]}"))
	rawRequest.Header.Set("Content-Type", "application/json; charset=utf-8")
	executed = false
	res = testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal("hello world", r.Data["string"])
		suite.Equal(42.0, r.Data["number"])
		slice, ok := r.Data["array"].([]interface{})
		suite.True(ok)
		suite.Equal(2, len(slice))
		suite.Equal("val1", slice[0])
		suite.Equal("val2", slice[1])
		executed = true
	})
	res.Body.Close()
	suite.True(executed)

}

func (suite *MiddlewareTestSuite) TestParseMultipartRequestMiddleware() {
	executed := false
	rawRequest := createTestFileRequest("/test-route?test=hello", "resources/img/logo/goyave_16.png")
	res := testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal(3, len(r.Data))
		suite.Equal("hello", r.Data["test"])
		suite.Equal("world", r.Data["field"])
		files, ok := r.Data["file"].([]fsutil.File)
		suite.True(ok)
		suite.Equal(1, len(files))
		executed = true
	})
	suite.True(executed)
	res.Body.Close()

	// Test payload too large
	prev := config.Get("server.maxUploadSize")
	config.Set("server.maxUploadSize", -10.0)
	maxPayloadSize = int64(config.GetFloat("server.maxUploadSize") * 1024 * 1024)
	rawRequest = createTestFileRequest("/test-route?test=hello", "resources/img/logo/goyave_16.png")

	request := createTestRequest(rawRequest)
	response := newResponse(httptest.NewRecorder(), nil)
	parseRequestMiddleware(nil)(response, request)
	suite.Equal(http.StatusRequestEntityTooLarge, response.GetStatus())
	config.Set("server.maxUploadSize", prev)

	prev = config.Get("server.maxUploadSize")
	config.Set("server.maxUploadSize", 0.0006)
	maxPayloadSize = int64(config.GetFloat("server.maxUploadSize") * 1024 * 1024)
	rawRequest = createTestFileRequest("/test-route?test=hello", "resources/img/logo/goyave_16.png")

	request = createTestRequest(rawRequest)
	response = newResponse(httptest.NewRecorder(), nil)
	parseRequestMiddleware(nil)(response, request)
	suite.Equal(http.StatusRequestEntityTooLarge, response.GetStatus())
	config.Set("server.maxUploadSize", prev)
	maxPayloadSize = int64(config.GetFloat("server.maxUploadSize") * 1024 * 1024)
}

func (suite *MiddlewareTestSuite) TestParseMultipartOverrideMiddleware() {
	executed := false
	rawRequest := createTestFileRequest("/test-route?field=hello", "resources/img/logo/goyave_16.png")
	res := testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
		suite.Equal(2, len(r.Data))
		suite.Equal("world", r.Data["field"])
		files, ok := r.Data["file"].([]fsutil.File)
		suite.True(ok)
		suite.Equal(1, len(files))
		executed = true
	})
	suite.True(executed)
	res.Body.Close()
}

func (suite *MiddlewareTestSuite) TestParseMiddlewareWithArray() {
	executed := false
	rawRequest := httptest.NewRequest("GET", "/test-route?arr=hello&arr=world", nil)
	res := testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
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
	res.Body.Close()

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
	res = testMiddleware(parseRequestMiddleware, rawRequest, nil, validation.RuleSet{}, nil, func(response *Response, r *Request) {
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
	res.Body.Close()
}

func (suite *MiddlewareTestSuite) TestValidateMiddleware() {
	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	data := map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}
	rules := validation.RuleSet{
		"string": validation.List{"required", "string"},
		"number": validation.List{"required", "numeric", "min:10"},
	}
	request := suite.CreateTestRequest(rawRequest)
	request.Data = data
	request.Rules = rules.AsRules()
	result := suite.Middleware(validateRequestMiddleware, request, func(response *Response, r *Request) {})
	suite.Equal(http.StatusNoContent, result.StatusCode)
	result.Body.Close()

	rawRequest = httptest.NewRequest("POST", "/test-route", strings.NewReader("string=hello%20world&number=42"))
	rawRequest.Header.Set("Content-Type", "application/json")
	data = map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}
	rules = validation.RuleSet{
		"string": validation.List{"required", "string"},
		"number": validation.List{"required", "numeric", "min:50"},
	}

	request = suite.CreateTestRequest(rawRequest)
	request.Data = data
	request.Rules = rules.AsRules()
	result = suite.Middleware(validateRequestMiddleware, request, func(response *Response, r *Request) {})
	body, err := io.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	result.Body.Close()
	suite.Equal(http.StatusUnprocessableEntity, result.StatusCode)
	suite.Equal("{\"validationError\":{\"number\":{\"errors\":[\"The number must be at least 50.\"]}}}\n", string(body))

	rawRequest = httptest.NewRequest("POST", "/test-route", nil)
	rawRequest.Header.Set("Content-Type", "application/json")
	request = suite.CreateTestRequest(rawRequest)
	request.Data = nil
	request.Rules = rules.AsRules()
	result = suite.Middleware(validateRequestMiddleware, request, func(response *Response, r *Request) {})
	body, err = io.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	result.Body.Close()
	suite.Equal(http.StatusBadRequest, result.StatusCode)
	suite.Equal("{\"validationError\":{\"[data]\":{\"errors\":[\"Malformed JSON\"]}}}\n", string(body))
}

func (suite *MiddlewareTestSuite) TestCORSMiddleware() {
	// No CORS options
	rawRequest := httptest.NewRequest("GET", "/test-route", nil)
	result := testMiddleware(corsMiddleware, rawRequest, nil, nil, nil, func(response *Response, r *Request) {})
	suite.Equal(200, result.StatusCode)
	result.Body.Close()

	// Preflight
	options := cors.Default()
	rawRequest = httptest.NewRequest("OPTIONS", "/test-route", nil)
	rawRequest.Header.Set("Origin", "https://google.com")
	rawRequest.Header.Set("Access-Control-Request-Method", "GET")
	result = testMiddleware(corsMiddleware, rawRequest, nil, nil, options, func(response *Response, r *Request) {
		response.String(200, "Hi!")
	})
	body, err := io.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	result.Body.Close()
	suite.Equal(204, result.StatusCode)
	suite.Empty(body)

	// Preflight passthrough
	options = cors.Default()
	options.OptionsPassthrough = true
	result = testMiddleware(corsMiddleware, rawRequest, nil, nil, options, func(response *Response, r *Request) {
		response.String(200, "Passthrough")
	})
	body, err = io.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	result.Body.Close()
	suite.Equal(200, result.StatusCode)
	suite.Equal("Passthrough", string(body))

	// Preflight without Access-Control-Request-Method
	rawRequest = httptest.NewRequest("OPTIONS", "/test-route", nil)
	result = testMiddleware(corsMiddleware, rawRequest, nil, nil, options, func(response *Response, r *Request) {
		response.String(200, "Hi!")
	})
	body, err = io.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	result.Body.Close()
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
	body, err = io.ReadAll(result.Body)
	if err != nil {
		panic(err)
	}
	result.Body.Close()
	suite.Equal("Hi!", string(body))
	suite.Equal("https://images.google.com", result.Header.Get("Access-Control-Allow-Origin"))
	suite.Equal("Origin", result.Header.Get("Vary"))
}

func TestMiddlewareTestSuite(t *testing.T) {
	RunTest(t, new(MiddlewareTestSuite))
}
