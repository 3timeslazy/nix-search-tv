package nixpkgs

import (
	"cmp"
	"fmt"
	"io"
	"strings"

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
	}
	fmt.Fprintln(out, desc)

	// A small hack, mostly for gnomeExtensions.* packages. These packages' long descriptions
	// usually start with the main description. It looks fine at search.nixos.org, but not here
	pkg.Meta.LongDescription = strings.TrimPrefix(pkg.Meta.LongDescription, pkg.Meta.Description)

	if pkg.Meta.LongDescription != "" && pkg.Meta.Description != pkg.Meta.LongDescription {
		longDesc := style.StyleLongDescription(styler, pkg.Meta.LongDescription)
		longDesc += "\n"
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

func licensesString(ls []License) string {
	if len(ls) == 0 {
		return "No License"
	}

	ss := []string{}
	for _, l := range ls {
		ss = append(ss, cmp.Or(l.SpdxID, l.FullName))
	}

	return strings.Join(ss, "\n")
}

func (pkg *Package) GetSource() string {
	src := pkg.Meta.Position
	if src == "" {
		return src
	}

	src, _, _ = strings.Cut(src, ":")
	return "https://github.com/NixOS/nixpkgs/blob/nixos-unstable/" + src
}

func (pkg *Package) GetHomepage() string {
	if len(pkg.Meta.Homepages) > 0 {
		return pkg.Meta.Homepages[0]
	}

	return pkg.GetSource()
}
