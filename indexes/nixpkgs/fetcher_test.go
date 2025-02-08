package nixpkgs

import (
	"bytes"
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/alecthomas/assert/v2"
)

// TestFetcherOutput tests that the indexer return data
// in the format that the indexer can understand
//
// To generate the expected keys from the packages.json, use:
//   - cat packages.json.br | brotli -d | jq '.packages | keys []' | tr -d '"' > keys.txt
func TestFetcherOutput(t *testing.T) {
	indexer, err := indexer.NewBadger(indexer.BadgerConfig{
		InMemory: true,
	})
	assert.NoError(t, err)
	defer indexer.Close()

	pkgs, err := os.Open("./testdata/packages.json.br")
	assert.NoError(t, err)
	pkgsbr := newBrotli(pkgs)
	defer pkgsbr.Close()

	expectedKeys, err := os.ReadFile("./testdata/keys.txt")
	assert.NoError(t, err)
	actualKeys := bytes.Buffer{}

	err = indexer.Index(pkgsbr, &actualKeys)
	assert.NoError(t, err)

	expectedLines := strings.Split(string(expectedKeys), "\n")
	actualLines := strings.Split(actualKeys.String(), "\n")
	slices.Sort(expectedLines)
	slices.Sort(actualLines)

	assert.Equal(t, expectedLines, actualLines)

	// Skip the first line because actualLines contain
	// an empty string
	for _, pkgName := range actualLines[1:] {
		pkgContent, err := indexer.Load(pkgName)
		assert.NoError(t, err)
		if !json.Valid(pkgContent) {
			t.Fatalf("package content is not a valid JSON:\nPackage: %s\nContent:%s", pkgName, string(pkgContent))
		}
	}
}
