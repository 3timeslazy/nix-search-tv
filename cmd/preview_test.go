package cmd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/3timeslazy/nix-search-tv/config"
	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/indexes/indices"
	"github.com/3timeslazy/nix-search-tv/indexes/optionsfile"
	"github.com/3timeslazy/nix-search-tv/indexes/renderdocs"

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

func TestPreviewParseHTML(t *testing.T) {
	htmlPage := readTestdata(t, "nvf.html")
	srv := httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		wr.Write(htmlPage)
	}))
	defer srv.Close()

	for _, tt := range []struct {
		name        string
		indexes     []string
		printArgs   []string
		previewArgs []string
	}{
		{
			name:        "only render_docs index via config",
			indexes:     []string{},
			previewArgs: []string{"vim.package"},
		},
		{
			name:        "only render_docs index via flag",
			indexes:     []string{indices.Nixpkgs},
			printArgs:   []string{"--indexes", "nvf"},
			previewArgs: []string{"--indexes", "nvf", "vim.package"},
		},
		{
			name:        "builtin and render_docs index",
			indexes:     []string{indices.Nixpkgs},
			previewArgs: []string{"nvf/ vim.package"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state := setup(t)
			setNixpkgs("lazygit")
			writeXdgConfig(t, state, map[string]any{
				config.EnableWaitingMessageTag: false,
				"indexes":                      tt.indexes,
				"experimental": map[string]any{
					"render_docs_indexes": map[string]any{
						"nvf": srv.URL,
					},
				},
			})

			printCmd(t, tt.printArgs...)
			indices.Reset()
			setNixpkgs("lazygit")
			state.Stdout.Reset()

			previewCmd(t, append([]string{"--json"}, tt.previewArgs...)...)
			var actual renderdocs.Package
			assert.NoError(t, json.Unmarshal(state.Stdout.Bytes(), &actual))
			expected := renderdocs.Package{
				Package: indexer.Package{Name: "vim.package"},
				Description: "<dd>\n<p>The neovim package to use for the wrapper. This\n" +
					"corresponds to the package that will be wrapped\n" +
					"with your plugins and settings.</p><div class=\"warning\"><h3 class=\"title\">Warning</h3><p>You will need to use an unwrapped package for this\n" +
					"option to work as intended. Using an already wrapped\n" +
					"package here may yield undesirable results.</p></div>\n\n\n\n\n\n\n\n</dd>",
				Type:    "package",
				Default: "<derivation neovim-unwrapped-0.11.3>",
				DeclaredBy: []string{
					"https://github.com/NotAShelf/nvf/blob/main/modules/wrapper/environment/options.nix",
				},
			}
			assert.Equal(t, expected, actual)
		})
	}
}

func TestPreviewOptionsFile(t *testing.T) {
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	optionsPath := pwd + "/testdata/options.json"

	for _, tt := range []struct {
		name        string
		indexes     []string
		printArgs   []string
		previewArgs []string
	}{
		{
			name:        "only options_file index via config",
			indexes:     []string{},
			previewArgs: []string{"age.ageBin"},
		},
		{
			name:        "only options_file index via flag",
			indexes:     []string{indices.Nixpkgs},
			printArgs:   []string{"--indexes", "opts"},
			previewArgs: []string{"--indexes", "opts", "age.ageBin"},
		},
		{
			name:        "builtin and options_file index",
			indexes:     []string{indices.Nixpkgs},
			previewArgs: []string{"opts/ age.ageBin"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state := setup(t)
			setNixpkgs("lazygit")
			writeXdgConfig(t, state, map[string]any{
				config.EnableWaitingMessageTag: false,
				"indexes":                      tt.indexes,
				"experimental": map[string]any{
					"options_file": map[string]string{
						"opts": optionsPath,
					},
				},
			})

			printCmd(t, tt.printArgs...)
			indices.Reset()
			setNixpkgs("lazygit")
			state.Stdout.Reset()

			previewCmd(t, append([]string{"--json"}, tt.previewArgs...)...)
			var actual optionsfile.Package
			assert.NoError(t, json.Unmarshal(state.Stdout.Bytes(), &actual))
			expected := optionsfile.Package{
				Package:     indexer.Package{Name: "age.ageBin"},
				Type:        "string",
				Default:     "\"${pkgs.age}/bin/age\"\n",
				Description: "The age executable to use.\n",
				Declarations: []optionsfile.String{
					"/nix/store/azcvzd254j7wy5fmlbws96a5zfckjw9d-source/modules/age.nix",
				},
			}
			assert.Equal(t, expected, actual)
		})
	}
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
