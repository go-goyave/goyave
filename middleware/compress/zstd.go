package compress

import (
	"io"

	"github.com/klauspost/compress/zstd"
	"goyave.dev/goyave/v5/util/errwrap"
)

// Zstd encoder for the Zstandard compression algorithm
// You may provide a list of zstd.EOptions as parameters.
// Refer to the package documentation for more information
type Zstd struct {
	Options []zstd.EOption
}

// Encoding returns "zstd".
func (w *Zstd) Encoding() string {
	return "zstd"
}

// NewWriter returns a new `zstd.Encoder` using the zstd.EOptions
// defined in the Zstd encoder.
func (w *Zstd) NewWriter(wr io.Writer) io.WriteCloser {
	writer, err := zstd.NewWriter(wr, w.Options...)
	if err != nil {
		panic(errwrap.New(err))
	}
	return writer
}
