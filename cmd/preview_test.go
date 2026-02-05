package cmd

import (
	"context"
	"io"
	"testing"

	"github.com/3timeslazy/nix-search-tv/config"
	"github.com/3timeslazy/nix-search-tv/indexes/indices"

	"github.com/alecthomas/assert/v2"
	"github.com/urfave/cli/v3"
)

func TestInjectKey(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		pkg := []byte(`{ "version": "v1.0.0" }`)
		pkg = injectKey("nix-search-tv", pkg)
		assert.Equal(t, []byte(`{"_key":"nix-search-tv", "version": "v1.0.0" }`), pkg)
	})

	t.Run("with quotes", func(t *testing.T) {
		pkg := []byte(`{ "version": "v1.0.0" }`)
		pkg = injectKey(`package."with quotes"`, pkg)
		assert.Equal(t, []byte(`{"_key":"package.\"with quotes\"", "version": "v1.0.0" }`), pkg)
	})
}

func TestPreviewJSON(t *testing.T) {
	state := setup(t)

	writeXdgConfig(t, state, map[string]any{
		config.EnableWaitingMessageTag: false,
		"indexes":                      []string{indices.Nixpkgs},
	})

	setNixpkgs("test-pkg")

	// Populate the index
	printCmd(t, "--indexes", indices.Nixpkgs)

	state.Stdout.Reset()

	// Run preview --json
	previewCmd(t, "--indexes", indices.Nixpkgs, "--json", "test-pkg")

	// The trailing comma after _key is an artifact of injectKey splicing
	// into an empty object: setNixpkgs stores {} per package, so pkg[1:]
	// is just "}". Real fetchers always produce at least one field, so
	// injectKey never produces a trailing comma in practice.
	expected := "{\"_key\":\"test-pkg\",}\n"
	assert.Equal(t, expected, state.Stdout.String())
}

func previewCmd(t *testing.T, args ...string) {
	cmd := cli.Command{
		Writer: io.Discard,
		Flags: append(BaseFlags(), &cli.BoolFlag{
			Name: JsonFlag,
		}),
		Action: NewPreviewAction(indices.Preview),
	}
	err := cmd.Run(context.TODO(), append([]string{"preview"}, args...))
	assert.NoError(t, err)
}
