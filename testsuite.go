package goyave

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
	"goyave.dev/goyave/v4/database"
	"goyave.dev/goyave/v4/util/fsutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	testify "github.com/stretchr/testify/suite"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/lang"
)

// ITestSuite is an extension of testify's Suite for
// Goyave-specific testing.
type ITestSuite interface {
	RunServer(func(*Router), func())
	Timeout() time.Duration
	SetTimeout(time.Duration)
	Middleware(Middleware, *Request, Handler) *http.Response

	Get(string, map[string]string) (*http.Response, error)
	Post(string, map[string]string, io.Reader) (*http.Response, error)
	Put(string, map[string]string, io.Reader) (*http.Response, error)
	Patch(string, map[string]string, io.Reader) (*http.Response, error)
	Delete(string, map[string]string, io.Reader) (*http.Response, error)
	Request(string, string, map[string]string, io.Reader) (*http.Response, error)

	T() *testing.T
	SetT(*testing.T)
	SetS(suite suite.TestingSuite)

	GetBody(*http.Response) []byte
	GetJSONBody(*http.Response, interface{}) error
	CreateTestFiles(paths ...string) []fsutil.File
	WriteFile(*multipart.Writer, string, string, string)
	WriteField(*multipart.Writer, string, string)
	CreateTestRequest(*http.Request) *Request
	CreateTestResponse(http.ResponseWriter) *Response
	getHTTPClient() *http.Client
}

// TestSuite is an extension of testify's Suite for
// Goyave-specific testing.
type TestSuite struct {
	testify.Suite
	httpClient *http.Client
	timeout    time.Duration // Timeout for functional tests
	mu         sync.Mutex
}

var _ ITestSuite = (*TestSuite)(nil) // implements ITestSuite

// Use a mutex to avoid parallel goyave test suites to be run concurrently.
var mu sync.Mutex

// Timeout get the timeout for test failure when using RunServer or requests.
func (s *TestSuite) Timeout() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.timeout
}

// SetTimeout set the timeout for test failure when using RunServer or requests.
func (s *TestSuite) SetTimeout(timeout time.Duration) {
	s.mu.Lock()
	s.timeout = timeout
	s.mu.Unlock()
}

// CreateTestRequest create a "goyave.Request" from the given raw request.
// This function is aimed at making it easier to unit test Requests.
//
// If passed request is "nil", a default GET request to "/" is used.
//
//	rawRequest := httptest.NewRequest("GET", "/test-route", nil)
//	rawRequest.Header.Set("Content-Type", "application/json")
//	request := goyave.CreateTestRequest(rawRequest)
//	request.Lang = "en-US"
//	request.Data = map[string]interface{}{"field": "value"}
func (s *TestSuite) CreateTestRequest(rawRequest *http.Request) *Request {
	if rawRequest == nil {
		rawRequest = httptest.NewRequest("GET", "/", nil)
	}
	return &Request{
		httpRequest: rawRequest,
		route:       nil,
		Data:        nil,
		Rules:       nil,
		Lang:        "en-US",
		Params:      map[string]string{},
		Extra:       map[string]interface{}{},
	}
}

// CreateTestResponse create an empty response with the given response writer.
// This function is aimed at making it easier to unit test Responses.
//
//	writer := httptest.NewRecorder()
//	response := suite.CreateTestResponse(writer)
//	response.Status(http.StatusNoContent)
//	result := writer.Result()
//	fmt.Println(result.StatusCode) // 204
func (s *TestSuite) CreateTestResponse(recorder http.ResponseWriter) *Response {
	return newResponse(recorder, nil)
}

// CreateTestResponseWithRequest create an empty response with the given response writer HTTP request.
// This function is aimed at making it easier to unit test Responses needing the raw request's
// information, such as redirects.
//
//	writer := httptest.NewRecorder()
//	rawRequest := httptest.NewRequest("POST", "/test-route", strings.NewReader("body"))
//	response := suite.CreateTestResponseWithRequest(writer, rawRequest)
//	response.Status(http.StatusNoContent)
//	result := writer.Result()
//	fmt.Println(result.StatusCode) // 204
func (s *TestSuite) CreateTestResponseWithRequest(recorder http.ResponseWriter, rawRequest *http.Request) *Response {
	return newResponse(recorder, rawRequest)
}

