package style

import (
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

var s = TextStyle

const dimStart = "\x1b[2m"

func TestStyleLongDescription(t *testing.T) {
	// A hack to disable wrapping. Because tests might be running in different environments,
	// the terminal width may vary between runs. Disable wrapping to make
	// the tests more stable
	assert.NoError(t, os.Setenv("FZF_PREVIEW_COLUMNS", "-1"))

	cases := []struct {
		Desc     string
		Input    string
		Expected []string
	}{
		{
			Desc:     "plain text [nixpkgs.fzf]",
			Input:    "Command-line fuzzy finder written in Go",
			Expected: []string{"Command-line fuzzy finder written in Go"},
		},
		{
			Desc:  "inline code [nixpkgs.vale]",
			Input: "Vale in Nixpkgs offers the `.withStyles` helper",
			Expected: []string{
				"Vale in Nixpkgs offers the " + s.Bold(".withStyles") + dimStart + " helper",
			},
		},
		{
			Desc:  "fenced code blocks [nixpkgs.valent]",
			Input: "To open firewall ports for other devices to connect to it. Use either:\n```nix\nprograms.kdeconnect = {\n  enable = true;\n  package = pkgs.valent;\n}\n```\nor open corresponding firewall ports directly:\n```nix\nnetworking.firewall = rec {\n  allowedTCPPortRanges = [ { from = 1714; to = 1764; } ];\n  allowedUDPPortRanges = allowedTCPPortRanges;\n}\n```\n",
			Expected: []string{
				"To open firewall ports for other devices to connect to it. Use either:",
				"",
				"  programs.kdeconnect = {",
				"    enable = true;",
				"    package = pkgs.valent;",
				"  }",
				"",
				"or open corresponding firewall ports directly:",
				"",
				"  networking.firewall = rec {",
				"    allowedTCPPortRanges = [ { from = 1714; to = 1764; } ];",
				"    allowedUDPPortRanges = allowedTCPPortRanges;",
				"  }",
			},
		},
		{
			Desc:  "fenced code block without language [nixpkgs.firefoxpwa]",
			Input: "To install the package on NixOS, you need to add the following options:\n```\nprograms.firefox.nativeMessagingHosts.packages = [ pkgs.firefoxpwa ];\nenvironment.systemPackages = [ pkgs.firefoxpwa ];\n```",
			Expected: []string{
				"To install the package on NixOS, you need to add the following options:",
				"  programs.firefox.nativeMessagingHosts.packages = [ pkgs.firefoxpwa ];",
				"  environment.systemPackages = [ pkgs.firefoxpwa ];",
			},
		},
		{
			Desc:  "hyperlink [nixpkgs.grmon]",
			Input: "To use it, instrument your Go code following the [usage description](https://github.com/bcicen/grmon?tab=readme-ov-file#usage).\n",
			Expected: []string{
				"To use it, instrument your Go code following the " + s.Bold("usage description") + dimStart +
					" (https://github.com/bcicen/grmon?tab=readme-ov-file#usage).",
			},
		},
		{
			Desc:  "inline block inside callout [nixos.virtualisation.vmware.host.enable]",
			Input: "::: {.note}\ndisable `TRANSPARENT_HUGEPAGE`\n:::",
			Expected: []string{
				s.Bold("| "),
				s.Bold("| ") + "disable " + s.Bold("TRANSPARENT_HUGEPAGE") + dimStart,
				s.Bold("| "),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.Desc, func(t *testing.T) {
			actual := StyleLongDescription(s, c.Input)
			expected := strings.Join(c.Expected, "\n")
			expected = s.Dim(expected)
			assert.Equal(t, expected, actual)
		})
	}
}

func TestStyleCallouts(t *testing.T) {
	warn := s.Red("> ")
	note := s.Bold("| ")

	cases := []struct {
		Desc     string
		Input    string
		Expected []string
	}{
		{
			"warning callout",
			"::: {.warning}\n  warning\n:::",
			[]string{
				warn,
				warn + "  warning",
				warn,
			},
		},
		{
			"note callout",
			"::: {.note}\n  note\n:::",
			[]string{
				note,
				note + "  note",
				note,
			},
		},
		{
			"unknown callout",
			"::: {.unknown}\n  unknown\n:::",
			[]string{
				note,
				note + "  unknown",
				note,
			},
		},
		{
			":: at the end",
			"::: {.warning}\n  warning\n::",
			[]string{
				warn,
				warn + "  warning",
				warn,
			},
		},
		{
			"multiple",
			"::: {.warning}\n warning\n::\n::: {.note}\n note\n:::",
			[]string{
				warn,
				warn + " warning",
				warn,
				note,
				note + " note",
				note,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Desc, func(t *testing.T) {
			actual := styleCallouts(c.Input)
			expected := strings.Join(c.Expected, "\n")
			assert.Equal(t, expected, actual)
		})
	}
}
