package compress

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/samber/lo"
	"goyave.dev/goyave/v5"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/httputil"
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
	n, err := w.WriteCloser.Write(b)
	return n, errors.New(err)
}

func (w *compressWriter) Close() error {
	err := errors.New(w.WriteCloser.Close())

	if wr, ok := w.childWriter.(io.Closer); ok {
		return errors.New(wr.Close())
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
		panic(errors.New(err))
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
// In case of equal priority, the encoding that is the earliest in the slice is chosen.
// If the header's value is `*` and no encoding already matched,
// the first element of the slice is used.
//
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
func (m *Middleware) Handle(next goyave.Handler) goyave.Handler {
	return func(response *goyave.Response, request *goyave.Request) {
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

func (m *Middleware) getEncoder(response *goyave.Response, request *goyave.Request) Encoder {
	if response.Hijacked() || request.Header().Get("Upgrade") != "" {
		return nil
	}
	acceptedEncodings := httputil.ParseMultiValuesHeader(request.Header().Get("Accept-Encoding"))
	groupedByPriority := lo.PartitionBy(acceptedEncodings, func(h httputil.HeaderValue) float64 {
		return h.Priority
	})
	for _, h := range groupedByPriority {
		w, ok := lo.Find(m.Encoders, func(w Encoder) bool {
			return lo.ContainsBy(h, func(h httputil.HeaderValue) bool { return h.Value == w.Encoding() })
		})
		if ok {
			return w
		}

		hasWildCard := lo.ContainsBy(h, func(h httputil.HeaderValue) bool { return h.Value == "*" })
		if hasWildCard {
			return m.Encoders[0]
		}
	}

	return nil
}
