package indexer

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand/v2"
	"os"
	"strings"
	"testing"

	"github.com/3timeslazy/nix-search-tv/indexes/readutil"
	"github.com/alecthomas/assert/v2"
)

type Indexer interface {
	Index(io.Reader, io.Writer) error
	Load(string) (json.RawMessage, error)
	Close() error
}

func TestBadgerSetGet(t *testing.T) {
	testSetGet(t, func(dir string) (Indexer, error) {
		return NewBadger(BadgerConfig{
			Dir:      dir,
			InMemory: false,
		})
	})
}

func TestSimpleSetGet(t *testing.T) {
	testSetGet(t, func(dir string) (Indexer, error) {
		return NewSimple(dir)
	})
}

func testSetGet(t *testing.T, newIndexer func(dir string) (Indexer, error)) {
	t.Parallel()

	data, err := os.ReadFile("./testdata/packages.json.br")
	assert.NoError(t, err)

	tmpdir, err := os.MkdirTemp("", "nix-search-tv-indexer-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	indexer, err := newIndexer(tmpdir)
	assert.NoError(t, err)
	defer indexer.Close()

	brd := readutil.NewBrotli(io.NopCloser(bytes.NewReader(data)))

	keysbuf := bytes.NewBuffer(nil)
	err = indexer.Index(brd, keysbuf)
	assert.NoError(t, err)

	keys := strings.Split(keysbuf.String(), "\n")
	key := keys[rand.IntN(len(keys))]

	pkg, err := indexer.Load(key)
	assert.NoError(t, err)

	assert.True(t, json.Valid(pkg))
}

func BenchmarkBadger(b *testing.B) {
	benchmark(b, func(dir string) (Indexer, error) {
		return NewBadger(BadgerConfig{
			Dir:      dir,
			InMemory: false,
		})
	})
}

func BenchmarkSimple(b *testing.B) {
	benchmark(b, func(dir string) (Indexer, error) {
		return NewSimple(dir)
	})
}

func benchmark(b *testing.B, newIndexer func(dir string) (Indexer, error)) {
	data, err := os.ReadFile("./testdata/packages.json.br")
	assert.NoError(b, err)

	tmpdir, err := os.MkdirTemp("", "nix-search-tv-indexer-*")
	assert.NoError(b, err)
	defer os.RemoveAll(tmpdir)

	indexer, err := newIndexer(tmpdir)
	assert.NoError(b, err)
	defer indexer.Close()

	pkgs := readutil.NewBrotli(io.NopCloser(bytes.NewReader(data)))

	for b.Loop() {
		err = indexer.Index(pkgs, io.Discard)
		assert.NoError(b, err)
		pkgs.Reset(io.NopCloser(bytes.NewReader(data)))
	}
}
