package noogle

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/alecthomas/assert/v2"
)

// TestFetcherOutput tests that transformJSON produces data
// in the format that the indexer can understand.
//
// To generate the expected keys from noogle.json, use:
//   - cat noogle.json | jq -r '.data[].meta.title' | sort -u > keys.txt
func TestFetcherOutput(t *testing.T) {
	t.Parallel()

	full := loadTestdata(t)

	indexer, err := indexer.NewBadger(indexer.BadgerConfig{InMemory: true})
	assert.NoError(t, err)
	defer indexer.Close()

	rd, err := transformJSON(bytes.NewReader(full.Data), &full)
	assert.NoError(t, err)

	expectedKeys, err := os.ReadFile("./testdata/keys.txt")
	assert.NoError(t, err)
	actualKeys := bytes.Buffer{}

	err = indexer.Index(io.NopCloser(rd), &actualKeys)
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

// TestSetsNixpkgsCommit verifies that every
// package gets NixpkgsCommit set from UpstreamInfo.Rev.
func TestSetsNixpkgsCommit(t *testing.T) {
	t.Parallel()

	full := loadTestdata(t)

	rd, err := transformJSON(bytes.NewReader(full.Data), &full)
	assert.NoError(t, err)

	var result struct {
		Packages map[string]Package `json:"packages"`
	}
	err = json.NewDecoder(rd).Decode(&result)
	assert.NoError(t, err)
	assert.True(t, len(result.Packages) > 0)

	for title, pkg := range result.Packages {
		assert.Equal(t, full.UpstreamInfo.Rev, pkg.NixpkgsCommit, "NixpkgsCommit mismatch for %s", title)
	}
}

// TestBuiltinSignatureFallback verifies that
// builtins.* packages with no signature get their type
// from the BuiltinTypes map.
func TestBuiltinSignatureFallback(t *testing.T) {
	t.Parallel()

	full := loadTestdata(t)

	rd, err := transformJSON(bytes.NewReader(full.Data), &full)
	assert.NoError(t, err)

	var result struct {
		Packages map[string]Package `json:"packages"`
	}
	err = json.NewDecoder(rd).Decode(&result)
	assert.NoError(t, err)

	expectedKeys, err := os.ReadFile("./testdata/signature_fallback.txt")
	assert.NoError(t, err)

	for title := range strings.SplitSeq(strings.TrimSpace(string(expectedKeys)), "\n") {
		pkg, ok := result.Packages[title]
		assert.True(t, ok, "package %q not found", title)
		assert.NotEqual(t, "", pkg.Meta.Signature, "builtin %s should have a signature", title)
	}
}

func TestInvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := transformJSON(strings.NewReader("not json"), &NoogleFull{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse input json")
}

func loadTestdata(t *testing.T) NoogleFull {
	t.Helper()

	f, err := os.Open("./testdata/noogle.json")
	assert.NoError(t, err)
	defer f.Close()

	var full NoogleFull
	err = json.NewDecoder(f).Decode(&full)
	assert.NoError(t, err)

	return full
}
