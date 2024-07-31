package compress

import (
	"bytes"
	"compress/zlib"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/testutil"
)

func TestZlibEncoder(t *testing.T) {
	encoder := &Zlib{
		Level: zlib.BestCompression,
		Dict:  nil,
	}

	assert.Equal(t, "deflate", encoder.Encoding())
	assert.Equal(t, zlib.BestCompression, encoder.Level)

	buf := bytes.NewBuffer([]byte{})
	writer := encoder.NewWriter(buf)
	if assert.NotNil(t, writer) {
		_, ok := writer.(*zlib.Writer)
		assert.True(t, ok)
	}

	assert.Panics(t, func() {
		// Invalid level
		encoder := &Zlib{
			Level: -3,
			Dict:  nil,
		}
		encoder.NewWriter(bytes.NewBuffer([]byte{}))
	})
}

func TestZlibCompression(t *testing.T) {
	server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: config.LoadDefault()})

	handler := func(resp *goyave.Response, _ *goyave.Request) {
		resp.Header().Set("Content-Length", "1234")
		resp.String(http.StatusOK, "hello world")
	}

	compressMiddleWare := &Middleware{
		Encoders: []Encoder{
			&Zlib{Level: zlib.BestCompression, Dict: nil},
		},
	}

	request := testutil.NewTestRequest(http.MethodGet, "/zlib", nil)
	request.Header().Set("Accept-Encoding", "deflate")
	result := server.TestMiddleware(compressMiddleWare, request, handler)

	reader, err := zlib.NewReader(result.Body)
	require.NoError(t, err)
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.NoError(t, result.Body.Close())
	assert.Equal(t, "hello world", string(body))
	assert.Equal(t, "deflate", result.Header.Get("Content-Encoding"))
	assert.Empty(t, result.Header.Get("Content-Length"))
}

func TestZlibCompressionNoDict(t *testing.T) {
	server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: config.LoadDefault()})

	handler := func(resp *goyave.Response, _ *goyave.Request) {
		resp.Header().Set("Content-Length", "1234")
		resp.String(http.StatusOK, "hello world")
	}

	compressMiddleWare := &Middleware{
		Encoders: []Encoder{
			&Zlib{Level: zlib.BestCompression},
		},
	}

	request := testutil.NewTestRequest(http.MethodGet, "/zlib", nil)
	request.Header().Set("Accept-Encoding", "deflate")
	result := server.TestMiddleware(compressMiddleWare, request, handler)

	reader, err := zlib.NewReader(result.Body)
	require.NoError(t, err)
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.NoError(t, result.Body.Close())
	assert.Equal(t, "hello world", string(body))
	assert.Equal(t, "deflate", result.Header.Get("Content-Encoding"))
	assert.Empty(t, result.Header.Get("Content-Length"))
}
