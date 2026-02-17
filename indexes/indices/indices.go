// Package indices is a wrapper around indexes/ package providing
// a more convenient API for validating available indexes and picking up
// their fetches and previews
//
// Also, as a lot of code uses `indexes` word for variable names, this package
// creates an `indices` alias to avoid conflicts with `indexes`.
package indices

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/indexes/darwin"
	"github.com/3timeslazy/nix-search-tv/indexes/homemanager"
	"github.com/3timeslazy/nix-search-tv/indexes/nixos"
	"github.com/3timeslazy/nix-search-tv/indexes/nixpkgs"
	"github.com/3timeslazy/nix-search-tv/indexes/noogle"
	"github.com/3timeslazy/nix-search-tv/indexes/nur"
)

type Pkg interface {
	Preview(io.Writer)
	GetSource() string
	GetHomepage() string
}

const (
	Nixpkgs     = "nixpkgs"
	HomeManager = "home-manager"
	Nur         = "nur"
	NixOS       = "nixos"
	Darwin      = "darwin"
	Noogle      = "noogle"
)

var BuiltinIndexes = map[string]bool{
	Nixpkgs:     true,
	HomeManager: true,
	Nur:         true,
	NixOS:       true,
	Darwin:      true,
	Noogle:      true,
}

var newPkgs = map[string]func() Pkg{
	Nixpkgs:     func() Pkg { return &nixpkgs.Package{} },
	HomeManager: func() Pkg { return &homemanager.Package{} },
	Nur:         func() Pkg { return &nur.Package{} },
	NixOS:       func() Pkg { return &nixos.Package{} },
	Darwin:      func() Pkg { return &darwin.Package{} },
	Noogle:      func() Pkg { return &noogle.Package{} },
}

var fetchers = map[string]indexer.Fetcher{
	Nixpkgs:     &nixpkgs.Fetcher{},
	HomeManager: &homemanager.Fetcher{},
	Nur:         &nur.Fetcher{},
	NixOS:       &nixos.Fetcher{},
	Darwin:      &darwin.Fetcher{},
	Noogle:      &noogle.Fetcher{},
}

func Register(
	index string,
	fetcher indexer.Fetcher,
	newpkg func() Pkg,
) error {
	err := registerFetcher(index, fetcher)
	if err != nil {
		return fmt.Errorf("builtin %q fetcher already registered", index)
	}

	err = registerNewPkg(index, newpkg)
	if err != nil {
		return fmt.Errorf("builtin %q builder already registered", index)
	}

	return nil
}

func Preview(index string, out io.Writer, pkgContent json.RawMessage) error {
	pkg, err := getPkg(index, pkgContent)
	if err != nil {
		return err
	}

	pkg.Preview(out)
	return nil
}

func SourcePreview(index string, out io.Writer, pkgContent json.RawMessage) error {
	pkg, err := getPkg(index, pkgContent)
	if err != nil {
		return err
	}

	_, err = out.Write([]byte(pkg.GetSource()))
	return err
}

func HomepagePreview(index string, out io.Writer, pkgContent json.RawMessage) error {
	pkg, err := getPkg(index, pkgContent)
	if err != nil {
		return err
	}

	_, err = out.Write([]byte(pkg.GetHomepage()))
	return err
}

func registerNewPkg(index string, newpkg func() Pkg) error {
	if _, ok := newPkgs[index]; ok {
		return fmt.Errorf("index %q already registered", index)
	}

	newPkgs[index] = newpkg
	return nil
}

func getPkg(index string, pkgContent json.RawMessage) (Pkg, error) {
	newPkg, ok := newPkgs[index]
	if !ok {
		return nil, fmt.Errorf("unknown index: %q", index)
	}

	pkg := newPkg()
	if err := json.Unmarshal(pkgContent, &pkg); err != nil {
		return nil, fmt.Errorf("unmarshal package: %w", err)
	}

	return pkg, nil
}

func registerFetcher(index string, fetcher indexer.Fetcher) error {
	if _, ok := fetchers[index]; ok {
		return fmt.Errorf("index %q already registered", index)
	}

	fetchers[index] = fetcher
	return nil
}

func GetFetcher(index string) (indexer.Fetcher, bool) {
	f, ok := fetchers[index]
	return f, ok
}

// SetFetchers overrides internal fetchers var
// and only used for testing
func SetFetchers(newFetchers map[string]indexer.Fetcher) {
	fetchers = newFetchers
}

func Reset() {
	fetchers = map[string]indexer.Fetcher{}
	newPkgs = map[string]func() Pkg{}
}
