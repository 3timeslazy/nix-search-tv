package indexer

import (
	_ "embed"

	"github.com/valyala/gozstd"
)

var (
	//go:embed gozstd.dict
	dict []byte

	cdict = must(gozstd.NewCDict(dict))
	ddict = must(gozstd.NewDDict(dict))
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
