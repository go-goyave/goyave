package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"github.com/System-Glitch/goyave/v3"
	"github.com/System-Glitch/goyave/v3/helper"
)

type gzipWriter struct {
	*gzip.Writer
	http.ResponseWriter
	childWriter io.Closer
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	h := w.ResponseWriter.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(b))
	}
	h.Del("Content-Length")

	return w.Writer.Write(b)
}

func (w *gzipWriter) Close() error {
	err := w.Writer.Close()

	if w.childWriter != nil {
		return w.childWriter.Close()
	}

	return err
}

// Gzip compresses HTTP responses with default compression level
// for clients that support it via the 'Accept-Encoding' header.
func Gzip() goyave.Middleware {
	return GzipLevel(gzip.DefaultCompression)
}

// GzipLevel compresses HTTP responses with specified compression level
// for clients that support it via the 'Accept-Encoding' header.
//
// The compression level should be gzip.DefaultCompression, gzip.NoCompression,
// or any integer value between gzip.BestSpeed and gzip.BestCompression inclusive.
func GzipLevel(level int) goyave.Middleware {
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		panic(fmt.Errorf("gzip: invalid compression level: %d", level))
	}
	return func(next goyave.Handler) goyave.Handler {
		return func(response *goyave.Response, request *goyave.Request) {
			if !acceptsGzip(request) {
				next(response, request)
				return
			}

			request.Header().Del("Accept-Encoding")

			respWriter := response.Writer()
			writer, _ := gzip.NewWriterLevel(respWriter, level)
			compressWriter := &gzipWriter{
				Writer:         writer,
				ResponseWriter: response,
			}
			if w, ok := respWriter.(io.Closer); ok {
				compressWriter.childWriter = w
			}
			response.SetWriter(compressWriter)
			response.Header().Set("Content-Encoding", "gzip")

			next(response, request)

			response.Header().Del("Content-Length")
		}
	}
}

func acceptsGzip(request *goyave.Request) bool {
	encodings := helper.ParseMultiValuesHeader(request.Header().Get("Accept-Encoding"))
	for _, h := range encodings {
		if h.Value == "gzip" {
			return true
		}
	}

	return false
}
