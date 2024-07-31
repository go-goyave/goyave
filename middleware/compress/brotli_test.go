package compress

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/util/testutil"
)

func TestBrotliEncoder(t *testing.T) {
	encoder := &Brotli{
		Quality: 6,
	}

	assert.Equal(t, "br", encoder.Encoding())

	buf := bytes.NewBuffer([]byte{})
	writer := encoder.NewWriter(buf)
	if assert.NotNil(t, writer) {
		_, ok := writer.(*brotli.Writer)
		assert.True(t, ok)
	}

	assert.Panics(t, func() {
		// Invalid Quality
		encoder := &Brotli{
			Quality: 12,
		}
		encoder.NewWriter(bytes.NewBuffer([]byte{}))
	})

	assert.Panics(t, func() {
		// Invalid LGWin
		encoder := &Brotli{
			Quality: 11,
			LGWin:   9,
		}
		encoder.NewWriter(bytes.NewBuffer([]byte{}))
	})
}

func TestBrotliCompression(t *testing.T) {
	server := testutil.NewTestServerWithOptions(t, goyave.Options{Config: config.LoadDefault()})

	handler := func(resp *goyave.Response, _ *goyave.Request) {
		resp.Header().Set("Content-Length", "1234")
		resp.String(http.StatusOK, "hello world")
	}

	compressMiddleWare := &Middleware{
		Encoders: []Encoder{
			&Brotli{
				Quality: 6,
				LGWin:   15,
			},
		},
	}

	request := testutil.NewTestRequest(http.MethodGet, "/brotli", nil)
	request.Header().Set("Accept-Encoding", "br")
	result := server.TestMiddleware(compressMiddleWare, request, handler)

	reader := brotli.NewReader(result.Body)
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.NoError(t, result.Body.Close())
	assert.Equal(t, "hello world", string(body))
	assert.Equal(t, "br", result.Header.Get("Content-Encoding"))
	assert.Empty(t, result.Header.Get("Content-Length"))
}
