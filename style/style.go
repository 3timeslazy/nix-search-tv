package style

const (
	BoldStart  = "\x1b[1m"
	DimStart   = "\x1b[2m"
	DimBoldEnd = "\x1b[22m"

	HyperlinkStart = "\x1b]8;;"
	HyperlinkEnd   = "\x07"
)

func Bold(s string) string {
	return BoldStart + s + DimBoldEnd
}

func DimBold(s string) string {
	return BoldStart + s + DimBoldEnd + DimStart
}

func Dim(s string) string {
	return DimStart + s + DimBoldEnd
}

func Hyperlink(url, text string) string {
	return HyperlinkStart + url + HyperlinkEnd + text + (HyperlinkStart + HyperlinkEnd)
}
