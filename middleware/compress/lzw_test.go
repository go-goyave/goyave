package compress

import (
	"bytes"
	"compress/lzw"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/testutil"
)

func TestLzwEncoder(t *testing.T) {
	encoder := &Lzw{
		Order:    lzw.LSB,
		LitWidth: 5,
	}

	assert.Equal(t, "lzw", encoder.Encoding())

	buf := bytes.NewBuffer([]byte{})
	writer := encoder.NewWriter(buf)
	if assert.NotNil(t, writer) {
		_, ok := writer.(*lzw.Writer)
		assert.True(t, ok)
	}

	assert.Panics(t, func() {
		// Invalid LitWidth
		encoder := &Lzw{
			Order:    lzw.LSB,
			LitWidth: 9,
		}
		encoder.NewWriter(bytes.NewBuffer([]byte{}))
	})
}

func TestLzwCompression(t *testing.T) {
	server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: config.LoadDefault()})

	handler := func(resp *goyave.Response, _ *goyave.Request) {
		resp.Header().Set("Content-Length", "1234")
		resp.String(http.StatusOK, "hello world")
	}

	compressMiddleWare := &Middleware{
		Encoders: []Encoder{
			&Lzw{
				Order:    lzw.LSB,
				LitWidth: 8,
			},
		},
	}

	request := testutil.NewTestRequest(http.MethodGet, "/lzw", nil)
	request.Header().Set("Accept-Encoding", "lzw")
	result := server.TestMiddleware(compressMiddleWare, request, handler)

	reader := lzw.NewReader(result.Body, lzw.LSB, 8)
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.NoError(t, result.Body.Close())
	assert.Equal(t, "hello world", string(body))
	assert.Equal(t, "lzw", result.Header.Get("Content-Encoding"))
	assert.Empty(t, result.Header.Get("Content-Length"))
}
