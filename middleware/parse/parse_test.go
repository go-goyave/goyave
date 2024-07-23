package parse

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/fsutil"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
	"goyave.dev/goyave/v5/util/testutil"
)

func TestParseMiddleware(t *testing.T) {
	server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: config.LoadDefault()})

	t.Run("Max Upload Size", func(t *testing.T) {
		m := &Middleware{}
		m.Init(server.Server)
		assert.InEpsilon(t, 10.0, m.getMaxUploadSize(), 0) // Default
		m.MaxUploadSize = 2.3
		assert.InEpsilon(t, 2.3, m.getMaxUploadSize(), 0)

		m = &Middleware{
			MaxUploadSize: 2.3,
		}
		m.Init(server.Server)
		assert.InEpsilon(t, 2.3, m.getMaxUploadSize(), 0)
	})

	t.Run("Parse Query", func(t *testing.T) {
		request := testutil.NewTestRequest(http.MethodGet, "/parse?a=b&c=d&array=1&array=2", nil)

		result := server.TestMiddleware(&Middleware{}, request, func(_ *goyave.Response, req *goyave.Request) {
			expected := map[string]any{
				"a":     "b",
				"c":     "d",
				"array": []string{"1", "2"},
			}
			assert.Equal(t, expected, req.Query)
		})
		assert.NoError(t, result.Body.Close())
	})

	t.Run("Parse Query Error", func(t *testing.T) {
		request := testutil.NewTestRequest(http.MethodGet, "/parse?inv;alid", nil)
		request.Lang = server.Lang.GetDefault()

		result := server.TestMiddleware(&Middleware{}, request, func(_ *goyave.Response, req *goyave.Request) {
			assert.Equal(t, map[string]any{}, req.Query)
		})
		assert.NoError(t, result.Body.Close())
		assert.Equal(t, http.StatusBadRequest, result.StatusCode)
		assert.NotNil(t, request.Extra[goyave.ExtraParseError{}])
		assert.NotPanics(t, func() {
			extraError, ok := request.Extra[goyave.ExtraParseError{}].(error)
			require.True(t, ok)
			assert.ErrorIs(t, extraError, goyave.ErrInvalidQuery)
		})
	})

	t.Run("Entity Too Large", func(t *testing.T) {
		request := testutil.NewTestRequest(http.MethodPost, "/parse", strings.NewReader(strings.Repeat("a", 1024*1024)))
		request.Header().Set("Content-Type", "application/octet-stream")

		result := server.TestMiddleware(&Middleware{MaxUploadSize: 0.01}, request, func(_ *goyave.Response, _ *goyave.Request) {
			assert.Fail(t, "Middleware should not pass")
		})
		assert.NoError(t, result.Body.Close())
		assert.Equal(t, http.StatusRequestEntityTooLarge, result.StatusCode)
	})

	t.Run("JSON", func(t *testing.T) {
		data := map[string]any{
			"a": "b",
			"c": "d",
			"e": map[string]any{
				"f": "g",
			},
			"h": []string{"i", "j"},
		}

		request := testutil.NewTestRequest(http.MethodPost, "/parse", testutil.ToJSON(data))
		request.Header().Set("Content-Type", "application/json")

		result := server.TestMiddleware(&Middleware{}, request, func(resp *goyave.Response, req *goyave.Request) {
			expected := map[string]any{
				"a": "b",
				"c": "d",
				"e": map[string]any{
					"f": "g",
				},
				"h": []any{"i", "j"},
			}
			assert.Equal(t, expected, req.Data)
			resp.Status(http.StatusOK)
		})

		assert.NoError(t, result.Body.Close())
		assert.Equal(t, http.StatusOK, result.StatusCode)
	})

	t.Run("JSON Parsing Tests", func(t *testing.T) {
		tests := []struct {
			name     string
			body     []byte
			expected int
		}{
			{
				name:     "JSON Invalid",
				body:     []byte(`{"unclosed"`),
				expected: http.StatusBadRequest,
			},
			{
				name:     "JSON Empty",
				body:     []byte(""),
				expected: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				request := testutil.NewTestRequest(http.MethodPost, "/parse", bytes.NewBuffer(tt.body))
				request.Lang = server.Lang.GetDefault()
				request.Header().Set("Content-Type", "application/json")

				result := server.TestMiddleware(&Middleware{MaxUploadSize: 0.01}, request, func(_ *goyave.Response, _ *goyave.Request) {
					assert.Fail(t, "Middleware should not pass")
				})

				assert.NoError(t, result.Body.Close())
				assert.Equal(t, tt.expected, result.StatusCode)
				assert.NotNil(t, request.Extra[goyave.ExtraParseError{}])
				assert.NotPanics(t, func() {
					extraError, ok := request.Extra[goyave.ExtraParseError{}].(error)
					require.True(t, ok)
					assert.ErrorIs(t, extraError, goyave.ErrInvalidJSONBody)
				})
			})
		}
	})

	t.Run("Multipart", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		require.NoError(t, testutil.WriteMultipartFile(writer, &osfs.FS{}, "../../resources/img/logo/goyave_16.png", "profile_picture", "goyave_16.png"))
		require.NoError(t, writer.WriteField("email", "johndoe@example.org"))

		request := testutil.NewTestRequest(http.MethodPost, "/parse", body)
		request.Header().Set("Content-Type", writer.FormDataContentType())

		require.NoError(t, writer.Close())

		result := server.TestMiddleware(&Middleware{}, request, func(resp *goyave.Response, req *goyave.Request) {
			data, ok := req.Data.(map[string]any)
			if !assert.True(t, ok) {
				return
			}
			assert.Equal(t, "johndoe@example.org", data["email"])

			picture, ok := data["profile_picture"].([]fsutil.File)
			if !assert.True(t, ok) {
				return
			}

			if !assert.Len(t, picture, 1) {
				return
			}
			assert.Equal(t, "image/png", picture[0].MIMEType)
			assert.Equal(t, "goyave_16.png", picture[0].Header.Filename)
			assert.Equal(t, int64(630), picture[0].Header.Size)

			resp.Status(http.StatusOK)
		})

		assert.NoError(t, result.Body.Close())
		assert.Equal(t, http.StatusOK, result.StatusCode)
	})

	t.Run("Invalid Multipart", func(t *testing.T) {
		// Write empty body, which is not allowed for content multipart.
		writer := multipart.NewWriter(nil)

		request := testutil.NewTestRequest(http.MethodPost, "/parse", nil)
		request.Lang = server.Lang.GetDefault()
		request.Header().Set("Content-Type", writer.FormDataContentType())

		result := server.TestMiddleware(&Middleware{}, request, func(resp *goyave.Response, _ *goyave.Request) {
			resp.Status(http.StatusOK)
		})

		assert.NoError(t, result.Body.Close())
		assert.Equal(t, http.StatusBadRequest, result.StatusCode)
		assert.NotNil(t, request.Extra[goyave.ExtraParseError{}])
		assert.NotPanics(t, func() {
			extraError, ok := request.Extra[goyave.ExtraParseError{}].(error)
			require.True(t, ok)
			assert.ErrorIs(t, extraError, goyave.ErrInvalidContentForType)
		})
	})

	t.Run("Error Reading Request Body", func(t *testing.T) {
		// Create a test server that sends partial data
		faultyRequest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "10")
			w.WriteHeader(http.StatusOK)
			// Write only part of the promised data
			_, err := w.Write([]byte("Partial "))
			assert.NoError(t, err)

			// Then flush the writer to send the incomplete response
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			// Don't close the connection, just stop writing
		}))
		defer faultyRequest.Close()

		// Create a client that will make a request to our test server
		client := faultyRequest.Client()
		resp, err := client.Get(faultyRequest.URL)
		require.NoError(t, err)

		// Use the response body from our test server as the request body for our middleware
		request := testutil.NewTestRequest(http.MethodPost, "/parse", resp.Body)
		request.Lang = server.Lang.GetDefault()
		request.Header().Set("Content-Type", "multipart/form-data")

		result := server.TestMiddleware(&Middleware{}, request, func(resp *goyave.Response, _ *goyave.Request) {
			resp.Status(http.StatusBadRequest)
		})

		assert.NoError(t, result.Body.Close())
		assert.NoError(t, resp.Body.Close())

		assert.Equal(t, http.StatusBadRequest, result.StatusCode)
		assert.NotNil(t, request.Extra[goyave.ExtraParseError{}])
		assert.NotPanics(t, func() {
			extraError, ok := request.Extra[goyave.ExtraParseError{}].(error)
			require.True(t, ok)
			assert.ErrorIs(t, extraError, goyave.ErrErrorInRequestBody)
			assert.Contains(t, extraError.Error(), "unexpected EOF")
		})
	})

	t.Run("Form URL-encoded", func(t *testing.T) {
		data := "a=b&c=d&h=i&h=j"

		request := testutil.NewTestRequest(http.MethodPost, "/parse", strings.NewReader(data))
		request.Header().Set("Content-Type", "application/x-www-form-urlencoded; param=value")

		result := server.TestMiddleware(&Middleware{}, request, func(resp *goyave.Response, req *goyave.Request) {
			expected := map[string]any{
				"a": "b",
				"c": "d",
				"h": []string{"i", "j"},
			}
			assert.Equal(t, expected, req.Data)
			resp.Status(http.StatusOK)
		})

		assert.NoError(t, result.Body.Close())
		assert.Equal(t, http.StatusOK, result.StatusCode)
	})

	t.Run("Body already parsed", func(t *testing.T) {
		data := map[string]any{
			"a": "b",
			"c": "d",
			"e": map[string]any{
				"f": "g",
			},
			"h": []string{"i", "j"},
		}
		request := testutil.NewTestRequest(http.MethodPost, "/parse?a=b&c=d&array=1&array=2", testutil.ToJSON(data))
		request.Data = map[string]any{"a": "b"}

		result := server.TestMiddleware(&Middleware{}, request, func(_ *goyave.Response, req *goyave.Request) {
			expectedQuery := map[string]any{
				"a":     "b",
				"c":     "d",
				"array": []string{"1", "2"},
			}
			assert.Equal(t, expectedQuery, req.Query) // Query parsed but not body
			assert.Equal(t, map[string]any{"a": "b"}, req.Data)
		})
		assert.NoError(t, result.Body.Close())
	})
}
