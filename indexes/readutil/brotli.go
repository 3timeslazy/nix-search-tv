package readutil

import (
	"io"

	// "github.com/andybalholm/brotli"
	"github.com/google/brotli/go/cbrotli"
)

type brotliReadCloser struct {
	rd io.ReadCloser
	// brd *brotli.Reader
	brd io.ReadCloser
}

func NewBrotli(rd io.ReadCloser) *brotliReadCloser {
	return &brotliReadCloser{
		rd: rd,
		// brd: brotli.NewReader(rd),
		brd: cbrotli.NewReader(rd),
	}
}

func (br *brotliReadCloser) Close() error {
	// return br.rd.Close()
	return br.brd.Close()
}

func (br *brotliReadCloser) Read(p []byte) (n int, err error) {
	return br.brd.Read(p)
}
