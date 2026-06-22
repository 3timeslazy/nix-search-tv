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
		name := optionName(h)
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

			if inlineValue != "" {
				setProp(&pkg, prop, inlineValue)
				continue
			}
			// Value lives in the following <pre>/<ul> sibling. A
			// "Declared by" list may hold several entries; each becomes a
			// separate declaration.
			for _, value := range takeFollowingValues(prop, &n) {
				setProp(&pkg, prop, value)
			}
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

	// The value is whatever the paragraph holds besides the <em> label.
	// Remove the label node so it does not appear in the value.
	em.Parent.RemoveChild(em)

	// Default and Example values carry inline markdown in options.json:
	// code spans (`<code>true</code>` -> `` `true` ``) and cross-reference
	// links (`<a href="#opt-other"></a>` -> `[](#opt-other)`), with
	// significant whitespace preserved, so render them with RenderValue.
	// Type is plain text and may carry a trailing `<em>(read only)</em>`
	// marker that must not become markdown emphasis, so keep it as text.
	var inline string
	if label == renderdocs.PropType {
		inline = htmlquery.InnerText(n)
	} else {
		inline = renderdocs.RenderValue(n)
	}
	return label, renderdocs.NormProp(inline), true
}

// takeFollowingValues extracts the value(s) stored in the sibling that
// follows a property paragraph: a `<pre>` (for Default/Example/Type),
// which yields a single value, or a `<ul>` of links (for Declared by),
// which yields one value per list item. It advances *np past the
// consumed sibling.
func takeFollowingValues(prop string, np **html.Node) []string {
	next := nextElement((*np).NextSibling)
	if next == nil {
		return nil
	}

	switch {
	case next.Data == "pre":
		*np = next
		return []string{renderdocs.NormProp(htmlquery.InnerText(next))}

	case next.Data == "ul" && prop == renderdocs.PropDeclaredBy:
		*np = next
		// Each declaration is a link whose href is the canonical source
		// URL (the link text is only a display label like
		// <home-manager/modules/...>). options.json stores the href, so
		// take @href to match.
		hrefs := htmlquery.Find(next, `.//li//a/@href`)
		values := make([]string, 0, len(hrefs))
		for _, a := range hrefs {
			values = append(values, renderdocs.NormProp(htmlquery.InnerText(a)))
		}
		return values
	}

	return nil
}

func nextElement(n *html.Node) *html.Node {
	for ; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode {
			return n
		}
	}
	return nil
}

// optionName returns the canonical option name for an `<h2 id="opt-...">`
// heading, matching the key form used by options.json.
//
// Two heading representations are available and each is lossy on its own:
//
//   - The id attribute (opt-<name>) preserves markdown-significant
//     characters such as the doubled underscores in
//     `programs.nvchecker.settings.__config__`, but strips the quotes
//     around dotted attribute-set keys
//     (id: ...defaults.com.apple.dock.tilesize).
//   - The heading text preserves those quotes (rendered as typographic
//     “ ”), but markdown rendering eats other syntax — e.g. `__config__`
//     becomes a bold "config".
//
// In the Home Manager option set these two cases never co-occur (no key
// has both quotes and markdown-significant characters), so we can pick
// the lossless source per heading: use the id when it carries markdown
// syntax that the text would have mangled, otherwise use the text with
// typographic quotes normalised to the straight quotes options.json uses.
func optionName(h *html.Node) string {
	id := htmlquery.SelectAttr(h, "id")
	id = strings.TrimPrefix(id, "opt-")

	// `__` is the only markdown construct observed in option names; it is
	// preserved verbatim in the id but rendered as bold in the text.
	if strings.Contains(id, "__") {
		return strings.TrimSpace(id)
	}

	// NormProp normalises typographic quotes (and en-dashes / ellipses,
	// which never appear in option paths) to their ASCII equivalents.
	return renderdocs.NormProp(strings.TrimSpace(htmlquery.InnerText(h)))
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
