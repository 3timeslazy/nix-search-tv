package homemanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/indexes/readutil"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type Fetcher struct{}

// Home Manager publishes its options as an mdBook site. print.html is
// the single-page rendering containing every option, parsed by
// ParseMdBook. (The old nixos-render-docs options.xhtml is now a
// redirect stub.)
const htmlURL = "https://nix-community.github.io/home-manager/print.html"

func (Fetcher) GetLatestRelease(ctx context.Context, md indexer.IndexMetadata) (string, error) {
	return time.Now().String(), nil
}

func (Fetcher) DownloadRelease(_ context.Context, release string) (io.ReadCloser, error) {
	var doc *html.Node
	var err error

	_, path, ok := strings.Cut(release, "file://")
	if ok {
		doc, err = htmlquery.LoadDoc(path)
	} else {
		doc, err = htmlquery.LoadURL(htmlURL)
	}
	if err != nil {
		return nil, fmt.Errorf("download print.html: %w", err)
	}

	htmlPkgs, err := ParseMdBook(doc)
	if err != nil {
		return nil, fmt.Errorf("parse print.html: %w", err)
	}
	if len(htmlPkgs) == 0 {
		return nil, fmt.Errorf("parse print.html: no options found (docs layout may have changed)")
	}

	// `Package` represents how data is stored
	// on disk. Its structure inhered from previous `nix build` fetcher.
	// To not break compatibility with older versions, convert
	// it here
	pkgs := map[string]Package{}
	for name, htmlPkg := range htmlPkgs {
		pkg := Package{
			Package: indexer.Package{
				Name: htmlPkg.Name,
			},
			Example: Example{
				Text: htmlPkg.Example,
			},
			Type:        htmlPkg.Type,
			Description: htmlPkg.Description,
			Default: Default{
				Text: htmlPkg.Default,
			},
		}
		for _, decl := range htmlPkg.DeclaredBy {
			pkg.Declarations = append(pkg.Declarations, Declarations{
				URL: decl,
			})
		}

		pkgs[name] = pkg
	}

	buf := &bytes.Buffer{}
	err = json.NewEncoder(buf).Encode(pkgs)
	if err != nil {
		return nil, fmt.Errorf("encode json: %w", err)
	}

	return readutil.PackagesWrapper(io.NopCloser(buf)), nil
}
