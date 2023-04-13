package compress

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/samber/lo"
	"goyave.dev/goyave/v4"
	"goyave.dev/goyave/v4/util/httputil"
)

// Encoder is an interface that wraps the methods returning the information
// necessary for the compress middleware to work.
//
// `NewWriter` returns any `io.WriteCloser`, allowing the middleware to support
// any compression algorithm.
//
// `Encoding` returns the name of the compression algorithm. Using the returned value,
// the middleware:
//  1. detects the client's preferred encoding with the `Accept-Encoding` request header
//  2. replaces the response writer with the writer returned by `NewWriter`
//  3. sets the `Content-Encoding` response header
type Encoder interface {
	NewWriter(io.Writer) io.WriteCloser
	Encoding() string
}

type compressWriter struct {
	io.WriteCloser
	http.ResponseWriter
	childWriter io.Writer
}

func (w *compressWriter) PreWrite(b []byte) {
	if pr, ok := w.childWriter.(goyave.PreWriter); ok {
		pr.PreWrite(b)
	}
	h := w.ResponseWriter.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(b))
	}
	h.Del("Content-Length")
}

func (w *compressWriter) Write(b []byte) (int, error) {
	return w.WriteCloser.Write(b)
}

func (w *compressWriter) Close() error {
	err := w.WriteCloser.Close()

	if wr, ok := w.childWriter.(io.Closer); ok {
		return wr.Close()
	}

	return err
}

// Gzip encoder for the gzip format using Go's standard `compress/gzip` package.
//
// Takes a compression level as parameter. Accepted values are defined by constants
// in the standard `compress/gzip` package.
type Gzip struct {
	Level int
}

// Encoding returns "gzip".
func (w *Gzip) Encoding() string {
	return "gzip"
}

// NewWriter returns a new `compress/gzip.Writer` using the compression level
// defined in this Gzip encoder.
func (w *Gzip) NewWriter(wr io.Writer) io.WriteCloser {
	writer, err := gzip.NewWriterLevel(wr, w.Level)
	if err != nil {
		panic(err)
	}
	return writer
}

// Middleware compresses HTTP responses.
//
// This middleware supports multiple algorithms thanks to the `Encoders` slice.
// The encoder will be chosen depending on the request's `Accept-Encoding` header,
// and the value returned by the `Encoder`'s `Encoding()` method. Quality values in
// the headers are taken into account.
//
// If the header's value is `*`, the first element of the slice is used.
// If none of the accepted encodings are available in the `Encoders` slice, then the
// response will not be compressed and the middleware immediately passes.
//
// If the middleware successfully replaces the response writer, the `Accept-Encoding`
// header is removed from the request to avoid potential clashes with potential other
// encoding middleware.
//
// If not set at the first call of `Write()`, the middleware will automatically detect
// and set the `Content-Type` header using `http.DetectContentType()`.
//
// The middleware ignores hijacked responses or requests containing the `Upgrade` header.
//
// **Example:**
//
//	compressMiddleware := &compress.Middleware{
//		Encoders: []compress.Encoder{
//			&compress.Gzip{Level: gzip.BestCompression},
//		},
//	}
type Middleware struct {
	goyave.Component
	Encoders []Encoder
}

// Handle implementation of `goyave.Middleware`.
func (m *Middleware) Handle(next goyave.HandlerV5) goyave.HandlerV5 {
	return func(response *goyave.ResponseV5, request *goyave.RequestV5) {
		encoder := m.getEncoder(response, request)
		if encoder == nil {
			next(response, request)
			return
		}

		request.Header().Del("Accept-Encoding")

		respWriter := response.Writer()
		compressWriter := &compressWriter{
			WriteCloser:    encoder.NewWriter(respWriter),
			ResponseWriter: response,
			childWriter:    respWriter,
		}
		response.SetWriter(compressWriter)
		response.Header().Set("Content-Encoding", encoder.Encoding())

		next(response, request)
	}
}

func (m *Middleware) getEncoder(response *goyave.ResponseV5, request *goyave.RequestV5) Encoder {
	if response.Hijacked() || request.Header().Get("Upgrade") != "" {
		return nil
	}
	encodings := httputil.ParseMultiValuesHeader(request.Header().Get("Accept-Encoding"))
	for _, h := range encodings {
		if h.Value == "*" {
			return m.Encoders[0]
		}
		w, ok := lo.Find(m.Encoders, func(w Encoder) bool {
			return w.Encoding() == h.Value
		})
		if ok {
			return w
		}
	}

	return nil
}
