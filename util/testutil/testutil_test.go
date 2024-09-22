package testutil

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/fsutil"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
)

type extraKey struct{}

type testMiddleware struct {
	goyave.Component
	procedure goyave.Handler
}

func (m *testMiddleware) Handle(_ goyave.Handler) goyave.Handler {
	return m.procedure
}

func TestTestServer(t *testing.T) {
	t.Run("NewTestServer", func(t *testing.T) {
		server := NewTestServer(t, "resources/custom_config.json")
		assert.Equal(t, "value", server.Config().Get("custom-entry"))
		assert.Equal(t, slog.DiscardLogger(), server.Logger)
	})

	t.Run("NewTestServerWithOptions", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("test-entry", "test-value")

		server := NewTestServerWithOptions(t, goyave.Options{Config: cfg})

		assert.NotNil(t, server.Lang)
		assert.Equal(t, "test-value", server.Config().Get("test-entry"))
		assert.Equal(t, slog.DiscardLogger(), server.Logger)
	})

	t.Run("NewTestServer_AutoConfig", func(t *testing.T) {
		t.Setenv("GOYAVE_ENV", "test")
		server := NewTestServerWithOptions(t, goyave.Options{})

		assert.NotNil(t, server.Lang)
		assert.Equal(t, "test-value", server.Config().Get("test-entry"))

		assert.Panics(t, func() {
			// Config file not found
			t.Setenv("GOYAVE_ENV", "")
			NewTestServerWithOptions(t, goyave.Options{})
		})
	})

	t.Run("TestRequest", func(t *testing.T) {
		server := NewTestServerWithOptions(t, goyave.Options{Config: config.LoadDefault()})
		server.RegisterRoutes(func(_ *goyave.Server, r *goyave.Router) {
			r.Get("/route", func(resp *goyave.Response, _ *goyave.Request) {
				resp.String(http.StatusOK, "OK")
			})
		})

		resp := server.TestRequest(httptest.NewRequest(http.MethodGet, "/route", nil))
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, "OK", string(respBody))
	})

	t.Run("TestMiddleware", func(t *testing.T) {
		server := NewTestServerWithOptions(t, goyave.Options{Config: config.LoadDefault()})

		request := server.NewTestRequest(http.MethodGet, "/route", nil)
		request.Data = map[string]any{"key": "value"}
		request.Extra = map[any]any{extraKey{}: "value"}
		request.Query = map[string]any{"key": "value"}
		request.RouteParams = map[string]string{"key": "value"}
		request.User = map[string]string{"key": "value"}

		middleware := &testMiddleware{
			procedure: func(resp *goyave.Response, req *goyave.Request) {
				assert.Equal(t, request.Now, req.Now)
				assert.Equal(t, request.Data, req.Data)
				assert.Equal(t, request.Extra, req.Extra)
				assert.Equal(t, request.Lang, req.Lang)
				assert.Equal(t, request.Query, req.Query)
				assert.Equal(t, request.RouteParams, req.RouteParams)
				assert.Equal(t, request.User, req.User)
				resp.String(http.StatusOK, "OK")
			},
		}

		resp := server.TestMiddleware(middleware, request, func(response *goyave.Response, _ *goyave.Request) {
			response.Status(http.StatusBadRequest)
		})
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, "OK", string(respBody))
	})

	t.Run("NewTestRequest", func(t *testing.T) {
		server := NewTestServerWithOptions(t, goyave.Options{Config: config.LoadDefault()})
		body := bytes.NewBufferString("body")
		req := server.NewTestRequest(http.MethodPost, "/uri", body)

		assert.Equal(t, http.MethodPost, req.Method())
		assert.Equal(t, "/uri", req.URL().String())
		assert.NotNil(t, req.Extra)

		b, err := io.ReadAll(req.Body())
		require.NoError(t, err)
		assert.Equal(t, "body", string(b))
		assert.Equal(t, server.Lang.GetDefault(), req.Lang)
	})

	t.Run("NewTestResponse", func(t *testing.T) {
		server := NewTestServerWithOptions(t, goyave.Options{Config: config.LoadDefault(), Logger: slog.New(slog.NewHandler(false, &bytes.Buffer{}))})
		req := server.NewTestRequest(http.MethodGet, "/uri", nil)
		resp, recorder := server.NewTestResponse(req)

		resp.String(http.StatusOK, "hello")
		result := recorder.Result()
		b, err := io.ReadAll(result.Body)
		require.NoError(t, result.Body.Close())
		require.NoError(t, err)
		assert.Equal(t, "hello", string(b))
		assert.NotPanics(t, func() {
			// No panics because the server is accessible so the ErrLogger.Println succeeds
			resp.Error(nil)
		})
	})
}

