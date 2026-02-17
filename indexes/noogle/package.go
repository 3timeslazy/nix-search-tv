package noogle

import (
	"fmt"
	"strings"
)

type Package struct {
	Meta    Meta     `json:"meta"`
	Content *Content `json:"content"`

	// Not a part of the original noogle structure,
	// set by fetcher
	NixpkgsCommit string `json:"nixpkgs_commit"`
}

type Meta struct {
	Title          string     `json:"title"`
	Aliases        [][]string `json:"aliases"`
	Signature      string     `json:"signature"`
	LambdaExpr     string     `json:"lambda_expr"`
	IsFunctor      bool       `json:"is_functor"`
	AttrPosition   *Position  `json:"attr_position"`
	LambdaPosition *Position  `json:"lambda_position"`
}

type Content struct {
	Content string  `json:"content"`
	Source  *Source `json:"source"`
}

type Source struct {
	Position *Position `json:"position"`
}

type Position struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

func (pkg *Package) GetSource() string {
	var pos *Position
	if pkg.Meta.LambdaPosition != nil {
		pos = pkg.Meta.LambdaPosition
	}
	if pkg.Meta.AttrPosition != nil {
		pos = pkg.Meta.AttrPosition
	}
	if pkg.Content != nil &&
		pkg.Content.Source != nil &&
		pkg.Content.Source.Position != nil {
		pos = pkg.Content.Source.Position
	}

	path := ""
	if pos != nil {
		segments := strings.Split(pos.File, "/")
		if len(segments) > 4 {
			if segments[1] == "nix" &&
				segments[2] == "store" &&
				strings.Contains(segments[3], "source") {
				path = strings.Join(segments[4:], "/")
			}
		}
	}

	if fn, ok := strings.CutPrefix(pkg.Meta.Title, "builtins."); ok {
		return fmt.Sprintf("https://github.com/search?q=repo:NixOS/nix+symbol:prim_%s&type=code", fn)
	}
	if pkg.Meta.IsFunctor {
		if pos == nil || path == "" {
			return pkg.GetHomepage()
		}
		return fmt.Sprintf(
			"https://github.com/NixOS/nixpkgs/blob/%s/%s#L%d:C%d",
			pkg.NixpkgsCommit,
			path,
			pos.Line,
			pos.Column,
		)
	}
	if pos != nil && path != "" {
		return fmt.Sprintf(
			"https://github.com/NixOS/nixpkgs/blob/%s/%s#L%d:C%d",
			pkg.NixpkgsCommit,
			path,
			pos.Line,
			pos.Column,
		)
	}

	return pkg.GetHomepage()
}

func (pkg *Package) GetHomepage() string {
	path := strings.ReplaceAll(pkg.Meta.Title, ".", "/")
	return fmt.Sprintf("https://noogle.dev/f/%s", path)
}
