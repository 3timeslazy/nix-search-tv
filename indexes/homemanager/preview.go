package homemanager

import (
	"fmt"
	"io"

	"github.com/3timeslazy/nix-search-tv/indexes/textutil"
	"github.com/3timeslazy/nix-search-tv/style"
)

func (pkg *Package) Preview(out io.Writer) {
	pkgTitle := textutil.PkgName(pkg.Name) + "\n"
	fmt.Fprint(out, pkgTitle)

	desc := style.StyleHTML(pkg.Description)
	desc += "\n"
	fmt.Fprintln(out, desc)

	typ := textutil.Prop("type", "", pkg.Type)
	fmt.Fprintln(out, typ)

	def := pkg.Default.Text
	if def != "" {
		def = textutil.Prop(
			"default", "",
			style.PrintCodeBlock(pkg.Default.Text),
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
