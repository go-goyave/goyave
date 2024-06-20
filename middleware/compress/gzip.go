package compress

import (
	"compress/gzip"
	"io"

	"goyave.dev/goyave/v5/util/errors"
)

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
