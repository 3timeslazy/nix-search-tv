package textutil

import (
	"github.com/3timeslazy/nix-search-tv/style"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	fences "github.com/stefanfritsch/goldmark-fences"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func NewMarkdown(color bool) goldmark.Markdown {
	theme := styles.NoTTYStyleConfig
	t := true

	theme.Document.Margin = new(uint)
	theme.Document.BlockPrefix = ""

	n := "│ "
	theme.BlockQuote.IndentToken = &n

	if color {
		theme.CodeBlock.Theme = "nord"
	}

	theme.H1.Prefix = ""
	theme.H1.Underline = &t
	theme.H1.Bold = &t

	theme.Code.Bold = &t

	theme.H2.Prefix = ""

	theme.Strong.BlockPrefix = ""
	theme.Strong.BlockSuffix = ""
	theme.Strong.Bold = &t

	theme.Emph.BlockPrefix = ""
	theme.Emph.BlockSuffix = ""
	theme.Emph.Italic = &t

	ar := ansi.NewRenderer(ansi.Options{
		Styles:   theme,
		WordWrap: style.MaxTextWidth(),
	})
	md := goldmark.New(
		goldmark.WithExtensions(&fences.Extender{}),
		goldmark.WithParserOptions(parser.WithASTTransformers(
			util.Prioritized(&FlattenFencedContainers{}, 1),
			util.Prioritized(&RewriteDefinitionLists{}, 2),
		)),
		goldmark.WithRenderer(
			renderer.NewRenderer(renderer.WithNodeRenderers(
				util.Prioritized(ar, 900),
			)),
		),
	)

	return md
}

// FlattenFencedContainers moves children of FencedContainer nodes to the
// same level as the container, then removes the empty container.
//
// Before:
//
//	Document {
//	    Heading { "Examples" }
//	    FencedContainer {
//	        Heading { "usage example" }
//	        FencedCodeBlock { ... }
//	    }
//	}
//
// After:
//
//	Document {
//	    Heading { "Examples" }
//	    Heading { "usage example" }
//	    FencedCodeBlock { ... }
//	}
//
// This is needed because glamour library can't handle them and prints garbage like `::{.example}`
// and `<div>...`
type FlattenFencedContainers struct{}

func (t *FlattenFencedContainers) Transform(node *ast.Document, rd text.Reader, pc parser.Context) {
	var containers []ast.Node
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == fences.KindFencedContainer {
			containers = append(containers, child)
		}
	}

	for _, container := range containers {
		parent := container.Parent()
		ref := container // insert children after this node, advancing as we go

		// Move each child out of the container.
		for child := container.FirstChild(); child != nil; {
			next := child.NextSibling()
			parent.InsertAfter(parent, ref, child)
			ref = child
			child = next
		}

		// Remove the now-empty container.
		parent.RemoveChild(parent, container)
	}
}

// RewriteDefinitionLists rewrites definition list patterns under "# Inputs" headings.
//
// It turns this AST structure:
//
//	Paragraph{ CodeSpan("x") }
//	Paragraph{ Text(": description") }
//
// Into a List where each item contains the term (with ":") and description:
//
//	List {
//	    ListItem {
//	        Paragraph{ CodeSpan("x"), Text(":") }
//	        Paragraph{ Text("description") }
//	    }
//	}
type RewriteDefinitionLists struct{}

func (t *RewriteDefinitionLists) Transform(node *ast.Document, rd text.Reader, pc parser.Context) {
	source := rd.Source()
	inInputs := false

	type defPair struct {
		term *ast.Paragraph
		def  *ast.Paragraph
	}
	var pairs []defPair

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if child.Kind() == ast.KindHeading {
			heading := child.(*ast.Heading)
			htext := string(heading.Lines().Value(source))
			inInputs = htext == "Inputs" || htext == "Input"
			continue
		}

		if !inInputs {
			continue
		}
		if child.Kind() != ast.KindParagraph {
			continue
		}

		next := child.NextSibling()
		if next == nil || next.Kind() != ast.KindParagraph {
			continue
		}

		firstText := findFirstText(next)
		if firstText == nil {
			continue
		}
		seg := firstText.Segment
		content := seg.Value(source)
		if len(content) < 2 || content[0] != ':' || content[1] != ' ' {
			continue
		}

		pairs = append(pairs, defPair{
			term: child.(*ast.Paragraph),
			def:  next.(*ast.Paragraph),
		})
	}

	for _, p := range pairs {
		firstText := findFirstText(p.def)

		// Create a text node pointing to the ":" character in the source.
		colonSeg := firstText.Segment
		colonSeg.Stop = colonSeg.Start + 1 // just the ":"
		colonText := ast.NewTextSegment(colonSeg)
		p.term.AppendChild(p.term, colonText)

		// Strip ": " prefix from the definition's first text node.
		firstText.Segment.Start += 2
	}
}

func findFirstText(n ast.Node) *ast.Text {
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			return t
		}
	}
	return nil
}
