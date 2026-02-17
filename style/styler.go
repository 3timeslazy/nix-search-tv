package style

import (
	"os"
	"strings"
)

type TextStyler uint8

var TextStyle = detectStyle()

const (
	noColor = iota
	colored
	dontEndStyle

	defaultANSIEscapeColor = "\x1b[31m" // FgRed
)

func NoColor() bool {
	return os.Getenv("NO_COLOR") != ""
}

func detectStyle() TextStyler {
	if NoColor() {
		return noColor
	}
	return colored
}

func (s TextStyler) Strikethrough(text string) string {
	return s.styleTextBlock(text, "\x1b[9m", "\x1b[29m")
}

func (s TextStyler) Dim(text string) string {
	return s.styleTextBlock(text, "\x1b[2m", "\x1b[22m")
}

func (s TextStyler) Bold(text string) string {
	return s.styleTextBlock(text, "\x1b[1m", "\x1b[22m")
}

func (s TextStyler) Grey(text string) string {
	return s.styleTextBlock(text, "\x1b[38;5;8m", "\x1b[0m")
}

func (s TextStyler) Underline(text string) string {
	return s.styleTextBlock(text, "\x1b[4m", "\x1b[24m")
}

func (s TextStyler) with(o TextStyler) TextStyler {
	return s | o
}

func (s TextStyler) Red(text string) string {
	return s.style(text, defaultANSIEscapeColor, "\x1b[0m")
}

func (s TextStyler) style(text, prefix, suffix string) string {
	if s&1 == 0 {
		return text
	}
	if s&dontEndStyle != 0 {
		suffix = ""
	}
	return prefix + text + suffix
}

func (s TextStyler) styleTextBlock(text string, prefix, suffix string) string {
	if s&1 == 0 {
		return text
	}
	if s&dontEndStyle != 0 {
		suffix = ""
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = prefix + line + suffix
	}
	return strings.Join(lines, "\n")
}
