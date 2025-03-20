package nixpkgs

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/3timeslazy/nix-search-tv/indexes/textutil"
	"github.com/3timeslazy/nix-search-tv/style"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func Preview(out io.Writer, pkg Package) {
	styler := style.StyledText

	pkgTitle := textutil.PkgName(pkg.Name) + " " + styler.Dim("("+pkg.GetVersion()+")")
	if pkg.Meta.Broken {
		pkgTitle += " " + styler.Red("(broken)")
	}
	fmt.Fprintln(out, pkgTitle)

	desc := ""
	if pkg.Meta.Description != "" {
		desc = style.Wrap(pkg.Meta.Description, "") + "\n"
	}
	fmt.Fprintln(out, desc)

	if pkg.Meta.LongDescription != "" && pkg.Meta.Description != pkg.Meta.LongDescription {
		// longDesc := style.StyleLongDescription(style.StyledText, pkg.Meta.LongDescription)
		longDesc := StyleLongDescription(pkg.Meta.LongDescription)
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

func StyleLongDescription(desc string) string {
	opts := goldmark.WithRendererOptions(
		renderer.WithNodeRenderers(util.Prioritized(&TextRenderer{}, 200)),
	)
	md := goldmark.New(opts)

	node := md.Parser().Parse(text.NewReader([]byte(desc)))
	node.Dump([]byte(desc), 2)

	buf := bytes.Buffer{}
	md.Convert([]byte(desc), &buf)

	return style.Dim(strings.TrimSpace(buf.String()))
}

type TextRenderer struct{}

func (tr *TextRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, tr.renderFencedCodeBlock)
	reg.Register(ast.KindParagraph, tr.renderParagraph)
	reg.Register(ast.KindCodeSpan, tr.renderCodeSpan)
	reg.Register(ast.KindText, tr.renderText)
	reg.Register(ast.KindListItem, tr.renderListItem)
	reg.Register(ast.KindList, tr.renderList)
	reg.Register(ast.KindBlockquote, tr.renderBlockquote)
	reg.Register(ast.KindLink, tr.renderLink)
	reg.Register(ast.KindCodeBlock, tr.renderFencedCodeBlock)
	reg.Register(ast.KindEmphasis, tr.renderEmphasis)
}

func (tr *TextRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// fenced := node.(*ast.FencedCodeBlock)
	fenced := node
	if entering {
		w.Write([]byte("\n\n"))

		lines := fenced.Lines()
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			seg.Padding = 2
			lines.Set(i, seg)
		}

		w.Write(fenced.Lines().Value(source))
		w.Write([]byte("\n"))
	}
	return ast.WalkContinue, nil
}

func (tr *TextRenderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// This renderer actually removes <p> tags around paragraphs
	return ast.WalkContinue, nil
}

func (tr *TextRenderer) renderCodeSpan(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	p := node.(*ast.CodeSpan)
	if entering {
		text := string(p.Text(source))
		text = style.DimBold(text)
		w.WriteString(text)
		return ast.WalkSkipChildren, nil
	}

	return ast.WalkContinue, nil
}

func (tr *TextRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	p := node.(*ast.Text)
	if entering {
		text := string(p.Text(source))
		w.WriteString(text)
		if p.SoftLineBreak() {
			w.Write([]byte("\n"))
		}
		return ast.WalkContinue, nil
	}

	return ast.WalkContinue, nil
}

func (tr *TextRenderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// p := node.(*ast.ListItem)

	if !entering {
		w.Write([]byte("\n"))
		return ast.WalkContinue, nil
	}

	bullet := "-"
	if ordered {
		bullet = strconv.Itoa(listItem) + "."
	}

	padding := strings.Repeat("  ", list)
	// ■
	w.Write([]byte(padding + bullet + " "))

	listItem++

	return ast.WalkContinue, nil
}

var list = -1
var ordered = false
var listItem = -1

func (tr *TextRenderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	l := node.(*ast.List)

	ordered = l.IsOrdered()
	if ordered {
		listItem = l.Start
	}

	if entering {
		list++
		if list == 0 {
			w.Write([]byte("\n"))
		}
	} else {
		list--
		if list == -1 {
			w.Write([]byte("\n"))
		}
	}

	return ast.WalkContinue, nil
}

func (tr *TextRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	w.Write([]byte("\n\n"))
	return ast.WalkContinue, nil
}

func (tr *TextRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	link := node.(*ast.Link)

	text := string(link.Text(source))
	text = strings.ReplaceAll(text, "\n", " ")

	dest := string(link.Destination)

	dest = style.Hyperlink(dest, text)

	w.Write([]byte(dest))

	return ast.WalkSkipChildren, nil
}

func (tr *TextRenderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.Write([]byte(style.BoldStart))
	} else {
		w.Write([]byte(style.DimBoldEnd + style.DimStart))
	}

	return ast.WalkContinue, nil
}
