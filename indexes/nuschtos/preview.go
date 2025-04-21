package nuschtos

import (
	"fmt"
	"io"
	"strings"

	"github.com/3timeslazy/nix-search-tv/indexes/textutil"
	"github.com/3timeslazy/nix-search-tv/style"
)

func (pkg *Package) Preview(out io.Writer) {
	pkgTitle := textutil.PkgName(pkg.Name) + "\n"
	fmt.Fprint(out, pkgTitle)

	desc := strings.TrimSpace(pkg.Description)
	desc = style.StyleHTML(desc)
	fmt.Fprintln(out, desc+"\n")

	typ := textutil.Prop("type", "", pkg.Type)
	fmt.Fprintln(out, typ)

	if pkg.Default != "" {
		def := textutil.Prop(
			"default", "",
			style.PrintCodeBlock(pkg.Default),
		)
		fmt.Fprintln(out, def)
	}

	if pkg.Example != "" {
		example := textutil.Prop(
			"example", "",
			style.PrintCodeBlock(pkg.Example),
		)
		fmt.Fprintln(out, example)
	}
}

func (pkg *Package) GetSource() string {
	return "TODO"
}

func (pkg *Package) GetHomepage() string {
	return pkg.GetSource()
}
