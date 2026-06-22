package homemanager

import (
	"cmp"
	"context"
	"encoding/json"
	"io"
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/pkgs/renderdocs"

	html2md "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/alecthomas/assert/v2"
)

func TestFetcher(t *testing.T) {
	nixPkgs := buildNix(t)
	htmlPkgs := parseHTML(t)

	// Verify that `nix build` and html parsing
	// give the same packages set
	nixKeys := slices.Collect(maps.Keys(nixPkgs))
	htmlKeys := slices.Collect(maps.Keys(htmlPkgs))
	slices.Sort(nixKeys)
	slices.Sort(htmlKeys)
	assert.Equal(t, nixKeys, htmlKeys)

	reInlineCodeType := regexp.MustCompile(`(?m)({\w+})\x60`)
	reURL := regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)
	reNonWords := regexp.MustCompile(`\W`)

	// Verify that properties are the same
	for _, name := range nixKeys {
		nixPkg := nixPkgs[name]
		htmlPkg := htmlPkgs[name]

		{
			// Some packages are mark as `readOnly` in the options.json
			// In HTML version, however, it adds a `(read only)` prefix
			_, suffix, ok := strings.Cut(htmlPkg.Type, nixPkg.Type)
			skipReadOnly := ok && suffix == " (read only)"

			if !skipReadOnly {
				nixtype := strings.TrimSpace(nixPkg.Type)
				assert.Equal(t, nixtype, htmlPkg.Type, "Incorrect type (%s)", name)
			}
		}
		{
			nixValue := renderdocs.NormProp(nixPkg.Default.Text)
			htmlValue := htmlPkg.Default.Text

			skipOpt := "[](#opt-"+htmlValue+")" == nixValue
			skipQuotes := strings.Trim(htmlValue, "`") == nixValue
			skipPkg := name == "wayland.windowManager.hyprland.finalPortalPackage"

			if !skipOpt && !skipQuotes && !skipPkg {
				assert.Equal(t, []byte(nixValue), []byte(htmlValue), "Incorrect default (%s)", name)
			}
		}
		{
			nixValue := renderdocs.NormProp(nixPkg.Example.Text)
			htmlValue := htmlPkg.Example.Text

			skipQuotes := strings.Trim(htmlValue, "`") == nixValue

			if !skipQuotes {
				assert.Equal(t, nixValue, htmlValue, "Incorrect example (%s)", name)
			}
		}
		{
			nixValue := nixPkg.Declarations
			htmlValue := htmlPkg.Declarations

			skipEmpty := len(nixValue) == 0 && len(htmlValue) == 0

			if !skipEmpty {
				comp := func(d1, d2 Declarations) int { return cmp.Compare(d1.URL, d2.URL) }
				slices.SortFunc(nixValue, comp)
				slices.SortFunc(htmlValue, comp)
				assert.Equal(t, len(nixValue), len(htmlValue), "Incorrect declarations length (%s)", name)
				assert.Equal(t, nixValue, htmlValue, "Incorrect declarations (%s)", name)
			}
		}
		{
			nixValue := nixPkg.Description
			nixValue = reInlineCodeType.ReplaceAllString(nixValue, "`")

			htmlValue, err := html2md.ConvertString(htmlPkg.Description)
			assert.NoError(t, err)

			nixValue = reNonWords.ReplaceAllString(nixValue, "")
			htmlValue = reNonWords.ReplaceAllString(htmlValue, "")

			nixValue = reURL.ReplaceAllString(htmlValue, "")
			htmlValue = reURL.ReplaceAllString(htmlValue, "")

			assert.Equal(t, nixValue, htmlValue, "Incorrect description (%s)", name)
		}
	}
}

func parseHTML(t *testing.T) map[string]Package {
	fetcher := Fetcher{}

	pkgs, err := fetcher.DownloadRelease(context.Background(), "file://./testdata/print.html")
	assert.NoError(t, err)
	defer pkgs.Close()

	return parsePkgs(t, pkgs)
}

func buildNix(t *testing.T) map[string]Package {
	fetcher := NixBuilder{}

	rdr, err := fetcher.DownloadRelease(context.Background(), "./testdata/options.json")
	assert.NoError(t, err)
	defer rdr.Close()

	return parsePkgs(t, rdr)
}

func parsePkgs(t *testing.T, rdr io.Reader) map[string]Package {
	pkgs := indexer.Indexable{}
	err := json.NewDecoder(rdr).Decode(&pkgs)
	assert.NoError(t, err)

	out := map[string]Package{}

	for name, content := range pkgs.Packages {
		var p Package
		err = json.Unmarshal(content, &p)
		assert.NoError(t, err)
		out[name] = p
	}

	return out
}
