package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/testutil"
)

type closeableChildWriter struct {
	io.Writer
	closed     bool
	preWritten bool
}

func (w *closeableChildWriter) PreWrite(b []byte) {
	w.preWritten = true
	if pr, ok := w.Writer.(goyave.PreWriter); ok {
		pr.PreWrite(b)
	}
}

func (w *closeableChildWriter) Close() error {
	w.closed = true
	if wr, ok := w.Writer.(io.Closer); ok {
		return wr.Close()
	}
	return nil
}

func TestCompressMiddleware(t *testing.T) {

	server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: config.LoadDefault()}, nil)

	handler := func(resp *goyave.Response, req *goyave.Request) {
		resp.Header().Set("Content-Length", "1234")
		resp.String(http.StatusOK, "hello world")
	}

	compressMiddleware := &Middleware{
		Encoders: []Encoder{
			&Gzip{Level: gzip.BestCompression},
		},
	}

	t.Run("No compression", func(t *testing.T) {
		request := testutil.NewTestRequest(http.MethodGet, "/gzip", nil)

		result := server.TestMiddleware(compressMiddleware, request, handler)

		body, err := io.ReadAll(result.Body)
		if err != nil {
			panic(err)
		}
		assert.NoError(t, result.Body.Close())
		assert.Equal(t, "hello world", string(body)) // Not compressed
		assert.NotEqual(t, "gzip", result.Header.Get("Content-Encoding"))
		assert.Equal(t, result.Header.Get("Content-Length"), "1234")
		assert.Equal(t, http.StatusOK, result.StatusCode)
	})

	t.Run("Gzip", func(t *testing.T) {
		request := testutil.NewTestRequest(http.MethodGet, "/gzip", nil)
		request.Header().Set("Accept-Encoding", "gzip")
		result := server.TestMiddleware(compressMiddleware, request, handler)

		reader, err := gzip.NewReader(result.Body)
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(reader)
		if err != nil {
			panic(err)
		}
		assert.NoError(t, result.Body.Close())
		assert.Equal(t, "hello world", string(body))
		assert.Equal(t, "gzip", result.Header.Get("Content-Encoding"))
		assert.Empty(t, result.Header.Get("Content-Length"))
	})

	t.Run("Accept all", func(t *testing.T) {
		request := testutil.NewTestRequest(http.MethodGet, "/gzip", nil)
		request.Header().Set("Accept-Encoding", "*")
		result := server.TestMiddleware(compressMiddleware, request, handler)

		assert.Equal(t, "gzip", result.Header.Get("Content-Encoding"))

		reader, err := gzip.NewReader(result.Body)
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(reader)
		if err != nil {
			panic(err)
		}
		assert.NoError(t, result.Body.Close())
		assert.Equal(t, "hello world", string(body))
		assert.Empty(t, result.Header.Get("Content-Length"))
	})

	t.Run("Unsupported encoding", func(t *testing.T) {
		request := testutil.NewTestRequest(http.MethodGet, "/gzip", nil)
		request.Header().Set("Accept-Encoding", "bz")

		result := server.TestMiddleware(compressMiddleware, request, handler)

		body, err := io.ReadAll(result.Body)
		if err != nil {
			panic(err)
		}
		assert.NoError(t, result.Body.Close())
		assert.Equal(t, "hello world", string(body)) // Not compressed
		assert.Empty(t, result.Header.Get("Content-Encoding"))
		assert.Equal(t, http.StatusOK, result.StatusCode)
	})

	t.Run("Upgrade", func(t *testing.T) {
		request := testutil.NewTestRequest(http.MethodGet, "/gzip", nil)
		request.Header().Set("Accept-Encoding", "gzip")
		request.Header().Set("Upgrade", "example/1, foo/2")

		result := server.TestMiddleware(compressMiddleware, request, handler)

		body, err := io.ReadAll(result.Body)
		if err != nil {
			panic(err)
		}
		assert.NoError(t, result.Body.Close())
		assert.Equal(t, "hello world", string(body)) // Not compressed
		assert.NotEqual(t, "gzip", result.Header.Get("Content-Encoding"))
		assert.Equal(t, result.Header.Get("Content-Length"), "1234")
		assert.Equal(t, http.StatusOK, result.StatusCode)
	})

	t.Run("Write file", func(t *testing.T) {
		request := testutil.NewTestRequest(http.MethodGet, "/gzip", nil)
		request.Header().Set("Accept-Encoding", "gzip")
		result := server.TestMiddleware(compressMiddleware, request, func(r *goyave.Response, _ *goyave.Request) {
			r.File("../../resources/custom_config.json")
		})

		reader, err := gzip.NewReader(result.Body)
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(reader)
		if err != nil {
			panic(err)
		}
		assert.NoError(t, result.Body.Close())
		assert.Equal(t, "gzip", result.Header.Get("Content-Encoding"))
		assert.Empty(t, result.Header.Get("Content-Length"))
		assert.Equal(t, "application/json", result.Header.Get("Content-Type"))
		assert.Equal(t, "{\n    \"custom-entry\": \"value\"\n}", string(body))
	})

}

func TestGzipEncoder(t *testing.T) {
	encoder := &Gzip{
		Level: gzip.BestCompression,
	}
	assert.Equal(t, "gzip", encoder.Encoding())
	assert.Equal(t, gzip.BestCompression, encoder.Level)

	buf := bytes.NewBuffer([]byte{})
	writer := encoder.NewWriter(buf)
	if assert.NotNil(t, writer) {
		_, ok := writer.(*gzip.Writer)
		assert.True(t, ok)
	}

	assert.Panics(t, func() {
		// Invalid level
		encoder := &Gzip{
			Level: -3,
		}
		encoder.NewWriter(bytes.NewBuffer([]byte{}))
	})
}

func TestCompressWriter(t *testing.T) {
	encoder := &Gzip{
		Level: gzip.BestCompression,
	}

	buf := bytes.NewBuffer([]byte{})
	closeableWriter := &closeableChildWriter{
		Writer: buf,
		closed: false,
	}

	response := httptest.NewRecorder()

	writer := &compressWriter{
		WriteCloser:    encoder.NewWriter(closeableWriter),
		ResponseWriter: response,
		childWriter:    closeableWriter,
	}

	writer.PreWrite([]byte("hello world"))

	assert.True(t, closeableWriter.preWritten)

	assert.NoError(t, writer.Close())
	assert.True(t, closeableWriter.closed)
}
