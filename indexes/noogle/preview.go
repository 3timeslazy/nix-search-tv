package noogle

import (
	"fmt"
	"io"
	"strings"

	"github.com/3timeslazy/nix-search-tv/indexes/textutil"
	"github.com/3timeslazy/nix-search-tv/style"
	"github.com/yuin/goldmark/text"
)

func (pkg *Package) Preview(wr io.Writer) {
	styler := style.TextStyle
	markdown := textutil.NewMarkdown(!style.NoColor())

	title := textutil.PkgName(pkg.Meta.Title)
	fmt.Fprintln(wr, title)
	fmt.Println()

	if len(pkg.Meta.Aliases) > 0 {
		fmt.Fprintln(wr, styler.Bold("aliases"))
		for _, alias := range pkg.Meta.Aliases {
			fmt.Fprintln(wr, strings.Join(alias, "."))
		}
		fmt.Println()
	}

	if pkg.Meta.Signature != "" {
		fmt.Fprintln(wr, styler.Bold("signature"))
		sig := strings.TrimSpace(pkg.Meta.Signature)
		sig = style.Wrap(sig)
		fmt.Fprintln(wr, style.PrintCodeBlock(sig))
		fmt.Println()
	}

	if pkg.Content != nil && len(pkg.Content.Content) > 0 {
		sep := strings.Repeat("─", style.MaxTextWidth())
		fmt.Fprintln(wr, style.TextStyle.Grey(sep))
		content := []byte(pkg.Content.Content)
		parsed := markdown.Parser().Parse(text.NewReader(content))
		markdown.Renderer().Render(wr, content, parsed)
		fmt.Println()
	}

	if pkg.Meta.LambdaExpr != "" {
		fmt.Fprintln(wr, styler.Bold("implementation"))

		srcCode := fmt.Appendf(nil, "```nix\n%s\n```", pkg.Meta.LambdaExpr)
		parsed := markdown.Parser().Parse(text.NewReader(srcCode))
		markdown.Renderer().Render(wr, srcCode, parsed)
	}
}
