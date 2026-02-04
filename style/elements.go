package style

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/3timeslazy/nix-search-tv/pkgs/renderdocs"

	"github.com/mitchellh/go-wordwrap"
	"golang.org/x/term"
)

var (
	// ` - \x60

	reInlineHyperlink = regexp.MustCompile(`(?m)\[(.*?)\]\n*\((http.*?)\)`)
	reInlineCode      = regexp.MustCompile(`(?m)\x60\x60?(.*?)\x60\x60?`)

	// reInlineCodeType matches things like
	//
	// {command}`ls .`
	//
	// to later turne them into
	//
	// `ls .`
	reInlineCodeType = regexp.MustCompile(`(?m)({\w+})\x60`)
)

func maxTextWidth() int {
	if c := os.Getenv("FZF_PREVIEW_COLUMNS"); c != "" {
		textWidth, _ := strconv.Atoi(c)
		return textWidth
	}

	// Because this code runs within the preview command, which is
	// called by a fuzzy finder, it's likely this call will return
	// an error, so the output will be 0.
	//
	// Keeping this check for rare cases when I need to test the
	// preview in the terminal
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil {
		return termWidth
	}

	// Let's hope this width will work for
	// most of the people
	return 80
}

func Wrap(text string) string {
	width := maxTextWidth()

	if width < 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = wordwrap.WrapString(line, uint(width))
	}

	return strings.Join(lines, "\n")
}

func StyleHTML(text string) string {
	md := renderdocs.RenderHTML(text)
	return StyleLongDescription(TextStyle, md)
}

func StyleLongDescription(styler TextStyler, text string) string {
	linkReplace := styler.Bold("$1") + styler.with(dontEndStyle).Dim(" ($2)")
	codeReplace := styler.Bold("$1") + styler.with(dontEndStyle).Dim("")

	for _, f := range []func(string) string{
		styleFencedCodeBlock,
		styleCallouts,
		func(text string) string { return reInlineHyperlink.ReplaceAllString(text, linkReplace) },
		func(text string) string { return reInlineCodeType.ReplaceAllString(text, "`") },
		func(text string) string { return reInlineCode.ReplaceAllString(text, codeReplace) },
		func(text string) string { return Wrap(text) },
		func(text string) string { return strings.TrimSpace(text) },
		styler.Dim,
	} {
		text = f(text)
	}

	return text
}

var reFencedCodeBlock = regexp.MustCompile(`(?ms)\x60\x60\x60+\s*(.*?)\s*\x60\x60\x60+`)

func styleFencedCodeBlock(text string) string {
	var sb strings.Builder
	sb.Grow(len(text))

	var start int
	for _, is := range reFencedCodeBlock.FindAllStringSubmatchIndex(text, -1) {
		sb.WriteString(text[start:is[0]])
		start = is[1]
		for _, codeLine := range strings.Split(text[is[2]:is[3]], "\n") {
			// This check removes the language info from markdown-like
			// expressions like
			//
			// ```go
			//   fmt.Println("something")
			// ```
			//
			// turning the text above into
			//
			//   fmt.Println("something")
			//
			// This is a dumb way to do it, but
			// works for now
			if strings.Contains(text, "```"+codeLine) {
				sb.WriteString("\n")
				continue
			}
			sb.WriteString("  ")
			sb.WriteString(codeLine)
			sb.WriteString("\n")
		}
	}
	sb.WriteString(text[start:])

	return sb.String()
}

var (
	reCallouts    = regexp.MustCompile(`(?msU):::\s*{\.\w+}\s*(.*)\s*:{2,3}?`)
	reCalloutType = regexp.MustCompile(`{\.(\w+)}`)
)

func styleCallouts(text string) string {
	return reCallouts.ReplaceAllStringFunc(text, func(str string) string {
		lines := strings.Split(str, "\n")
		typ := reCalloutType.FindStringSubmatch(lines[0])[1]

		prefix := TextStyle.Bold("| ")
		switch typ {
		case "warning", "important", "caution":
			prefix = TextStyle.Red("> ")
		}

		lines[0], lines[len(lines)-1] = "", ""

		for i := range lines {
			lines[i] = prefix + lines[i]
		}

		return strings.Join(lines, "\n")
	})
}

// PrintCodeBlock prints the given content inside a styled code block
func PrintCodeBlock(content string) string {
	// Determine the maximum width of the content
	lineWidth := findMaxWidth(content) + 4 // Add padding for borders
	topBorder := "┌" + strings.Repeat("─", lineWidth-2) + "┐"
	bottomBorder := "└" + strings.Repeat("─", lineWidth-2) + "┘"
	leftBorder := "│ "
	rightBorder := " │"

	block := topBorder + "\n"
	for _, line := range strings.Split(content, "\n") {
		paddedLine := fmt.Sprintf("%-*s", lineWidth-4, line) // Ensure lines fit inside the box
		block += leftBorder + paddedLine + rightBorder + "\n"
	}

	return block + bottomBorder
}

// findMaxWidth calculates the longest line length in the given content
func findMaxWidth(content string) int {
	maxWidth := 0
	for _, line := range strings.Split(content, "\n") {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	return maxWidth
}
