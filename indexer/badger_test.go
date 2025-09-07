package indexer

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/3timeslazy/nix-search-tv/indexes/readutil"
	"github.com/alecthomas/assert/v2"
)

func BenchmarkBadger(b *testing.B) {
	data, err := os.ReadFile("./testdata/packages.json.br")
	assert.NoError(b, err)

	tmpdir, err := os.MkdirTemp("", "nix-search-tv-badger-*")
	assert.NoError(b, err)
	defer os.RemoveAll(tmpdir)

	badger, err := NewBadger(BadgerConfig{
		Dir:      tmpdir,
		InMemory: false,
	})
	assert.NoError(b, err)
	defer badger.Close()

	brd := readutil.NewBrotli(io.NopCloser(bytes.NewReader(data)))

	// for range b.N {
	t := time.Now()

	err = badger.Index(brd, io.Discard)
	assert.NoError(b, err)

	fmt.Println(time.Since(t))
	// }
}
