package zstd

import (
	_ "embed"

	"github.com/valyala/gozstd"
)

var (
	//go:embed zstd.dict
	dict []byte

	cdict = must(gozstd.NewCDict(dict))
	ddict = must(gozstd.NewDDict(dict))
)

func Compress(dst, src []byte) []byte {
	return gozstd.CompressDict(dst, src, cdict)
}

func Decompress(dst, src []byte) ([]byte, error) {
	return gozstd.DecompressDict(dst, src, ddict)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
