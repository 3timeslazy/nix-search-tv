package nur

import (
	"cmp"
	"fmt"
	"io"
	"strings"

	"github.com/3timeslazy/nix-search-tv/indexes/nixpkgs"
	"github.com/3timeslazy/nix-search-tv/indexes/textutil"
	"github.com/3timeslazy/nix-search-tv/style"
)

func (pkg *Package) Preview(out io.Writer) {
	styler := style.TextStyle

	pkgTitle := textutil.PkgName(pkg.Name) + " " + styler.Dim("("+pkg.GetVersion()+")")
	if pkg.Meta.Broken {
		pkgTitle += " " + styler.Red("(broken)")
	}
	fmt.Fprintln(out, pkgTitle)

	desc := ""
	if pkg.Meta.Description != "" {
		desc = style.Wrap(pkg.Meta.Description) + "\n"
		fmt.Fprintln(out, desc)
	}

	longDesc := ""
	if pkg.Meta.LongDescription != "" && pkg.Meta.Description != pkg.Meta.LongDescription {
		longDesc = style.StyleLongDescription(styler, pkg.Meta.LongDescription)
		fmt.Fprintln(out, longDesc)
	}

	homepages := ""
	if hmpgs := len(pkg.Meta.Homepages); hmpgs > 0 {
		homepages = textutil.Prop(
			textutil.IfElse(hmpgs == 1, "homepage", "homepages"), "",
			strings.Join(pkg.Meta.Homepages, "\n"),
		)
		fmt.Fprintln(out, homepages)
	}

	licenseType := textutil.IfElse(pkg.Meta.Unfree, "unfree", "free")
	license := textutil.Prop(
		"license", styler.Dim("("+licenseType+")"),
		licensesString(pkg.Meta.Licenses),
	)
	fmt.Fprintln(out, license)

	mainProg := ""
	if pkg.Meta.MainProgram != "" {
		mainProg = textutil.Prop(
			"main program", "",
			style.PrintCodeBlock("$ "+pkg.Meta.MainProgram),
		)
		fmt.Fprintln(out, mainProg)
	}

	platforms := ""
	if len(pkg.Meta.Platforms) > 0 {
		platforms = textutil.Prop(
			"platforms", "",
			textutil.Platforms(pkg.Meta.Platforms),
		)
		fmt.Fprintln(out, platforms)
	}
}

func licensesString(ls []nixpkgs.License) string {
	if len(ls) == 0 {
		return "No License"
	}

	ss := []string{}
	for _, l := range ls {
		ss = append(ss, cmp.Or(l.SpdxID, l.FullName))
	}

	return strings.Join(ss, "\n")
}
