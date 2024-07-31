package compress

import (
	"compress/lzw"
	"io"

	"goyave.dev/goyave/v5/util/errors"
)

// LZW encoder for the compress format using Go's standard `compress/lzw` package.
//
// Takes an Order specifying the bit ordering in an LZW data stream level,
// and number of bits to use for literal codes, litWidth, as parameters.
// Accepted values are defined by constants
// in the standard `compress/lzw` package.
type LZW struct {
	// Order specifies the bit ordering in an LZW data stream.
	// It is optional, and the default value is lzw.LSB (Least Significant Bits)
	Order lzw.Order
	// LitWidth specifies the number of bits to use for literal codes
	// Must be in the range [2,8] and is typically 8.
	// Input bytes must be less than 1<<litWidth.
	LitWidth int
}

// Encoding returns "compress".
func (w *LZW) Encoding() string {
	return "compress"
}

// NewWriter returns a new `compress/lzw.Writer` using an LZW bit ordering type,
// and an int for number of bits to use for literal codes - must be within range [2, 8]
// Default to a LitWidth of 8 if LitWidth was not set
func (w *LZW) NewWriter(wr io.Writer) io.WriteCloser {
	if w.LitWidth == 0 {
		w.LitWidth = 8
	}
	if w.LitWidth < 2 || w.LitWidth > 8 {
		panic(errors.New("LitWidth must be in range [2, 8]"))
	}
	return lzw.NewWriter(wr, w.Order, w.LitWidth)
}
