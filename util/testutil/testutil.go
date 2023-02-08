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
	"path"
	"path/filepath"
	"runtime"

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
func NewTestServer(configFileName string, routeRegistrer func(*goyave.Server, *goyave.RouterV5)) (*TestServer, error) {
	rootDirectory := FindRootDirectory()
	cfgPath := rootDirectory + configFileName
	cfg, err := config.LoadFromV5(cfgPath)
	if err != nil {
		return nil, err
	}

	srv, err := goyave.NewWithConfig(cfg)
	if err != nil {
		return nil, err
	}

	sep := string(os.PathSeparator)
	langDirectory := rootDirectory + sep + "resources" + sep + "lang" + sep
	if err := srv.Lang.LoadDirectory(langDirectory); err != nil {
		return nil, err
	}

	if routeRegistrer != nil {
		srv.RegisterRoutes(routeRegistrer)
	}
	return &TestServer{srv}, nil
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
	router.Route([]string{request.Method()}, request.Request().RequestURI, procedure).Middleware(middleware)
	router.ServeHTTP(recorder, request.Request())
	return recorder.Result()
}

// FindRootDirectory find relative path to the project's root directory based on the
// existence of a `go.mod` file. The returned path is relative to the source file of the caller.
// Returns an empty string if not found.
func FindRootDirectory() string {
	sep := string(os.PathSeparator)
	_, filename, _, _ := runtime.Caller(2)
	directory := path.Dir(filename) + sep
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

// ReadJSONBody decodes the given body reader into a new variable of type `*T`.
func ReadJSONBody[T any](body io.Reader) (T, error) {
	var data T
	err := json.NewDecoder(body).Decode(&data)
	if err != nil {
		return data, err
	}
	return data, nil
}

// WriteMultipartFile reads the file at the given path and writes it to the
// given multipart writer.
func WriteMultipartFile(writer *multipart.Writer, path, fieldName, fileName string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	return err
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
