package noogle

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestGetSource(t *testing.T) {
	packages, nixpkgsCommit := loadNooglePkgs(t)

	tests := []struct {
		Name     string
		Title    string
		Expected string
	}{
		{
			Name:     "A builtin",
			Title:    "builtins.add",
			Expected: "https://github.com/search?q=repo:NixOS/nix+symbol:prim_add&type=code",
		},
		{
			Name:     "A builtin alias",
			Title:    "lib.add",
			Expected: "https://github.com/NixOS/nixpkgs/blob/dfcc8a7bfb5b581331aeb110204076188636c7a2/lib/default.nix#L170:C9",
		},
		{
			Name:     "All three positions are present, neither functor, nor primop",
			Title:    "lib.and",
			Expected: "https://github.com/NixOS/nixpkgs/blob/dfcc8a7bfb5b581331aeb110204076188636c7a2/lib/trivial.nix#L219:C3",
		},
		{
			Name:     "All positions are present, is a functor",
			Title:    "pkgs.srcOnly",
			Expected: "https://github.com/NixOS/nixpkgs/blob/dfcc8a7bfb5b581331aeb110204076188636c7a2/pkgs/build-support/src-only/default.nix#L38:C1",
		},
		{
			Name:     "No content position, neither functor, nor primop",
			Title:    "pkgs.agdaPackages.callPackage",
			Expected: "https://github.com/NixOS/nixpkgs/blob/dfcc8a7bfb5b581331aeb110204076188636c7a2/lib/customisation.nix#L637:C9",
		},
		{
			Name:     "A function name without dots",
			Title:    "make-disk-image",
			Expected: "https://github.com/NixOS/nixpkgs/blob/dfcc8a7bfb5b581331aeb110204076188636c7a2/nixos/lib/make-disk-image.nix#L100:C1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			pkg, ok := packages[tt.Title]
			assert.True(t, ok, "package %q not found in testdata", tt.Title)

			pkg.NixpkgsCommit = nixpkgsCommit
			got := pkg.GetSource()
			assert.Equal(t, tt.Expected, got)
		})
	}
}

func TestGetHomepage(t *testing.T) {
	tests := []struct {
		Name     string
		Title    string
		Expected string
	}{
		{
			Name:     "single dot",
			Title:    "builtins.map",
			Expected: "https://noogle.dev/f/builtins/map",
		},
		{
			Name:     "multiple dots",
			Title:    "lib.strings.splitString",
			Expected: "https://noogle.dev/f/lib/strings/splitString",
		},
		{
			Name:     "no dots",
			Title:    "make-disk-image",
			Expected: "https://noogle.dev/f/make-disk-image",
		},
		// The cases below should not happen IRL because
		// those would be incorrect nix expressions
		{
			Name:     "empty title",
			Title:    "",
			Expected: "https://noogle.dev/f/",
		},
		{
			Name:     "trailing dot",
			Title:    "builtins.",
			Expected: "https://noogle.dev/f/builtins/",
		},
		{
			Name:     "consecutive dots",
			Title:    "lib..map",
			Expected: "https://noogle.dev/f/lib//map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			pkg := &Package{
				Meta: Meta{Title: tt.Title},
			}
			got := pkg.GetHomepage()
			assert.Equal(t, tt.Expected, got)
		})
	}
}

func loadNooglePkgs(t *testing.T) (map[string]Package, string) {
	t.Helper()

	data, err := os.ReadFile("./testdata/noogle.json")
	assert.NoError(t, err)

	noogle := NoogleFull{}
	err = json.Unmarshal(data, &noogle)
	assert.NoError(t, err)

	var srcPkgs []Package
	err = json.Unmarshal(noogle.Data, &srcPkgs)
	assert.NoError(t, err)

	pkgs := make(map[string]Package, len(srcPkgs))
	for _, item := range srcPkgs {
		pkgs[item.Meta.Title] = item
	}

	return pkgs, noogle.UpstreamInfo.Rev
}
