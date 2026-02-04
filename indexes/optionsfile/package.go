package optionsfile

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/indexes/textutil"
	"github.com/3timeslazy/nix-search-tv/style"
)

type Package struct {
	indexer.Package
	Type         string   `json:"type"`
	Description  string   `json:"description"`
	Example      String   `json:"example"`
	Declarations []String `json:"declarations"`
	Default      String   `json:"default"`
}

func (pkg *Package) Preview(out io.Writer) {
	styler := style.TextStyle

	pkgTitle := textutil.PkgName(pkg.Name) + "\n"
	fmt.Fprint(out, pkgTitle)

	desc := style.StyleLongDescription(styler, string(pkg.Description))
	desc += "\n"
	fmt.Fprintln(out, desc)

	pkgType := textutil.Prop("type", "", pkg.Type)
	fmt.Fprintln(out, pkgType)

	if def := string(pkg.Default); def != "" {
		def = textutil.Prop(
			"default", "",
			style.PrintCodeBlock(def),
		)
		fmt.Fprintln(out, def)
	}

	if example := string(pkg.Example); example != "" {
		example = textutil.Prop(
			"example", "",
			style.PrintCodeBlock(example),
		)
		fmt.Fprintln(out, example)
	}
}

func (pkg *Package) GetSource() string {
	if len(pkg.Declarations) > 0 {
		return string(pkg.Declarations[0])
	}

	return ""
}

func (pkg *Package) GetHomepage() string {
	return pkg.GetSource()
}

// String is type that can be decoded from either a string, or an object with
// certain fields often used in options.json files e.g. text, url
type String string

func (s *String) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	if data[0] == '"' {
		v := ""
		err := json.Unmarshal(data, &v)
		if err != nil {
			return fmt.Errorf("unmarshal into string: %w", err)
		}
		(*s) = String(v)
		return nil
	}

	text := struct {
		// `text` usually used by examples and default values
		Text string `json:"text"`

		// `url` usually used by declarations
		URL string `json:"url"`
	}{}
	err := json.Unmarshal(data, &text)
	if err != nil {
		return fmt.Errorf("unmarshal into struct: %w", err)
	}

	switch {
	case text.Text != "":
		(*s) = String(text.Text)

	case text.URL != "":
		(*s) = String(text.URL)
	}

	return nil
}
