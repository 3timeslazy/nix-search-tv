package homemanager

import (
	"fmt"
	"strings"

	"github.com/3timeslazy/nix-search-tv/indexer"
)

type Package struct {
	indexer.Package
	Example      Example        `json:"example"`
	Type         string         `json:"type"`
	Description  string         `json:"description"`
	Declarations []Declarations `json:"declarations"`
	Default      Default        `json:"default"`
}

type Example struct {
	Text string `json:"text"`
}

type Declarations struct {
	URL string `json:"url"`
}

type Default struct {
	Text string `json:"text"`
}

func (pkg *Package) GetSource() string {
	if len(pkg.Declarations) == 1 {
		return pkg.Declarations[0].URL
	}

	// Home Manager options might have multiple declarations, so
	// return the link to the official documentation with all the links
	return fmt.Sprintf(
		"https://nix-community.github.io/home-manager/print.html#opt-%s",

		// There are packages with quotes in their names, like
		// targets.darwin.defaults."com.apple.menuextra.battery".ShowPercent. For these,
		// the quotes must replaces with "_"
		strings.ReplaceAll(pkg.Name, `"`, "_"),
	)
}

func (pkg *Package) GetHomepage() string {
	return pkg.GetSource()
}
