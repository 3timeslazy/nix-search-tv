package nixpkgs

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/3timeslazy/nix-search-tv/indexer"
)

type Package struct {
	indexer.Package
	Meta       Meta           `json:"meta"`
	Version    string         `json:"version"`
	Outputs    map[string]any `json:"outputs"`
	OutputName string         `json:"outputName"`
}

type Meta struct {
	Description     string               `json:"description"`
	LongDescription string               `json:"longDescription"`
	MainProgram     string               `json:"mainProgram"`
	Homepages       ElemOrSlice[string]  `json:"homepage"`
	Licenses        ElemOrSlice[License] `json:"license"`
	Broken          bool                 `json:"broken"`
	Unfree          bool                 `json:"unfree"`
	Name            string               `json:"name"`
	Position        string               `json:"position"`
	Platforms       []string             `json:"platforms"`
}

type License struct {
	Free     bool   `json:"free"`
	FullName string `json:"fullName"`
	SpdxID   string `json:"spdxId"`
}

type LicenseNoUnmarshal struct {
	Free     bool   `json:"free"`
	FullName string `json:"fullName"`
	SpdxID   string `json:"spdxId"`
}

func (l *License) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data is empty")
	}

	switch {
	case data[0] == '"':
		s := ""
		err := json.Unmarshal(data, &s)
		if err != nil {
			return fmt.Errorf("unmarshal string: %w", err)
		}
		(*l).FullName = s

	default:
		lu := LicenseNoUnmarshal{}
		err := json.Unmarshal(data, &lu)
		if err != nil {
			return fmt.Errorf("unmarshal struct: %w", err)
		}
		*l = License(lu)
	}

	return nil
}

type ElemOrSlice[T any] []T

func (eos *ElemOrSlice[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("input data is empty")
	}

	switch {
	case data[0] == '[':
		s := []T{}
		err := json.Unmarshal(data, &s)
		if err != nil {
			return fmt.Errorf("unmarshal slice: %w", err)
		}
		*eos = s

	default:
		var e T
		err := json.Unmarshal(data, &e)
		if err != nil {
			return fmt.Errorf("unmarshal element: %w", err)
		}
		*eos = []T{e}
	}

	return nil
}

func (pkg *Package) GetVersion() string {
	if pkg.Version != "" {
		return pkg.Version
	}

	// TODO: comment about nvidia-docker
	return strings.TrimPrefix(pkg.Meta.Name, pkg.Name+"-")
}
