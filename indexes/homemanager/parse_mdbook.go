package homemanager

import (
	"strings"

	"github.com/3timeslazy/nix-search-tv/pkgs/renderdocs"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// ParseMdBook parses the Home Manager options from the single-page
// mdBook "print" output (print.html).
//
// The Home Manager documentation moved from the nixos-render-docs
// `options.xhtml` (a `dl.variablelist` of `dt`/`dd` pairs, still used
// by nix-darwin and parsed by renderdocs.Parse) to an mdBook site.
// In the mdBook output each option is an `<h2 id="opt-NAME">` heading
// followed, until the next heading, by paragraphs describing the
// option, e.g.:
//
//	<h2 id="opt-programs.git.enable"><a ...>programs.git.enable</a></h2>
//	<p>Whether to enable Git.</p>
//	<p><em>Type:</em>
//	boolean</p>
//	<p><em>Default:</em></p>
//	<pre><code class="language-nix">false
//	</code></pre>
//	<p><em>Example:</em></p>
//	<pre><code class="language-nix">true
//	</code></pre>
//	<p><em>Declared by:</em></p>
//	<ul>
//	  <li><a href="...">&lt;home-manager/modules/programs/git.nix&gt;</a></li>
//	</ul>
func ParseMdBook(doc *html.Node) (map[string]renderdocs.Package, error) {
	pkgs := map[string]renderdocs.Package{}

	headings := htmlquery.Find(doc, `//h2[starts-with(@id, "opt-")]`)
	for _, h := range headings {
		name := strings.TrimSpace(htmlquery.InnerText(h))
		if name == "" {
			continue
		}

		pkg := renderdocs.Package{Name: name}
		var descParts []string

		// Walk siblings until the next option heading (or the end of
		// the section).
		for n := h.NextSibling; n != nil; n = n.NextSibling {
			if n.Type == html.ElementNode && (n.Data == "h2" || n.Data == "h1") {
				break
			}
			if n.Type != html.ElementNode {
				continue
			}

			prop, inlineValue, ok := splitProperty(n)
			if !ok {
				// Not a Type/Default/... marker: part of the description.
				descParts = append(descParts, htmlquery.OutputHTML(n, true))
				continue
			}

			value := inlineValue
			if value == "" {
				// Value lives in the following <pre>/<ul> sibling.
				value = takeFollowingValue(prop, &n)
			}
			setProp(&pkg, prop, value)
		}

		pkg.Description = renderdocs.RenderHTML(strings.Join(descParts, "\n"))
		pkgs[name] = pkg
	}

	return pkgs, nil
}

// splitProperty reports whether node is a `<p><em>PROP:</em> value</p>`
// marker. It returns the canonical property label (as found in
// renderdocs, e.g. "Type:") and any inline value present in the same
// paragraph.
func splitProperty(n *html.Node) (prop string, inlineValue string, ok bool) {
	if n.Data != "p" {
		return "", "", false
	}
	em := htmlquery.FindOne(n, `/em`)
	if em == nil {
		return "", "", false
	}

	label := strings.TrimSpace(htmlquery.InnerText(em))
	if !knownProps[label] {
		return "", "", false
	}

	// Everything in the paragraph except the <em> label is an inline value
	// (e.g. `<p><em>Type:</em> boolean</p>`).
	full := strings.TrimSpace(htmlquery.InnerText(n))
	inline := strings.TrimSpace(strings.TrimPrefix(full, label))
	return label, inline, true
}

// takeFollowingValue extracts the value stored in the sibling that
// follows a property paragraph: a `<pre>` (for Default/Example/Type) or
// a `<ul>` of links (for Declared by). It advances *np past the consumed
// sibling.
func takeFollowingValue(prop string, np **html.Node) string {
	next := nextElement((*np).NextSibling)
	if next == nil {
		return ""
	}

	switch {
	case next.Data == "pre":
		*np = next
		return renderdocs.NormProp(htmlquery.InnerText(next))

	case next.Data == "ul" && prop == renderdocs.PropDeclaredBy:
		*np = next
		links := htmlquery.Find(next, `//li`)
		parts := make([]string, 0, len(links))
		for _, li := range links {
			parts = append(parts, renderdocs.NormProp(htmlquery.InnerText(li)))
		}
		return strings.Join(parts, "\n")
	}

	return ""
}

func nextElement(n *html.Node) *html.Node {
	for ; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode {
			return n
		}
	}
	return nil
}

var knownProps = map[string]bool{
	renderdocs.PropType:       true,
	renderdocs.PropDefault:    true,
	renderdocs.PropExample:    true,
	renderdocs.PropDeclaredBy: true,
}

func setProp(pkg *renderdocs.Package, name, value string) {
	switch name {
	case renderdocs.PropType:
		pkg.Type = value
	case renderdocs.PropDefault:
		pkg.Default = value
	case renderdocs.PropExample:
		pkg.Example = value
	case renderdocs.PropDeclaredBy:
		pkg.DeclaredBy = append(pkg.DeclaredBy, value)
	}
}
