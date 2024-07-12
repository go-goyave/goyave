package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
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
	server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: config.LoadDefault()})

	handler := func(resp *goyave.Response, _ *goyave.Request) {
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
		assert.Equal(t, "1234", result.Header.Get("Content-Length"))
		assert.Equal(t, http.StatusOK, result.StatusCode)
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
		assert.Equal(t, "1234", result.Header.Get("Content-Length"))
		assert.Equal(t, http.StatusOK, result.StatusCode)
	})

	t.Run("Write file", func(t *testing.T) {
		request := testutil.NewTestRequest(http.MethodGet, "/gzip", nil)
		request.Header().Set("Accept-Encoding", "gzip")
		result := server.TestMiddleware(compressMiddleware, request, func(r *goyave.Response, _ *goyave.Request) {
			r.File(&osfs.FS{}, "../../resources/custom_config.json")
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
		CommonWriter:   goyave.NewCommonWriter(encoder.NewWriter(closeableWriter)),
		responseWriter: response,
		childWriter:    closeableWriter,
	}

	writer.PreWrite([]byte("hello world"))

	assert.True(t, closeableWriter.preWritten)

	require.NoError(t, writer.Close())
	assert.True(t, closeableWriter.closed)

	// TODO flush test
}

type testEncoder struct {
	encoding string
}

func (e *testEncoder) NewWriter(_ io.Writer) io.WriteCloser {
	return nil
}

func (e *testEncoder) Encoding() string {
	return e.encoding
}

func TestEncoderPriority(t *testing.T) {
	gzip := &testEncoder{encoding: "gzip"}
	br := &testEncoder{encoding: "br"}
	zstd := &testEncoder{encoding: "zstd"}

	cases := []struct {
		want           Encoder
		acceptEncoding string
		encoders       []Encoder
	}{
		{
			encoders:       []Encoder{br, zstd, gzip},
			acceptEncoding: "gzip, deflate, br, zstd",
			want:           br,
		},
		{
			encoders:       []Encoder{br, zstd, gzip},
			acceptEncoding: "*",
			want:           br,
		},
		{
			encoders:       []Encoder{br, zstd, gzip},
			acceptEncoding: "gzip, *",
			want:           gzip,
		},
		{
			encoders:       []Encoder{br, zstd, gzip},
			acceptEncoding: "gzip, br, *",
			want:           br,
		},
		{
			encoders:       []Encoder{br, zstd, gzip},
			acceptEncoding: "*, gzip, br",
			want:           br,
		},
		{
			encoders:       []Encoder{br, zstd, gzip},
			acceptEncoding: "gzip, *;q=0.9",
			want:           gzip,
		},
		{
			encoders:       []Encoder{br, zstd, gzip},
			acceptEncoding: "gzip",
			want:           gzip,
		},
		{
			encoders:       []Encoder{br, zstd, gzip},
			acceptEncoding: "zstd;q=0.9, br;q=0.9",
			want:           br,
		},
		{
			encoders:       []Encoder{br, zstd, gzip},
			acceptEncoding: "zstd;q=0.9, br;q=0.8",
			want:           zstd,
		},
		{
			encoders:       []Encoder{br, zstd, gzip},
			acceptEncoding: "gzip;q=0.8, *;q=0.1",
			want:           gzip,
		},
		{
			encoders:       []Encoder{br, zstd, gzip},
			acceptEncoding: "gzip;q=0.8, *;q=1.0",
			want:           br,
		},
		{
			encoders:       []Encoder{br, zstd, gzip},
			acceptEncoding: "",
			want:           nil,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.acceptEncoding, func(t *testing.T) {
			middleware := &Middleware{
				Encoders: c.encoders,
			}
			request := testutil.NewTestRequest(http.MethodGet, "/", nil)
			request.Header().Set("Accept-Encoding", c.acceptEncoding)
			response, _ := testutil.NewTestResponse(request)
			e := middleware.getEncoder(response, request)
			assert.Equal(t, c.want, e)
		})
	}
}
