package nixos

import (
	"fmt"
	"io"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/indexes/textutil"
	"github.com/3timeslazy/nix-search-tv/style"
)

type Package struct {
	indexer.Package
	Example      Example  `json:"example"`
	Type         string   `json:"type"`
	Description  string   `json:"description"`
	Declarations []string `json:"declarations"`
	Default      Example  `json:"default"`
}

type Example struct {
	Text string `json:"text"`
}

func (pkg *Package) Preview(out io.Writer) {
	styler := style.TextStyle

	pkgTitle := textutil.PkgName(pkg.Name) + "\n"
	fmt.Fprint(out, pkgTitle)

	desc := style.StyleLongDescription(styler, pkg.Description)
	desc += "\n"
	fmt.Fprintln(out, desc)

	typ := textutil.Prop("type", "", pkg.Type)
	fmt.Fprintln(out, typ)

	def := pkg.Default.Text
	if def != "" {
		def = textutil.Prop(
			"default", "",
			style.PrintCodeBlock(def),
		)
		fmt.Fprintln(out, def)
	}

	example := ""
	if pkg.Example.Text != "" {
		example = textutil.Prop(
			"example", "",
			style.PrintCodeBlock(pkg.Example.Text),
		)
		fmt.Fprintln(out, example)
	}
}

func (pkg *Package) GetSource() string {
	if len(pkg.Declarations) == 1 {
		return fmt.Sprintf("https://github.com/NixOS/nixpkgs/blob/nixos-unstable/%s", pkg.Declarations[0])
	}

	return fmt.Sprintf(
		"https://search.nixos.org/options?"+
			"channel=unstable"+
			"&from=0&size=1"+
			"&sort=relevances&query=%[1]s"+
			// `show` automatically expands the package definion
			// on the search page. Save users a click!
			"&show=%[1]s",
		pkg.Name,
	)
}

func (pkg *Package) GetHomepage() string {
	return pkg.GetSource()
}
