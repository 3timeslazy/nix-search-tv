package nuschtos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/indexer/x/jsonstream"
)

type Fetcher struct {
	url string
}

func NewFetcher(url string) *Fetcher {
	return &Fetcher{
		url: url,
	}
}

type Package struct {
	Declarations []string `json:"declarations"`
	Default      string   `json:"default"`
	Description  string   `json:"description"`
	Example      string   `json:"example"`
	ReadOnly     bool     `json:"read_only"`
	Type         string   `json:"type"`
	Name         string   `json:"name"`
}

func (f *Fetcher) GetLatestRelease(_ context.Context, _ indexer.IndexMetadata) (string, error) {
	return time.Now().String(), nil
}

func (f *Fetcher) DownloadRelease(_ context.Context, _ string) (io.ReadCloser, error) {
	pkgs := indexer.Indexable{
		Packages: map[string]json.RawMessage{},
	}

	// 50 should be enough for small things,
	// for big things, use native nixpkgs index
	for i := 0; i < 50; i++ {
		chunk, err := f.downloadChunk(i)
		if errors.Is(err, ErrNotFound) {
			break
		}
		if err != nil {
			panic(err)
		}

		for _, pkg := range chunk {
			pkgName, err := jsonstream.FindPath(bytes.NewReader(pkg), "name")
			if err != nil {
				panic(err)
			}
			if pkgName == "" {
				panic(string(pkg))
			}

			pkgs.Packages[pkgName] = pkg
		}
	}

	data, err := json.Marshal(pkgs)
	return io.NopCloser(bytes.NewReader(data)), err
}

var ErrNotFound = errors.New("not found")

func (f *Fetcher) downloadChunk(chunk int) ([]json.RawMessage, error) {
	chunkURL, err := url.JoinPath(f.url, strconv.Itoa(chunk)+".json")
	if err != nil {
		panic(err)
	}

	resp, err := http.Get(chunkURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, ErrNotFound
	}

	// deb, err := httputil.DumpResponse(resp, true)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(deb))

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	pkgs := []json.RawMessage{}
	return pkgs, json.Unmarshal(data, &pkgs)
}
