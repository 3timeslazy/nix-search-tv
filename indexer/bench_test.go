package indexer_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/indexes/darwin"
	"github.com/3timeslazy/nix-search-tv/indexes/homemanager"
	"github.com/3timeslazy/nix-search-tv/indexes/readutil"
)

// Indexer is the interface that alternative implementations
// must satisfy to be benchmarked.
type Indexer interface {
	Index(data io.Reader, indexedKeys io.Writer) error
	Close() error
}

// NewIndexerFunc creates a new Indexer backed by dir on disk.
type NewIndexerFunc func(dir string) (Indexer, error)

// func newBadger(dir string) (Indexer, error) {
// 	return indexer.NewBadger(indexer.BadgerConfig{
// 		Dir: dir,
// 	})
// }

func newSimple(dir string) (Indexer, error) {
	return indexer.NewSimple(dir)
}

type indexSource struct {
	name string
	open func() (io.ReadCloser, error)
}

func defaultSources() []indexSource {
	return []indexSource{
		{
			name: "nixpkgs",
			open: func() (io.ReadCloser, error) {
				pkgs, err := os.Open("../indexes/nixpkgs/testdata/packages.json.br")
				if err != nil {
					return nil, err
				}
				return readutil.NewBrotli(pkgs), nil
			},
		},
		{
			name: "nixos",
			open: func() (io.ReadCloser, error) {
				pkgs, err := os.Open("../indexes/nixos/testdata/options.br.json")
				if err != nil {
					return nil, err
				}
				return readutil.PackagesWrapper(readutil.NewBrotli(pkgs)), nil
			},
		},
		{
			name: "home-manager",
			open: func() (io.ReadCloser, error) {
				return homemanager.Fetcher{}.DownloadRelease(
					context.Background(),
					"file://../indexes/homemanager/testdata/options.xhtml",
				)
			},
		},
		{
			name: "darwin",
			open: func() (io.ReadCloser, error) {
				return darwin.Fetcher{}.DownloadRelease(
					context.Background(),
					"file://../indexes/darwin/testdata/index.html",
				)
			},
		},
	}
}

// func BenchmarkBadgerIndexNixpkgs(b *testing.B) {
// 	benchmarkIndexNixpkgs(b, newBadger)
// }

// func BenchmarkBadgerIndexDefaultIndexes(b *testing.B) {
// 	benchmarkIndexDefaultIndexes(b, newBadger)
// }

func BenchmarkSimpleIndexNixpkgs(b *testing.B) {
	benchmarkIndexNixpkgs(b, newSimple)
}

func BenchmarkSimpleIndexDefaultIndexes(b *testing.B) {
	benchmarkIndexDefaultIndexes(b, newSimple)
}

func benchmarkIndexNixpkgs(b *testing.B, newIndexer NewIndexerFunc) {
	b.Helper()

	for range b.N {
		b.StopTimer()

		dir := b.TempDir()

		pkgs, err := os.Open("../indexes/nixpkgs/testdata/packages.json.br")
		if err != nil {
			b.Fatal(err)
		}
		pkgsbr := readutil.NewBrotli(pkgs)

		b.StartTimer()

		idx, err := newIndexer(dir)
		if err != nil {
			b.Fatal(err)
		}

		err = idx.Index(pkgsbr, io.Discard)
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()

		pkgsbr.Close()
		idx.Close()
	}
}

func benchmarkIndexDefaultIndexes(b *testing.B, newIndexer NewIndexerFunc) {
	b.Helper()

	sources := defaultSources()

	for range b.N {
		b.StopTimer()

		readers := make([]io.ReadCloser, len(sources))
		for i, src := range sources {
			rdr, err := src.open()
			if err != nil {
				b.Fatalf("open %s: %v", src.name, err)
			}
			readers[i] = rdr
		}

		dir := b.TempDir()

		idx, err := newIndexer(dir)
		if err != nil {
			b.Fatal(err)
		}

		b.StartTimer()

		for i, src := range sources {
			err := idx.Index(readers[i], io.Discard)
			if err != nil {
				b.Fatalf("index %s: %v", src.name, err)
			}
		}

		b.StopTimer()

		for _, rdr := range readers {
			rdr.Close()
		}
		idx.Close()
	}
}
