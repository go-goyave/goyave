package compress

import (
	"io"

	"github.com/andybalholm/brotli"
	"goyave.dev/goyave/v5/util/errors"
)

// Brotli encoder for the br compression format
type Brotli struct {
	// Quality controls the compression-speed vs compression-density trade-offs.
	// The higher the quality, the slower the compression. Range is 0 to 11.
	Quality int
	// LGWin is the base 2 logarithm of the sliding window size.
	// Range is 10 to 24. 0 indicates automatic configuration based on Quality.
	LGWin int
}

// Encoding returns "br".
func (w *Brotli) Encoding() string {
	return "br"
}

// NewWriter returns a new `brotli.Writer` using the
// Compression Quality and LGWin provided in the Brotli encoder
func (w *Brotli) NewWriter(wr io.Writer) io.WriteCloser {
	if w.Quality < brotli.BestSpeed || w.Quality > brotli.BestCompression {
		panic(errors.New("Brotli Compression Level must be in range [0, 11]"))
	}
	if w.LGWin != 0 {
		if w.LGWin < 10 || w.LGWin > 24 {
			panic(errors.New("Brotli LGWin must be either 0 or within range [10, 24]"))
		}
	}
	return brotli.NewWriterOptions(wr, brotli.WriterOptions{
		Quality: w.Quality,
		LGWin:   w.LGWin,
	})
}
