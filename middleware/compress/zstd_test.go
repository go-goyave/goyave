package compress

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/testutil"
)

func TestZstdEncoder(t *testing.T) {
	encoder := &Zstd{
		Options: nil,
	}

	assert.Equal(t, "zstd", encoder.Encoding())

	buf := bytes.NewBuffer([]byte{})
	writer := encoder.NewWriter(buf)
	require.NotNil(t, writer)
	_, ok := writer.(*zstd.Encoder)
	assert.True(t, ok)

	assert.Panics(t, func() {
		// Invalid EncoderConcurrency
		encoder := &Zstd{
			Options: []zstd.EOption{
				zstd.WithEncoderConcurrency(0),
			},
		}
		encoder.NewWriter(bytes.NewBuffer([]byte{}))
	})
}

func TestZstdCompression(t *testing.T) {
	server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: config.LoadDefault()})

	handler := func(resp *goyave.Response, _ *goyave.Request) {
		resp.Header().Set("Content-Length", "1234")
		resp.String(http.StatusOK, "hello world")
	}

	compressMiddleWare := &Middleware{
		Encoders: []Encoder{
			&Zstd{
				Options: []zstd.EOption{
					zstd.WithEncoderConcurrency(1),
					zstd.WithEncoderCRC(true),
				},
			},
		},
	}

	request := testutil.NewTestRequest(http.MethodGet, "/zstd", nil)
	request.Header().Set("Accept-Encoding", "zstd")
	result := server.TestMiddleware(compressMiddleWare, request, handler)

	reader, err := zstd.NewReader(result.Body)
	require.NoError(t, err)

	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.NoError(t, result.Body.Close())
	assert.Equal(t, "hello world", string(body))
	assert.Equal(t, "zstd", result.Header.Get("Content-Encoding"))
	assert.Empty(t, result.Header.Get("Content-Length"))
}
