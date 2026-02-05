package cmd

import (
	"cmp"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/3timeslazy/nix-search-tv/config"
	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/indexes/indices"

	"github.com/alecthomas/assert/v2"
	"github.com/urfave/cli/v3"
)

func TestPrintInvalidFlags(t *testing.T) {
	runPrint := func(args ...string) error {
		cmd := cli.Command{
			Writer: io.Discard,
			Flags:  BaseFlags(),
			Action: PrintAction,
		}
		return cmd.Run(context.TODO(), append([]string{"print"}, args...))
	}

	t.Run("given config not exist", func(t *testing.T) {
		setup(t)

		err := runPrint("--config", "unknown.json")
		assert.IsError(t, err, fs.ErrNotExist)
		assert.Contains(t, err.Error(), "unknown.json")
	})

	t.Run("unknown index", func(t *testing.T) {
		setup(t)

		err := runPrint("--indexes", "nixpkgs,unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown index")
	})
}

func TestPrintInvalidConfig(t *testing.T) {
	runPrint := func(args ...string) error {
		cmd := cli.Command{
			Writer: io.Discard,
			Flags:  BaseFlags(),
			Action: PrintAction,
		}
		return cmd.Run(context.TODO(), append([]string{"print"}, args...))
	}

	t.Run("builtin indexes are not allowed in custom config", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{indices.Nixpkgs},
			"experimental": map[string]any{
				"render_docs_indexes": map[string]string{
					indices.Nixpkgs: "http://localhost",
				},
			},
		})

		err := runPrint()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "conflicts")
	})
}

func TestPrintConfig(t *testing.T) {
	t.Run("read the file provided by --config", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: true,
			"indexes":                      []string{indices.HomeManager},
		})

		tmp, err := os.CreateTemp("", "config-file")
		assert.NoError(t, err)
		defer tmp.Close()
		_, err = tmp.WriteString(`{"enable_waiting_message": false, "indexes": ["nixpkgs"]}`)
		assert.NoError(t, err)

		setNixpkgs("nix-search-tv")

		printCmd(t, "--config", tmp.Name())

		expected := []string{
			"nix-search-tv",
			"",
		}
		output := state.Stdout.String()

		assert.Equal(t, expected, strings.Split(output, "\n"))
	})

	t.Run("read default file if no config given", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{indices.Nixpkgs},
		})

		setNixpkgs("nix-search-tv")

		printCmd(t)

		expected := []string{
			"nix-search-tv",
			"",
		}
		output := state.Stdout.String()

		assert.Equal(t, expected, strings.Split(output, "\n"))
	})

	t.Run("use defaults when default file not exist", func(t *testing.T) {
		state := setup(t)

		fetchers := map[string]indexer.Fetcher{
			indices.Nixpkgs:     &PkgsFetcher{[]string{"nixpkg"}},
			indices.HomeManager: &PkgsFetcher{[]string{"home-manager"}},
			indices.Nur:         &PkgsFetcher{[]string{"nur"}},
		}

		osPkg := ""
		osIndex := ""
		switch runtime.GOOS {
		case "darwin":
			osPkg = "darwin"
			osIndex = indices.Darwin

		case "linux":
			osPkg = "nixos"
			osIndex = indices.NixOS
		}
		fetchers[osIndex] = &PkgsFetcher{[]string{osPkg}}
		indices.SetFetchers(fetchers)

		printCmd(t)

		// enable_waiting_message is true by default, so
		// if we see it in the output - the defaults were used
		output := state.Stdout.String()
		expected := []string{
			"",
			waitingMessage,
			"nixpkgs/ nixpkg",
			"home-manager/ home-manager",
			"nur/ nur",
			osIndex + "/ " + osPkg,
		}

		assertSortEqual(t, expected, strings.Split(output, "\n"))
	})
}

