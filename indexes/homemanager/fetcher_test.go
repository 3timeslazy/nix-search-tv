package homemanager

import (
	"bytes"
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/indexes/readutil"

	"github.com/alecthomas/assert/v2"
)

// TestFetcherOutput tests that the indexer return data
// in the format that the indexer can understand
//
// To generate the expected keys from the packages.json, use:
//   - cat options.json | jq -r 'keys []' > keys.txt
func TestFetcherOutput(t *testing.T) {
	// return
	indexer, err := indexer.NewBadger(indexer.BadgerConfig{
		InMemory: true,
	})
	assert.NoError(t, err)
	defer indexer.Close()

	pkgs, err := os.Open("./testdata/options.json")
	assert.NoError(t, err)
	pkgsWrap := readutil.PackagesWrapper(pkgs)
	defer pkgsWrap.Close()

	expectedKeys, err := os.ReadFile("./testdata/keys.txt")
	assert.NoError(t, err)
	actualKeys := bytes.Buffer{}

	err = indexer.Index(pkgsWrap, &actualKeys)
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
