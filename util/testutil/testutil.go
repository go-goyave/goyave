package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"math"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"

	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/fsutil"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
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
		request.Route = m.request.Route
		next(response, request.WithContext(m.request.Context()))
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
// A default logger redirecting the output to `testing.T.Log()` is used.
//
// Automatically closes the DB connection (if there is one) using a test `Cleanup` function.
func NewTestServer(t *testing.T, configFileName string) *TestServer {
	rootDirectory := FindRootDirectory()
	cfgPath := path.Join(rootDirectory, configFileName)
	cfg, err := config.LoadFrom(cfgPath)
	if err != nil {
		panic(errors.New(err))
	}

	return NewTestServerWithOptions(t, goyave.Options{Config: cfg})
}

// NewTestServerWithOptions creates a new server using the given options.
// If not nil, the given `routeRegistrer` function is called to register
// routes without starting the server.
//
// By default, if no `Logger` is given in the options, a default logger redirecting the
// output to `testing.T.Log()` is used.
//
// Automatically closes the DB connection (if there is one) using a test `Cleanup` function.
func NewTestServerWithOptions(t *testing.T, opts goyave.Options) *TestServer {
	if opts.Config == nil {
		cfg, err := config.Load()
		if err != nil {
			panic(errors.New(err))
		}
		opts.Config = cfg
	}

	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewHandler(opts.Config.GetBool("app.debug"), &LogWriter{t: t}))
	}

	srv, err := goyave.New(opts)
	if err != nil {
		panic(err)
	}

	langDirectory := path.Join(FindRootDirectory(), "resources", "lang")
	if err := srv.Lang.LoadDirectory(&osfs.FS{}, langDirectory); err != nil {
		panic(err)
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
//
// The given request is cloned. If the middleware alters the request object, these changes won't be reflected on the input request.
// You can do your assertions inside the `procedure`.
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
// existence of a `go.mod` file. The returned path is a rooted path.
// Returns an empty string if not found.
func FindRootDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	directory := wd
	fs := &osfs.FS{}
	for !fs.FileExists(path.Join(directory, "go.mod")) {
		directory = path.Join(directory, "..")
		if !fs.IsDirectory(directory) {
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
	return goyave.NewResponse(NewTestServerWithOptions(nil, goyave.Options{Config: config.LoadDefault()}).Server, request, recorder), recorder
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
	return data, errors.New(err)
}

// WriteMultipartFile reads a file from the given FS and writes it to the given multipart writer.
func WriteMultipartFile(writer *multipart.Writer, filesystem fs.FS, path, fieldName, fileName string) (err error) {
	var file fs.File
	file, err = filesystem.Open(path)
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

// CreateTestFiles create a slice of "fsutil.File" from the given FS.
// To reproduce the way the files are obtained in real scenarios,
// files are first encoded in a multipart form, then decoded with
// a multipart form reader.
//
// Paths are relative to the caller, not relative to the project's root directory.
func CreateTestFiles(fs fs.FS, paths ...string) ([]fsutil.File, error) {
	fieldName := "file"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, p := range paths {
		if err := WriteMultipartFile(writer, fs, p, fieldName, filepath.Base(p)); err != nil {
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