func TestPrintIndexing(t *testing.T) {
	runPrint := func(t *testing.T) {
		printCmd(t, "--indexes", indices.Nixpkgs)
	}

	t.Run("no cache -> indexing", func(t *testing.T) {
		state := setup(t)

		setNixpkgs("nix-search-tv")

		runPrint(t)

		expectedPaths := map[string]bool{
			"nixpkgs":               false,
			"nixpkgs/badger":        false,
			"nixpkgs/cache.txt":     false,
			"nixpkgs/metadata.json": false,
		}
		err := filepath.WalkDir(state.CacheDir, func(path string, d fs.DirEntry, err error) error {
			for expectedPath := range expectedPaths {
				if strings.HasSuffix(path, expectedPath) {
					expectedPaths[expectedPath] = true
				}
			}
			return nil
		})
		assert.NoError(t, err)

		for expectedPath, ok := range expectedPaths {
			assert.True(t, ok, "File not found: %s", expectedPath)
		}

		cache := getCache(t, state)
		assertSortEqual(t, []string{"nix-search-tv"}, cache)
	})

	t.Run("has cache -> need indexing -> indexing", func(t *testing.T) {
		state := setup(t)

		setNixpkgs("nix-search-tv")

		// generate the cache
		runPrint(t)

		// reset metadata and run again
		newpkgs := []string{"televison", "fzf"}
		setNixpkgs(newpkgs...)
		setMetadata(t, state, indices.Nixpkgs, indexer.IndexMetadata{})

		runPrint(t)

		cache := getCache(t, state)
		assertSortEqual(t, newpkgs, cache)
	})

	t.Run("has cache -> indexing not needed", func(t *testing.T) {
		setup(t)

		setNixpkgs("nix-search-tv")

		// generate the cache
		runPrint(t)

		// set failing fetcher. If everything is correct, it won't be triggered. If
		// there's a bug, the test fail
		indices.SetFetchers(map[string]indexer.Fetcher{
			indices.Nixpkgs: &FailFetcher{},
		})

		runPrint(t)
	})

	t.Run("an index failed", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: true,
			"indexes":                      []string{indices.Nixpkgs, indices.HomeManager},
		})

		indices.SetFetchers(map[string]indexer.Fetcher{
			indices.Nixpkgs:     &FailFetcher{},
			indices.HomeManager: &PkgsFetcher{[]string{"programs.zsh"}},
		})

		printCmd(t, "--indexes", indices.Nixpkgs+","+indices.HomeManager)

		expected := []string{
			waitingMessage,
			"nixpkgs/ indexing failed: get latest release: failed to get latest release",
			"home-manager/ programs.zsh",
			"",
		}
		output := state.Stdout.String()

		assert.Equal(t, expected, strings.Split(output, "\n"))
	})

	t.Run("need update, but not new version", func(t *testing.T) {

	})
}

func TestPrint(t *testing.T) {
	t.Run("multiple builtin indexes", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{indices.Nixpkgs, indices.HomeManager},
		})

		indices.SetFetchers(map[string]indexer.Fetcher{
			indices.Nixpkgs: &PkgsFetcher{
				pkgs: []string{"lazygit"},
			},
			indices.HomeManager: &PkgsFetcher{
				pkgs: []string{"programs.lazygit.enable"},
			},
		})

		printCmd(t)

		expected := []string{
			"",
			"home-manager/ programs.lazygit.enable",
			"nixpkgs/ lazygit",
		}
		output := strings.Split(state.Stdout.String(), "\n")

		assertSortEqual(t, expected, output)
	})

	t.Run("packages printed in lexicographical order", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{indices.Nixpkgs},
		})

		setNixpkgs(
			"pkg-a",
			"pkg-z",
			"pkg-k",
		)

		printCmd(t)

		expected := []string{
			"",
			"pkg-a",
			"pkg-k",
			"pkg-z",
		}
		output := strings.Split(state.Stdout.String(), "\n")
		assertSortEqual(t, expected, output)
	})

	t.Run("only nixpkgs via flag", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{indices.Nixpkgs, indices.HomeManager},
			"experimental": map[string]any{
				"render_docs_indexes": map[string]any{
					"nvf": "http://localhost",
				},
			},
		})

		setNixpkgs("lazygit")

		printCmd(t, "--indexes", indices.Nixpkgs)

		expected := []string{
			"",
			"lazygit",
		}
		output := strings.Split(state.Stdout.String(), "\n")
		assertSortEqual(t, expected, output)
	})

	t.Run("index duplicate via flag", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{indices.Nixpkgs},
		})

		setNixpkgs("lazygit")

		printCmd(t, "--indexes", "nixpkgs,nixpkgs")

		expected := []string{
			"",
			"lazygit",
		}
		output := strings.Split(state.Stdout.String(), "\n")
		assertSortEqual(t, expected, output)
	})

	t.Run("index duplicate in config", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{indices.Nixpkgs, indices.Nixpkgs},
		})

		setNixpkgs("lazygit")

		printCmd(t)

		expected := []string{
			"",
			"lazygit",
		}
		output := strings.Split(state.Stdout.String(), "\n")
		assertSortEqual(t, expected, output)
	})
}

