package goyave

import (
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/fsutil"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
)

func TestCleanStaticPath(t *testing.T) {

	cases := []struct {
		directory string
		file      string
		want      string
	}{
		{directory: ".", file: "config/index.html", want: "config/index.html"},
		{directory: ".", file: "config", want: "config/index.html"},
		{directory: ".", file: "config/", want: "config/index.html"},
		{directory: ".", file: "config/defaults.json", want: "config/defaults.json"},
		{directory: "config", file: "index.html", want: "index.html"},
		{directory: "config", file: "", want: "index.html"},
		{directory: "config", file: "defaults.json", want: "defaults.json"},
		{directory: "resources", file: "lang/en-US/locale.json", want: "lang/en-US/locale.json"},
		{directory: "resources", file: "/lang/en-US/locale.json", want: "lang/en-US/locale.json"},
		{directory: "resources", file: "img/logo", want: "img/logo/index.html"},
		{directory: "resources", file: "img/logo/", want: "img/logo/index.html"},
		{directory: "resources", file: "img", want: "img/index.html"},
		{directory: "resources", file: "img/", want: "img/index.html"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.want, func(t *testing.T) {
			f, err := fs.Sub(&osfs.FS{}, c.directory)
			require.NoError(t, err)
			assert.Equal(t, c.want, cleanStaticPath(fsutil.NewEmbed(f.(fs.ReadDirFS)), c.file))
		})
	}
}

func TestStaticHandler(t *testing.T) {

	cases := []struct {
		expected  func(*testing.T, *Response, *http.Response, []byte)
		uri       string
		directory string
		download  bool
	}{
		{
			uri:       "/custom_config.json",
			directory: "resources",
			download:  false,
			expected: func(t *testing.T, response *Response, result *http.Response, body []byte) {
				assert.Equal(t, http.StatusOK, response.GetStatus())
				assert.Equal(t, "application/json", result.Header.Get("Content-Type"))
				assert.Equal(t, "inline", result.Header.Get("Content-Disposition"))
				assert.Equal(t, "{\n    \"custom-entry\": \"value\"\n}", string(body))
			},
		},
		{
			uri:       "/doesn'texist",
			directory: "resources",
			download:  false,
			expected: func(t *testing.T, response *Response, _ *http.Response, _ []byte) {
				assert.Equal(t, http.StatusNotFound, response.GetStatus())
			},
		},
		{
			uri:       "/custom_config.json",
			directory: "resources",
			download:  true,
			expected: func(t *testing.T, response *Response, result *http.Response, body []byte) {
				assert.Equal(t, http.StatusOK, response.GetStatus())
				assert.Equal(t, "application/json", result.Header.Get("Content-Type"))
				assert.Equal(t, "attachment; filename=\"custom_config.json\"", result.Header.Get("Content-Disposition"))
				assert.Equal(t, "{\n    \"custom-entry\": \"value\"\n}", string(body))
			},
		},
		{
			uri:       "/lang/en-US/fields.json",
			directory: "resources",
			download:  true,
			expected: func(t *testing.T, response *Response, result *http.Response, body []byte) {
				assert.Equal(t, http.StatusOK, response.GetStatus())
				assert.Equal(t, "application/json", result.Header.Get("Content-Type"))
				assert.Equal(t, "attachment; filename=\"fields.json\"", result.Header.Get("Content-Disposition"))
				assert.Equal(t, "{\n    \"email\": \"email address\"\n}", string(body))
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.uri, func(t *testing.T) {

			cfg := config.LoadDefault()
			srv, err := New(Options{Config: cfg})
			require.NoError(t, err)

			request := NewRequest(httptest.NewRequest(http.MethodGet, c.uri, nil))
			request.RouteParams = map[string]string{"resource": c.uri}

			recorder := httptest.NewRecorder()
			response := NewResponse(srv, request, recorder)

			f, err := fs.Sub(&osfs.FS{}, c.directory)
			require.NoError(t, err)
			handler := staticHandler(fsutil.NewEmbed(f.(fs.ReadDirFS)), c.download)
			handler(response, request)

			result := recorder.Result()
			body, err := io.ReadAll(result.Body)
			assert.NoError(t, result.Body.Close())
			require.NoError(t, err)
			c.expected(t, response, result, body)
		})
	}
}
