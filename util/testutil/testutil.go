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

	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/util/fsutil"
)

type copyRequestMiddleware struct {
	goyave.Component
	request *goyave.RequestV5
}

func (m *copyRequestMiddleware) Handle(next goyave.HandlerV5) goyave.HandlerV5 {
	return func(response *goyave.ResponseV5, request *goyave.RequestV5) {
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
func NewTestServer(configFileName string, routeRegistrer func(*goyave.Server, *goyave.RouterV5)) *TestServer {
	rootDirectory := FindRootDirectory()
	cfgPath := rootDirectory + configFileName
	cfg, err := config.LoadFromV5(cfgPath)
	if err != nil {
		panic(err)
	}

	return NewTestServerWithConfig(cfg, routeRegistrer)
}

// NewTestServerWithConfig creates a new server using the given config.
// If not nil, the given `routeRegistrer` function is called to register
// routes without starting the server.
func NewTestServerWithConfig(cfg *config.Config, routeRegistrer func(*goyave.Server, *goyave.RouterV5)) *TestServer {
	srv, err := goyave.NewWithConfig(cfg)
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
	return &TestServer{srv}
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
func (s *TestServer) TestMiddleware(middleware goyave.MiddlewareV5, request *goyave.RequestV5, procedure goyave.HandlerV5) *http.Response {
	recorder := httptest.NewRecorder()
	router := goyave.NewRouterV5(s.Server)
	router.GlobalMiddleware(&copyRequestMiddleware{request: request})
	router.Route([]string{request.Method()}, request.Request().URL.Path, procedure).Middleware(middleware)
	router.ServeHTTP(recorder, request.Request())
	return recorder.Result()
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
func NewTestRequest(method, uri string, body io.Reader) *goyave.RequestV5 {
	req := httptest.NewRequest(method, uri, body)
	return goyave.NewRequest(req)
}

// NewTestRequest create a new `goyave.Request` with an underlying HTTP request created
// usin the `httptest` package. This function sets the request language using the default
// language of the server.
func (s *TestServer) NewTestRequest(method, uri string, body io.Reader) *goyave.RequestV5 {
	req := NewTestRequest(method, uri, body)
	req.Lang = s.Lang.GetDefault()
	return req
}

// NewTestResponse create a new `goyave.Response` with an underlying HTTP response recorder created
// using the `httptest` package. This function uses a temporary `goyave.Server` with all defaults values loaded
// so all functions of `*goyave.Response` can be used safely.
func NewTestResponse(request *goyave.RequestV5) (*goyave.ResponseV5, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	return goyave.NewResponse(NewTestServerWithConfig(config.LoadDefault(), nil).Server, request, recorder), recorder
}

// NewTestResponse create a new `goyave.Response` with an underlying HTTP response recorder created
// using the `httptest` package.
func (s *TestServer) NewTestResponse(request *goyave.RequestV5) (*goyave.ResponseV5, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	return goyave.NewResponse(s.Server, request, recorder), recorder
}

// ReadJSONBody decodes the given body reader into a new variable of type `*T`.
func ReadJSONBody[T any](body io.Reader) (T, error) {
	var data T
	err := json.NewDecoder(body).Decode(&data)
	if err != nil {
		return data, err
	}
	return data, nil
}

// WriteMultipartFile reads a file from the given path and writes it to the given multipart writer.
func WriteMultipartFile(writer *multipart.Writer, path, fieldName, fileName string) (err error) {
	var file *os.File
	file, err = os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		e := file.Close()
		if err == nil {
			err = e
		}
	}()
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return
	}
	_, err = io.Copy(part, file)
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
			return nil, err
		}
	}
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(math.MaxInt64 - 1)
	if err != nil {
		return nil, err
	}
	return fsutil.ParseMultipartFiles(form.File[fieldName])
}

// ToJSON marshals the given data and creates a bytes reader from the result.
// Panics on error.
func ToJSON(data any) *bytes.Reader {
	res, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(res)
}