func TestPrintOffline(t *testing.T) {
	t.Run("single index, cache present", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{indices.Nixpkgs},
		})

		// populate the cache
		setNixpkgs("nix-search-tv")
		printCmd(t)

		// swap in a fetcher that panics on use; --offline must never reach it
		state.Stdout.Reset()
		indices.SetFetchers(map[string]indexer.Fetcher{
			indices.Nixpkgs: &FailFetcher{},
		})

		printCmd(t, "--offline")

		expected := []string{
			"nix-search-tv",
			"",
		}
		assert.Equal(t, expected, strings.Split(state.Stdout.String(), "\n"))
	})

	t.Run("single index, no cache", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{indices.Nixpkgs},
		})

		indices.SetFetchers(map[string]indexer.Fetcher{
			indices.Nixpkgs: &FailFetcher{},
		})

		printCmd(t, "--offline")

		// cache.txt does not exist yet; initFile creates an empty one, so nothing is printed
		assert.Equal(t, "", state.Stdout.String())
	})

	t.Run("two indexes, only one cached", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{indices.Nixpkgs, indices.HomeManager},
		})

		// index only nixpkgs
		indices.SetFetchers(map[string]indexer.Fetcher{
			indices.Nixpkgs:     &PkgsFetcher{pkgs: []string{"nix-search-tv"}},
			indices.HomeManager: &FailFetcher{},
		})
		printCmd(t, "--indexes", indices.Nixpkgs)

		// both fetchers are now failing; --offline must not call either
		state.Stdout.Reset()
		indices.SetFetchers(map[string]indexer.Fetcher{
			indices.Nixpkgs:     &FailFetcher{},
			indices.HomeManager: &FailFetcher{},
		})

		printCmd(t, "--offline")

		// nixpkgs prints from cache (with prefix because two indexes are configured);
		// home-manager has no cache.txt so it contributes nothing
		expected := []string{
			"nixpkgs/ nix-search-tv",
			"",
		}
		assert.Equal(t, expected, strings.Split(state.Stdout.String(), "\n"))
	})
}

func TestParseHTML(t *testing.T) {
	htmlPage := readTestdata(t, "nvf.html")
	srv := httptest.NewServer(http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		wr.Write(htmlPage)
	}))
	defer srv.Close()

	t.Run("only render_docs index via config", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{},
			"experimental": map[string]any{
				"render_docs_indexes": map[string]any{
					"nvf": srv.URL,
				},
			},
		})

		printCmd(t)

		expected := []string{
			"",
			"_module.args",
			"vim.enableLuaLoader",
			"vim.package",
		}
		output := strings.Split(state.Stdout.String(), "\n")
		assertSortEqual(t, expected, output)
	})

	t.Run("only render_docs index via flag", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{indices.Nixpkgs},
			"experimental": map[string]any{
				"render_docs_indexes": map[string]any{
					"nvf": srv.URL,
				},
			},
		})

		printCmd(t, "--indexes", "nvf")

		expected := []string{
			"",
			"_module.args",
			"vim.enableLuaLoader",
			"vim.package",
		}
		output := strings.Split(state.Stdout.String(), "\n")
		assertSortEqual(t, expected, output)
	})

	t.Run("builtin and render_docs index", func(t *testing.T) {
		state := setup(t)

		setNixpkgs("lazygit")

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{indices.Nixpkgs},
			"experimental": map[string]any{
				"render_docs_indexes": map[string]any{
					"nvf": srv.URL,
				},
			},
		})

		printCmd(t)

		expected := []string{
			"",
			"nixpkgs/ lazygit",
			"nvf/ _module.args",
			"nvf/ vim.enableLuaLoader",
			"nvf/ vim.package",
		}
		output := strings.Split(state.Stdout.String(), "\n")
		assertSortEqual(t, expected, output)
	})
}