// RunServer start the application and run the given functional test procedure.
//
// This function is the equivalent of "goyave.Start()".
// The test fails if the suite's timeout is exceeded.
// The server automatically shuts down when the function ends.
// This function is synchronized, that means that the server is properly stopped
// when the function returns.
func (s *TestSuite) RunServer(routeRegistrer func(*Router), procedure func()) {
	c := make(chan bool, 1)
	c2 := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout())
	defer cancel()

	RegisterStartupHook(func() {
		procedure()
		if ctx.Err() == nil {
			Stop()
			c <- true
		}
	})

	go func() {
		if err := Start(routeRegistrer); err != nil {
			s.Fail(err.Error())
			c <- true
		}
		c2 <- true
	}()

	select {
	case <-ctx.Done():
		s.Fail("Timeout exceeded in goyave.TestSuite.RunServer")
		Stop()
	case sig := <-c:
		s.True(sig)
	}
	ClearStartupHooks()
	ClearShutdownHooks()
	<-c2
}

// Middleware executes the given middleware and returns the HTTP response.
// Core middleware (recovery, parsing and language) is not executed.
func (s *TestSuite) Middleware(middleware Middleware, request *Request, procedure Handler) *http.Response {
	cacheCriticalConfig()
	recorder := httptest.NewRecorder()
	response := s.CreateTestResponse(recorder)
	router := NewRouter()
	router.Middleware(middleware)
	middleware(procedure)(response, request)
	router.finalize(response, request)

	return recorder.Result()
}

// Get execute a GET request on the given route.
// Headers are optional.
func (s *TestSuite) Get(route string, headers map[string]string) (*http.Response, error) {
	return s.Request(http.MethodGet, route, headers, nil)
}

// Post execute a POST request on the given route.
// Headers and body are optional.
func (s *TestSuite) Post(route string, headers map[string]string, body io.Reader) (*http.Response, error) {
	return s.Request(http.MethodPost, route, headers, body)
}

// Put execute a PUT request on the given route.
// Headers and body are optional.
func (s *TestSuite) Put(route string, headers map[string]string, body io.Reader) (*http.Response, error) {
	return s.Request(http.MethodPut, route, headers, body)
}

// Patch execute a PATCH request on the given route.
// Headers and body are optional.
func (s *TestSuite) Patch(route string, headers map[string]string, body io.Reader) (*http.Response, error) {
	return s.Request(http.MethodPatch, route, headers, body)
}

// Delete execute a DELETE request on the given route.
// Headers and body are optional.
func (s *TestSuite) Delete(route string, headers map[string]string, body io.Reader) (*http.Response, error) {
	return s.Request(http.MethodDelete, route, headers, body)
}

// Request execute a request on the given route.
// Headers and body are optional.
func (s *TestSuite) Request(method, route string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, BaseURL()+route, body)
	if err != nil {
		return nil, err
	}
	req.Close = true
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return s.getHTTPClient().Do(req)
}

// GetBody read the whole body of a response.
// If read failed, test fails and return empty byte slice.
func (s *TestSuite) GetBody(response *http.Response) []byte {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		s.Fail("Couldn't read response body", err)
	}
	return body
}

// GetJSONBody read the whole body of a response and decode it as JSON.
// If read or decode failed, test fails.
func (s *TestSuite) GetJSONBody(response *http.Response, data interface{}) error {
	err := json.NewDecoder(response.Body).Decode(data)
	if err != nil {
		s.Fail("Couldn't read response body as JSON", err)
		return err
	}
	return nil
}

