package style

import "strings"

type TextStyler uint8

const (
	StyledText TextStyler = 1 << iota
	dontEndStyle
)

func (s TextStyler) Strikethrough(text string) string {
	return s.styleTextBlock(text, "\x1b[9m", "\x1b[29m")
}

func (s TextStyler) Dim(text string) string {
	return s.styleTextBlock(text, "\x1b[2m", "\x1b[22m")
}

func (s TextStyler) Bold(text string) string {
	return s.styleTextBlock(text, "\x1b[1m", "\x1b[22m")
}

func (s TextStyler) Underline(text string) string {
	return s.styleTextBlock(text, "\x1b[4m", "\x1b[24m")
}

func (s TextStyler) with(o TextStyler) TextStyler {
	return s | o
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
