package goyave

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5/config"
)

func TestCleanStaticPath(t *testing.T) {

	cases := []struct {
		directory string
		file      string
		want      string
	}{
		{directory: "config", file: "index.html", want: "config/index.html"},
		{directory: "config", file: "", want: "config/index.html"},
		{directory: "config", file: "defaults.json", want: "config/defaults.json"},
		{directory: "resources", file: "lang/en-US/locale.json", want: "resources/lang/en-US/locale.json"},
		{directory: "resources", file: "/lang/en-US/locale.json", want: "resources/lang/en-US/locale.json"},
		{directory: "resources", file: "img/logo", want: "resources/img/logo/index.html"},
		{directory: "resources", file: "img/logo/", want: "resources/img/logo/index.html"},
		{directory: "resources", file: "img", want: "resources/img/index.html"},
		{directory: "resources", file: "img/", want: "resources/img/index.html"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.want, func(t *testing.T) {
			assert.Equal(t, c.want, cleanStaticPath(c.directory, c.file))
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
			expected: func(t *testing.T, response *Response, result *http.Response, body []byte) {
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
	}

	for _, c := range cases {
		c := c
		t.Run(c.uri, func(t *testing.T) {

			cfg := config.LoadDefault()
			srv, err := New(Options{Config: cfg})
			if err != nil {
				panic(err)
			}

			request := NewRequest(httptest.NewRequest(http.MethodGet, c.uri, nil))
			request.RouteParams = map[string]string{"resource": c.uri}

			recorder := httptest.NewRecorder()
			response := NewResponse(srv, request, recorder)

			handler := staticHandler(c.directory, c.download)
			handler(response, request)

			result := recorder.Result()
			body, err := io.ReadAll(result.Body)
			assert.NoError(t, result.Body.Close())
			if err != nil {
				panic(err)
			}
			c.expected(t, response, result, body)
		})
	}
}
