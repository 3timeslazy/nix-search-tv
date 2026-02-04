package textutil

import (
	"runtime"
	"slices"
	"strings"

	"github.com/3timeslazy/nix-search-tv/style"
)

var s = style.TextStyle

func PkgName(pkg string) string {
	if pkg == "" {
		return ""
	}

	idx := -1

	quoted := pkg[len(pkg)-1] == '"'
	if quoted {
		for idx = len(pkg) - 2; idx > 0 && pkg[idx] != '"'; idx-- {
		}
	} else {
		idx = strings.LastIndex(pkg, ".") + 1
	}

	styler := style.TextStyle
	left := pkg[:idx]
	if left != "" {
		left = styler.Red(pkg[:idx])
	}
	right := styler.Red(styler.Bold(pkg[idx:]))

	return left + right
}

func Prop(name string, mods string, text string) string {
	name = s.Bold(name)
	if mods != "" {
		name += " " + mods
	}
	return name + "\n" + text + "\n"
}

func IfElse(cond bool, ok, notok string) string {
	if cond {
		return ok
	}
	return notok
}

var printablePlatforms = []string{
	"x86_64-linux",
	"aarch64-linux",
	"x86_64-darwin",
	"aarch64-darwin",
}

func Platforms(platforms []string) string {
	toPrint := []string{}
	cp := currentPlatform()

	// iterate in order printablePlatforms->ps and not
	// vise versa to always print platforms in the same order
	// without sorting
	for _, printable := range printablePlatforms {
		if !slices.Contains(platforms, printable) {
			continue
		}
		if printable != cp {
			printable = style.TextStyle.Dim(printable)
		}
		toPrint = append(toPrint, printable)
	}

	return strings.Join(toPrint, "\n")
}

var go2nixArch = map[string]string{
	"arm64": "aarch64",
	"amd64": "x86_64",
}

func currentPlatform() string {
	arch := go2nixArch[runtime.GOARCH]
	kern := runtime.GOOS
	return arch + "-" + kern
}