func TestFindRootDirectory(t *testing.T) {
	dir := FindRootDirectory()
	assert.True(t, fsutil.FileExists(&osfs.FS{}, path.Join(dir, "go.mod")))
}

func TestNewTestRequest(t *testing.T) {
	body := bytes.NewBufferString("body")
	req := NewTestRequest(http.MethodPost, "/uri", body)

	assert.Equal(t, http.MethodPost, req.Method())
	assert.Equal(t, "/uri", req.URL().String())
	assert.NotNil(t, req.Extra)

	b, err := io.ReadAll(req.Body())
	require.NoError(t, err)
	assert.Equal(t, "body", string(b))
}

func TestNewTestResponse(t *testing.T) {
	req := NewTestRequest(http.MethodGet, "/uri", nil)
	resp, recorder := NewTestResponse(req)

	resp.String(http.StatusOK, "hello")
	result := recorder.Result()
	b, err := io.ReadAll(result.Body)
	require.NoError(t, result.Body.Close())
	require.NoError(t, err)
	assert.Equal(t, "hello", string(b))
}

func TestReadJSONBody(t *testing.T) {
	body := bytes.NewBufferString(`{"key":"value"}`)
	jsonBody, err := ReadJSONBody[map[string]string](body)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"key": "value"}, jsonBody)

	jsonBodyError, err := ReadJSONBody[string](body)
	require.Error(t, err)
	assert.Empty(t, jsonBodyError)
}

func TestWriteMultipartFile(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, WriteMultipartFile(writer, &osfs.FS{}, "../../resources/img/logo/goyave_16.png", "profile_picture", "goyave_16.png"))
	require.NoError(t, writer.Close())

	req := NewTestRequest(http.MethodPost, "/uri", body)
	req.Header().Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, req.Request().ParseMultipartForm(1024*1024*1024))

	files := req.Request().MultipartForm.File
	require.Len(t, files, 1)
	require.Contains(t, files, "profile_picture")

	file := files["profile_picture"]
	require.Len(t, file, 1)
	assert.Equal(t, "goyave_16.png", file[0].Filename)
	assert.Equal(t, int64(630), file[0].Size)
	assert.Equal(t, textproto.MIMEHeader{"Content-Type": []string{"application/octet-stream"}, "Content-Disposition": []string{"form-data; name=\"profile_picture\"; filename=\"goyave_16.png\""}}, file[0].Header)
}

func TestCreateTestFiles(t *testing.T) {
	files, err := CreateTestFiles(&osfs.FS{}, "../../resources/img/logo/goyave_16.png", "../../resources/test_file.txt")
	require.NoError(t, err)
	require.Len(t, files, 2)

	assert.Equal(t, "goyave_16.png", files[0].Header.Filename)
	assert.Equal(t, int64(630), files[0].Header.Size)
	assert.Equal(t, textproto.MIMEHeader{"Content-Type": []string{"application/octet-stream"}, "Content-Disposition": []string{"form-data; name=\"file\"; filename=\"goyave_16.png\""}}, files[0].Header.Header)

	assert.Equal(t, "test_file.txt", files[1].Header.Filename)
	assert.Equal(t, int64(25), files[1].Header.Size)
	assert.Equal(t, textproto.MIMEHeader{"Content-Type": []string{"application/octet-stream"}, "Content-Disposition": []string{"form-data; name=\"file\"; filename=\"test_file.txt\""}}, files[1].Header.Header)
}

func TestToJSON(t *testing.T) {
	reader := ToJSON(map[string]any{"key": "value"})
	result, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, `{"key":"value"}`, string(result))

	assert.Panics(t, func() {
		ToJSON(make(chan struct{}))
	})
}
