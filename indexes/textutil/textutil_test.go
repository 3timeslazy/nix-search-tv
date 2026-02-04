package textutil

import (
	"testing"

	"github.com/3timeslazy/nix-search-tv/style"
	"github.com/alecthomas/assert/v2"
)

func TestPkgName(t *testing.T) {
	styler := style.TextStyle

	cases := []struct {
		PkgName  string
		Expected string
	}{
		{
			PkgName:  "pkg",
			Expected: styler.Red(styler.Bold("pkg")),
		},
		{
			PkgName:  "R",
			Expected: styler.Red(styler.Bold("R")),
		},
		{
			PkgName:  "pkg.settings",
			Expected: styler.Red("pkg.") + styler.Red(styler.Bold("settings")),
		},
		{
			PkgName:  `pkg."settings"`,
			Expected: styler.Red("pkg.") + styler.Red(styler.Bold(`"settings"`)),
		},
		{
			PkgName:  `pkg."settings.global"`,
			Expected: styler.Red("pkg.") + styler.Red(styler.Bold(`"settings.global"`)),
		},
		{
			PkgName:  `pkg.settings.global"`,
			Expected: styler.Red(styler.Bold(`pkg.settings.global"`)),
		},
		{
			PkgName:  `pkg."settings.global`,
			Expected: styler.Red(`pkg."settings.`) + styler.Red(styler.Bold("global")),
		},
		{
			PkgName:  "",
			Expected: "",
		},
	}
	for _, c := range cases {
		t.Run(c.PkgName, func(t *testing.T) {
			actual := PkgName(c.PkgName)
			assert.Equal(t, []byte(c.Expected), []byte(actual))
		})
	}
}
