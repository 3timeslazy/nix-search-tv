package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Package defines fields set by the indexer during
// indexing
type Package struct {
	Name string `json:"_key"`
}

// Indexable represents the internal structure of the data
// that the indexer expects.
type Indexable struct {
	Packages map[string]json.RawMessage `json:"packages"`
}

type Fetcher interface {
	GetLatestRelease(context.Context, IndexMetadata) (string, error)
	DownloadRelease(context.Context, string) (io.ReadCloser, error)
}

type Index struct {
	Name     string
	Fetcher  Fetcher
	Metadata IndexMetadata
}

type IndexMetadata struct {
	LastIndexedAt time.Time `json:"last_indexed_at"`
	CurrRelease   string    `json:"curr_release"`
}

type IndexingResult struct {
	Index string
	Err   error
}

func RunIndexing(
	ctx context.Context,
	cacheDir string,
	indexes []Index,
) <-chan IndexingResult {
	results := make(chan IndexingResult)

	wg := sync.WaitGroup{}
	wg.Add(len(indexes))

	for _, index := range indexes {
		go func() {
			defer wg.Done()

			var err error
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("index panicked: %v", r)
				}

				results <- IndexingResult{index.Name, err}
			}()

			err = runIndex(ctx, cacheDir, index)
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func runIndex(
	ctx context.Context,
	cacheDir string,
	index Index,
) error {
	indexDir := filepath.Join(cacheDir, index.Name)
	latest, err := index.Fetcher.GetLatestRelease(ctx, index.Metadata)
	if err != nil {
		return fmt.Errorf("get latest release: %w", err)
	}
	if latest == index.Metadata.CurrRelease {
		_ = setIndexMetadata(indexDir, IndexMetadata{
			LastIndexedAt: time.Now(),
			CurrRelease:   latest,
		})
		return nil
	}

	pkgs, err := index.Fetcher.DownloadRelease(ctx, latest)
	if err != nil {
		return fmt.Errorf("download latest release: %w", err)
	}
	defer pkgs.Close()

	cache, err := CacheWriter(indexDir)
	if err != nil {
		return fmt.Errorf("open cache write: %w", err)
	}
	defer cache.Close()

	simpleDir := filepath.Join(indexDir, "simple")
	indexer, err := NewSimple(simpleDir)
	if err != nil {
		return fmt.Errorf("open indexer: %w", err)
	}
	defer indexer.Close()

	// badgerDir := filepath.Join(indexDir, "badger")
	// indexer, err := NewBadger(BadgerConfig{
	// 	Dir: badgerDir,
	// })
	// if err != nil {
	// 	return fmt.Errorf("open indexer: %w", err)
	// }
	// defer indexer.Close()

	err = indexer.Index(pkgs, cache)
	if err != nil {
		return fmt.Errorf("index packages: %w", err)
	}

	_ = setIndexMetadata(indexDir, IndexMetadata{
		LastIndexedAt: time.Now(),
		CurrRelease:   latest,
	})

	return nil
}

type OptionFileFetcher interface {
	Path() string
}

func NeedIndexing(
	cacheDir string,
	updateInterval time.Duration,
	indexes []Index,
) ([]Index, error) {
	needIndex := []Index{}

	for _, index := range indexes {
		if file, ok := index.Fetcher.(OptionFileFetcher); ok {
			path := file.Path()
			if path != index.Metadata.CurrRelease {
				needIndex = append(needIndex, index)
			}

			continue
		}

		if time.Since(index.Metadata.LastIndexedAt) > time.Duration(updateInterval) {
			needIndex = append(needIndex, index)
		}
	}

	return needIndex, nil
}

func OpenKeysReader(cacheDir, index string) (io.ReadCloser, error) {
	indexDir := filepath.Join(cacheDir, index)
	path, err := initFile(indexDir, cacheFile, nil)
	if err != nil {
		return nil, fmt.Errorf("init cache file: %w", err)
	}

	return os.OpenFile(path, os.O_RDONLY, 0666)
}

func LoadKey(cacheDir, index, key string) (json.RawMessage, error) {
	simpleDir := filepath.Join(cacheDir, index, "simple")
	indexer, err := NewSimple(simpleDir)
	if err != nil {
		return nil, fmt.Errorf("open indexer: %w", err)
	}
	defer indexer.Close()

	// badgerDir := filepath.Join(cacheDir, index, "badger")
	// indexer, err := NewBadger(BadgerConfig{
	// 	Dir: badgerDir,
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("open indexer: %w", err)
	// }
	// defer indexer.Close()

	data, err := indexer.Load(key)
	if err != nil {
		return nil, fmt.Errorf("load key: %w", err)
	}

	return data, nil
}
