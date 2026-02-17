package noogle

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/3timeslazy/nix-search-tv/indexer"
)

type Fetcher struct {
	// This is so far the first stateful fetcher. This is because
	// noogle returns both "version" and "data" in a single HTTP request.
	// To not make the same request twice, store it.
	data NoogleFull
}

type NoogleFull struct {
	Data         json.RawMessage   `json:"data"`
	BuiltinTypes map[string]FnType `json:"builtinTypes"`
	UpstreamInfo UpstreamInfo      `json:"upstreamInfo"`
}

type UpstreamInfo struct {
	Rev string `json:"rev"`
}

type FnType struct {
	FnType string `json:"fn_type"`
}

func (fetcher *Fetcher) GetLatestRelease(_ context.Context, _ indexer.IndexMetadata) (string, error) {
	resp, err := http.Get("https://noogle.dev/api/v1/data")
	if err != nil {
		return "", fmt.Errorf("fetch noogle data: %w", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&fetcher.data)
	if err != nil {
		return "", fmt.Errorf("parse noogle data: %w", err)
	}

	return fetcher.data.UpstreamInfo.Rev, nil
}

func (fetcher *Fetcher) DownloadRelease(_ context.Context, _ string) (io.ReadCloser, error) {
	rd, err := transformJSON(bytes.NewReader(fetcher.data.Data), &fetcher.data)
	if err != nil {
		return nil, fmt.Errorf("transform noogle data: %w", err)
	}

	return io.NopCloser(rd), nil
}

func transformJSON(r io.Reader, data *NoogleFull) (io.Reader, error) {
	var srcPkgs []Package
	if err := json.NewDecoder(r).Decode(&srcPkgs); err != nil {
		return nil, fmt.Errorf("parse input json: %w", err)
	}

	pkgs := make(map[string]Package)
	for _, item := range srcPkgs {
		title := item.Meta.Title

		btype := data.BuiltinTypes[strings.TrimPrefix(title, "builtins.")].FnType
		item.Meta.Signature = cmp.Or(item.Meta.Signature, btype)
		item.NixpkgsCommit = data.UpstreamInfo.Rev

		pkgs[title] = item
	}

	indexerPkgs := struct {
		Packages map[string]Package `json:"packages"`
	}{
		Packages: pkgs,
	}
	out, err := json.Marshal(indexerPkgs)
	if err != nil {
		return nil, fmt.Errorf("marshal noogle data for indexer: %w", err)
	}

	return bytes.NewReader(out), nil
}