func TestOptionsFile(t *testing.T) {
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	optionsPath := pwd + "/testdata/options.json"

	t.Run("only options_file index in config", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: false,
			"indexes":                      []string{},
			"experimental": map[string]any{
				"options_file": map[string]string{
					"file": optionsPath,
				},
			},
		})

		printCmd(t)

		expected := strings.Join([]string{
			"age.ageBin\n",
			"nixvim.autoCmd\n",
		}, "")
		output := state.Stdout.String()

		assert.Equal(t, expected, output)
	})

	t.Run("path did not change", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: true,
			"indexes":                      []string{},
			"experimental": map[string]any{
				"options_file": map[string]string{
					"file": optionsPath,
				},
			},
		})

		printCmd(t)

		expected := []string{
			waitingMessage,
			"age.ageBin",
			"nixvim.autoCmd",
			"",
		}
		output := strings.Split(state.Stdout.String(), "\n")
		assert.Equal(t, expected, output)

		indices.Reset()
		state.Stdout.Reset()

		setMetadata(t, state, "file", indexer.IndexMetadata{
			CurrRelease:   optionsPath,
			LastIndexedAt: time.Time{},
		})

		printCmd(t)

		expected = []string{
			// No "Indexing packages..." message means
			// indexing is not happening
			"age.ageBin",
			"nixvim.autoCmd",
			"",
		}
		output = strings.Split(state.Stdout.String(), "\n")

		assert.Equal(t, expected, output)
	})

	t.Run("the path changed", func(t *testing.T) {
		state := setup(t)

		writeXdgConfig(t, state, map[string]any{
			config.EnableWaitingMessageTag: true,
			"indexes":                      []string{},
			"experimental": map[string]any{
				"options_file": map[string]string{
					"file": optionsPath,
				},
			},
		})

		printCmd(t)

		indices.Reset()
		state.Stdout.Reset()

		setMetadata(t, state, "file", indexer.IndexMetadata{
			LastIndexedAt: time.Now().Add(time.Hour * 24 * 2000),
			CurrRelease:   "/nix/store/previous-path/options.json",
		})

		printCmd(t)

		expected := []string{
			waitingMessage,
			"age.ageBin",
			"nixvim.autoCmd",
			"",
		}
		output := strings.Split(state.Stdout.String(), "\n")
		assertSortEqual(t, expected, output)
	})
}

func printCmd(t *testing.T, args ...string) {
	cmd := cli.Command{
		Writer: io.Discard,
		Flags: append(BaseFlags(), &cli.BoolFlag{
			Name:  OfflineFlag,
			Usage: "disable fetching new indexes",
		}),
		Action: PrintAction,
	}
	err := cmd.Run(context.TODO(), append([]string{"print"}, args...))
	assert.NoError(t, err)
}

func writeXdgConfig(t *testing.T, state state, conf map[string]any) {
	confDir := filepath.Join(state.ConfigDir, "nix-search-tv")
	assert.NoError(t, os.MkdirAll(confDir, 0755))

	confPath := filepath.Join(confDir, "config.json")

	confBytes, err := json.Marshal(conf)
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile(confPath, confBytes, 0666))
}

func assertSortEqual[S ~[]E, E cmp.Ordered](t *testing.T, expected, actual S) {
	slices.Sort(expected)
	slices.Sort(actual)
	assert.Equal(t, expected, actual)
}

func readTestdata(t *testing.T, filename string) []byte {
	pwd, err := os.Getwd()
	assert.NoError(t, err)

	path := filepath.Join(pwd, "testdata", filename)

	data, err := os.ReadFile(path)
	assert.NoError(t, err)

	return data
}
