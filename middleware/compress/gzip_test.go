package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/testutil"
)

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

func TestGzipCompression(t *testing.T) {
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

	request := testutil.NewTestRequest(http.MethodGet, "/gzip", nil)
	request.Header().Set("Accept-Encoding", "gzip")
	result := server.TestMiddleware(compressMiddleware, request, handler)

	reader, err := gzip.NewReader(result.Body)
	require.NoError(t, err)
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.NoError(t, result.Body.Close())
	assert.Equal(t, "hello world", string(body))
	assert.Equal(t, "gzip", result.Header.Get("Content-Encoding"))
	assert.Empty(t, result.Header.Get("Content-Length"))
}