// CreateTestFiles create a slice of "fsutil.File" from the given paths.
// Files are passed to a temporary http request and parsed as Multipart form,
// to reproduce the way files are obtained in real scenarios.
func (s *TestSuite) CreateTestFiles(paths ...string) []fsutil.File {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, p := range paths {
		s.WriteFile(writer, p, "file", filepath.Base(p))
	}
	err := writer.Close()
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest("POST", "/test-route", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		panic(err)
	}
	return fsutil.ParseMultipartFiles(req, "file")
}

// WriteFile write a file to the given writer.
// This function is handy for file upload testing.
// The test fails if an error occurred.
func (s *TestSuite) WriteFile(writer *multipart.Writer, path, fieldName, fileName string) {
	file, err := os.Open(path)
	if err != nil {
		s.Fail(err.Error())
		return
	}
	defer file.Close()
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		s.Fail(err.Error())
		return
	}
	_, err = io.Copy(part, file)
	if err != nil {
		s.Fail(err.Error())
	}
}

// WriteField create and write a new multipart form field.
// The test fails if the field couldn't be written.
func (s *TestSuite) WriteField(writer *multipart.Writer, fieldName, value string) {
	if err := writer.WriteField(fieldName, value); err != nil {
		s.Fail(err.Error())
	}
}

// getHTTPClient get suite's http client or create it if it doesn't exist yet.
// The HTTP client is created with a timeout, disabled redirect and disabled TLS cert checking.
func (s *TestSuite) getHTTPClient() *http.Client {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	if s.httpClient == nil {
		s.httpClient = &http.Client{
			Timeout:   s.Timeout(),
			Transport: &http.Transport{TLSClientConfig: config},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	return s.httpClient
}

// ClearDatabase delete all records in all tables.
// This function only clears the tables of registered models, ignoring
// models implementing `database.IView`.
func (s *TestSuite) ClearDatabase() {
	db := database.GetConnection()
	for _, m := range database.GetRegisteredModels() {
		if view, ok := m.(database.IView); ok && view.IsView() {
			continue
		}
		tx := db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(m)
		if tx.Error != nil {
			panic(tx.Error)
		}
	}
}

// ClearDatabaseTables drop all tables.
// This function only clears the tables of registered models.
func (s *TestSuite) ClearDatabaseTables() {
	db := database.GetConnection()
	for _, m := range database.GetRegisteredModels() {
		if err := db.Migrator().DropTable(m); err != nil {
			panic(err)
		}
	}
}

// RunTest run a test suite with prior initialization of a test environment.
// The GOYAVE_ENV environment variable is automatically set to "test" and restored
// to its original value at the end of the test run.
// All tests are run using your project's root as working directory. This directory is determined
// by the presence of a "go.mod" file.
func RunTest(t *testing.T, suite ITestSuite) bool {
	mu.Lock()
	defer mu.Unlock()
	if suite.Timeout() == 0 {
		suite.SetTimeout(5 * time.Second)
	}
	_, ok := os.LookupEnv("GOYAVE_ENV")
	if !ok {
		os.Setenv("GOYAVE_ENV", "test")
		defer os.Unsetenv("GOYAVE_ENV")
	}
	setRootWorkingDirectory()

	if !config.IsLoaded() {
		if err := config.Load(); err != nil {
			return assert.Fail(t, "Failed to load config", err)
		}
	}
	defer config.Clear()
	lang.LoadDefault()
	lang.LoadAllAvailableLanguages()

	if config.GetBool("database.autoMigrate") && config.GetString("database.connection") != "none" {
		database.Migrate()
	}

	testify.Run(t, suite)

	database.Close()
	return !t.Failed()
}

func setRootWorkingDirectory() {
	sep := string(os.PathSeparator)
	_, filename, _, _ := runtime.Caller(2)
	directory := path.Dir(filename) + sep
	for !fsutil.FileExists(directory + sep + "go.mod") {
		directory += ".." + sep
		if !fsutil.IsDirectory(directory) {
			panic("Couldn't find project's root directory.")
		}
	}
	if err := os.Chdir(directory); err != nil {
		panic(err)
	}
}
