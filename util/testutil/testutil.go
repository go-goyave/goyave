package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/fsutil"
)

type copyRequestMiddleware struct {
	goyave.Component
	request *goyave.Request
}

func (m *copyRequestMiddleware) Handle(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
		request.Now = m.request.Now
		request.Data = m.request.Data
		request.Extra = m.request.Extra
		request.Lang = m.request.Lang
		request.Query = m.request.Query
		request.RouteParams = m.request.RouteParams
		request.User = m.request.User
		next(response, request)
	}
}

// TestServer extension of `goyave.Server` providing useful functions for testing.
type TestServer struct {
	*goyave.Server
}

// NewTestServer creates a new server using the given config file. The config path is relative to
// the project's directory. If not nil, the given `routeRegistrer` function is called to register
// routes without starting the server.
//
// Automatically closes the DB connection (if there is one) using a test `Cleanup` function.
func NewTestServer(t *testing.T, configFileName string, routeRegistrer func(*goyave.Server, *goyave.Router)) *TestServer {
	rootDirectory := FindRootDirectory()
	cfgPath := rootDirectory + configFileName
	cfg, err := config.LoadFrom(cfgPath)
	if err != nil {
		panic(errors.New(err))
	}

	return NewTestServerWithOptions(t, goyave.Options{Config: cfg}, routeRegistrer)
}

// NewTestServerWithOptions creates a new server using the given options.
// If not nil, the given `routeRegistrer` function is called to register
// routes without starting the server.
//
// Automatically closes the DB connection (if there is one) using a test `Cleanup` function.
func NewTestServerWithOptions(t *testing.T, opts goyave.Options, routeRegistrer func(*goyave.Server, *goyave.Router)) *TestServer {
	srv, err := goyave.New(opts)
	if err != nil {
		panic(err)
	}

	sep := string(os.PathSeparator)
	langDirectory := FindRootDirectory() + sep + "resources" + sep + "lang" + sep
	if err := srv.Lang.LoadDirectory(langDirectory); err != nil {
		panic(err)
	}

	if routeRegistrer != nil {
		srv.RegisterRoutes(routeRegistrer)
	}

	s := &TestServer{srv}
	if t != nil {
		t.Cleanup(func() { s.CloseDB() })
	}
	return s
}

// TestRequest execute a request by calling the root Router's `ServeHTTP()` implementation.
func (s *TestServer) TestRequest(request *http.Request) *http.Response {
	recorder := httptest.NewRecorder()
	s.Router().ServeHTTP(recorder, request)
	return recorder.Result()
}

// TestMiddleware executes with the given request and returns the response.
// The `procedure` parameter is the `next` handler passed to the middleware and can be used to
// make assertions. Keep in mind that this procedure won't be executed if your middleware is blocking.
//
// The request will go through the entire lifecycle like a regular request.
func (s *TestServer) TestMiddleware(middleware goyave.Middleware, request *goyave.Request, procedure goyave.Handler) *http.Response {
	recorder := httptest.NewRecorder()
	router := goyave.NewRouter(s.Server)
	router.GlobalMiddleware(&copyRequestMiddleware{request: request})
	router.Route([]string{request.Method()}, request.Request().URL.Path, procedure).Middleware(middleware)
	router.ServeHTTP(recorder, request.Request())
	return recorder.Result()
}

// CloseDB close the server DB if one is open. It is a good practice to always
// call this in a test `Cleanup` function when using a database.
func (s *TestServer) CloseDB() {
	if err := s.Server.CloseDB(); err != nil {
		s.Logger.Error(err)
	}
}

// FindRootDirectory find relative path to the project's root directory based on the
// existence of a `go.mod` file. The returned path is relative to the working directory
// Returns an empty string if not found.
func FindRootDirectory() string {
	sep := string(os.PathSeparator)
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	directory := wd + sep
	for !fsutil.FileExists(directory + sep + "go.mod") {
		directory += ".." + sep
		if !fsutil.IsDirectory(directory) {
			return ""
		}
	}
	return directory
}

// NewTestRequest create a new `goyave.Request` with an underlying HTTP request created
// usin the `httptest` package.
func NewTestRequest(method, uri string, body io.Reader) *goyave.Request {
	req := httptest.NewRequest(method, uri, body)
	return goyave.NewRequest(req)
}

// NewTestRequest create a new `goyave.Request` with an underlying HTTP request created
// usin the `httptest` package. This function sets the request language using the default
// language of the server.
func (s *TestServer) NewTestRequest(method, uri string, body io.Reader) *goyave.Request {
	req := NewTestRequest(method, uri, body)
	req.Lang = s.Lang.GetDefault()
	return req
}

// NewTestResponse create a new `goyave.Response` with an underlying HTTP response recorder created
// using the `httptest` package. This function uses a temporary `goyave.Server` with all defaults values loaded
// so all functions of `*goyave.Response` can be used safely.
func NewTestResponse(request *goyave.Request) (*goyave.Response, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	return goyave.NewResponse(NewTestServerWithConfig(nil, config.LoadDefault(), nil).Server, request, recorder), recorder
}

// NewTestResponse create a new `goyave.Response` with an underlying HTTP response recorder created
// using the `httptest` package.
func (s *TestServer) NewTestResponse(request *goyave.Request) (*goyave.Response, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	return goyave.NewResponse(s.Server, request, recorder), recorder
}

// ReadJSONBody decodes the given body reader into a new variable of type `*T`.
func ReadJSONBody[T any](body io.Reader) (T, error) {
	var data T
	err := json.NewDecoder(body).Decode(&data)
	if err != nil {
		return data, errors.New(err)
	}
	return data, nil
}

// WriteMultipartFile reads a file from the given path and writes it to the given multipart writer.
func WriteMultipartFile(writer *multipart.Writer, path, fieldName, fileName string) (err error) {
	var file *os.File
	file, err = os.Open(path)
	if err != nil {
		err = errors.New(err)
		return
	}
	defer func() {
		e := file.Close()
		if err == nil && e != nil {
			err = errors.New(e)
		}
	}()
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		err = errors.New(err)
		return
	}
	_, err = io.Copy(part, file)
	if err != nil {
		err = errors.New(err)
	}
	return
}

// CreateTestFiles create a slice of "fsutil.File" from the given paths.
// To reproduce the way the files are obtained in real scenarios,
// files are first encoded in a multipart form, then decoded with
// a multipart form reader.
//
// Paths are relative to the caller, not relative to the project's root directory.
func CreateTestFiles(paths ...string) ([]fsutil.File, error) {
	fieldName := "file"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, p := range paths {
		if err := WriteMultipartFile(writer, p, fieldName, filepath.Base(p)); err != nil {
			return nil, errors.New(err)
		}
	}
	err := writer.Close()
	if err != nil {
		return nil, errors.New(err)
	}

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(math.MaxInt64 - 1)
	if err != nil {
		return nil, errors.New(err)
	}
	return fsutil.ParseMultipartFiles(form.File[fieldName])
}

// ToJSON marshals the given data and creates a bytes reader from the result.
// Panics on error.
func ToJSON(data any) *bytes.Reader {
	res, err := json.Marshal(data)
	if err != nil {
		panic(errors.New(err))
	}
	return bytes.NewReader(res)
}
