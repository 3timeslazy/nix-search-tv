package nixpkgs

import (
	"strings"
	"testing"

	"github.com/3timeslazy/nix-search-tv/style"
	"github.com/alecthomas/assert/v2"
)

// antora-ui-default
// - SDL2_gfx
// - github-copilot-intellij-agent
// - descent3 (ordered list)

func TestStyleLongDescription(t *testing.T) {

	t.Run("fenced code block", func(t *testing.T) {
		valent := "Note that you have to open firewall ports for other devices\nto connect to it. Use either:\n```nix\nprograms.kdeconnect = {\n  enable = true;\n  package = pkgs.valent;\n}\n```\nor open corresponding firewall ports directly:\n```nix\nnetworking.firewall = rec {\n  allowedTCPPortRanges = [ { from = 1714; to = 1764; } ];\n  allowedUDPPortRanges = allowedTCPPortRanges;\n}\n```"

		expected := strings.Join([]string{
			"Note that you have to open firewall ports for other devices\nto connect to it. Use either:",
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
		}, "\n")
		expected = style.Dim(expected)
		actual := StyleLongDescription(valent)

		assert.Equal(t, []byte(expected), []byte(actual))
	})

	t.Run("list", func(t *testing.T) {
		acmesh := "An ACME Shell script: acme.sh\n\n- An ACME protocol client written purely in Shell (Unix shell) language.\n- Full ACME protocol implementation.\n- Support ECDSA certs\n- Support SAN and wildcard certs\n- Simple, powerful and very easy to use. You only need 3 minutes to learn it.\n- Bash, dash and sh compatible.\n- Purely written in Shell with no dependencies on python.\n- Just one script to issue, renew and install your certificates automatically.\n- DOES NOT require root/sudoer access.\n- Docker ready\n- IPv6 ready\n- Cron job notifications for renewal or error etc.\n"

		expected := strings.Join([]string{
			"An ACME Shell script: acme.sh",
			"",
			"- An ACME protocol client written purely in Shell (Unix shell) language.",
			"- Full ACME protocol implementation.",
			"- Support ECDSA certs",
			"- Support SAN and wildcard certs",
			"- Simple, powerful and very easy to use. You only need 3 minutes to learn it.",
			"- Bash, dash and sh compatible.",
			"- Purely written in Shell with no dependencies on python.",
			"- Just one script to issue, renew and install your certificates automatically.",
			"- DOES NOT require root/sudoer access.",
			"- Docker ready",
			"- IPv6 ready",
			"- Cron job notifications for renewal or error etc.",
		}, "\n")
		expected = style.Dim(expected)
		actual := StyleLongDescription(acmesh)

		assert.Equal(t, []byte(expected), []byte(actual))
	})

	t.Run("code span", func(t *testing.T) {
		vale := "Vale in Nixpkgs offers the helper `.withStyles` allow you to install it\npredefined styles:\n\n```nix\nvale.withStyles (s: [ s.alex s.google ])\n```\n"

		expected := strings.Join([]string{
			"Vale in Nixpkgs offers the helper " + style.DimBold(".withStyles") + " allow you to install it",
			"predefined styles:",
			"",
			"  vale.withStyles (s: [ s.alex s.google ])",
		}, "\n")
		expected = style.Dim(expected)
		actual := StyleLongDescription(vale)

		assert.Equal(t, []byte(expected), []byte(actual))
	})

	t.Run("code block", func(t *testing.T) {
		aseprite := "Aseprite is a program to create animated sprites. Its main features are:\n\n          - Sprites are composed by layers & frames (as separated concepts).\n          - Supported color modes: RGBA, Indexed (palettes up to 256 colors), and Grayscale.\n          - Load/save sequence of PNG files and GIF animations (and FLC, FLI, JPG, BMP, PCX, TGA).\n"

		expected := strings.Join([]string{
			"Aseprite is a program to create animated sprites. Its main features are:",
			"",
			"        - Sprites are composed by layers & frames (as separated concepts).",
			"        - Supported color modes: RGBA, Indexed (palettes up to 256 colors), and Grayscale.",
			"        - Load/save sequence of PNG files and GIF animations (and FLC, FLI, JPG, BMP, PCX, TGA).",
		}, "\n")
		expected = style.Dim(expected)
		actual := StyleLongDescription(aseprite)

		assert.Equal(t, []byte(expected), []byte(actual))
	})

	t.Run("nested list", func(t *testing.T) {
		babashka := "Goals:\n\n- Low latency Clojure scripting alternative to JVM Clojure.\n- Familiarity and portability:\n  - Scripts should be compatible with JVM Clojure as much as possible\n  - Scripts should be platform-independent as much as possible. Babashka\n    offers support for linux, macOS and Windows.\n- Allow interop with commonly used classes like java.io.File and System\n- Multi-threading support (pmap, future, core.async)\n"

		expected := strings.Join([]string{
			"Goals:",
			"- Low latency Clojure scripting alternative to JVM Clojure.",
			"- Familiarity and portability:",
			"  - Scripts should be compatible with JVM Clojure as much as possible",
			"  - Scripts should be platform-independent as much as possible. Babashka",
			// TODO: would be nice to add padding to this line as well
			"offers support for linux, macOS and Windows.",
			// TODO: find where the newline comes from
			"",
			"- Allow interop with commonly used classes like java.io.File and System",
			"- Multi-threading support (pmap, future, core.async)",
		}, "\n")
		expected = style.Dim(expected)
		actual := StyleLongDescription(babashka)

		assert.Equal(t, []byte(expected), []byte(actual))
	})

	t.Run("emphasis", func(t *testing.T) {
		bup := "Highly efficient file backup system based on the git packfile format.\nCapable of doing *fast* incremental backups of virtual machine images.\n"

		expected := strings.Join([]string{
			"Highly efficient file backup system based on the git packfile format.",
			"Capable of doing " + style.DimBold("fast") + " incremental backups of virtual machine images.",
		}, "\n")
		expected = style.Dim(expected)
		actual := StyleLongDescription(bup)

		assert.Equal(t, []byte(expected), []byte(actual))
	})
}
