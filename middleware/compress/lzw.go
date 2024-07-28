package compress

import (
	"compress/lzw"
	"io"

	"goyave.dev/goyave/v5/util/errors"
)

// Lzw encoder for the gzip format using Go's standard `compress/lzw` package.
//
// Takes an Order specifying the bit ordering in an LZW data stream level,
// and number of bits to use for literal codes, litWidth, as parameters.
// Accepted values are defined by constants
// in the standard `compress/lzw` package.
type Lzw struct {
	Order    lzw.Order
	LitWidth int
}

// Encoding returns "lzw".
func (w *Lzw) Encoding() string {
	return "lzw"
}

// NewWriter returns a new `compress/lzw.Writer` using an LZW bit ordering type,
// and an int for number of bits to use for literal codes - must be within range [2, 8]
func (w *Lzw) NewWriter(wr io.Writer) io.WriteCloser {
	if w.LitWidth < 2 || w.LitWidth > 8 {
		panic(errors.New("LitWidth must be in range [2, 8]"))
	}
	return lzw.NewWriter(wr, w.Order, w.LitWidth)
}
