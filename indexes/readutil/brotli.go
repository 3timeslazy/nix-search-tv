package readutil

import (
	"io"

	"github.com/andybalholm/brotli"
)

type brotliReadCloser struct {
	rd  io.ReadCloser
	brd *brotli.Reader
}

func NewBrotli(rd io.ReadCloser) *brotliReadCloser {
	return &brotliReadCloser{
		rd:  rd,
		brd: brotli.NewReader(rd),
	}
}

func (br *brotliReadCloser) Close() error {
	return br.rd.Close()
}

func (br *brotliReadCloser) Read(p []byte) (n int, err error) {
	return br.brd.Read(p)
}

// Reset resets the internal brotli reader. Used for testing only.
func (br *brotliReadCloser) Reset(rd io.Reader) {
	br.brd.Reset(rd)
}
