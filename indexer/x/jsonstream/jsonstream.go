package jsonstream

import (
	"encoding/json/jsontext"
	"fmt"
	"io"
)

// ParsePackages  parses packages.json file of the format below
//
//	{
//	  "packages": {
//	    "pkg1": { ... },
//	    "pkg2": { ... }
//	  },
//	  ...
//	}
//
// It allows to not wait for the entire file to be parsed and work with
// package descriptions in the stream manner.
func ParsePackages(pkgs io.Reader, cb func(name string, content []byte) error) error {
	dec := jsontext.NewDecoder(pkgs)

	// We're here
	// ↓
	// { "other": ..., "packages": { "pkg1": {...}, "pkg2": {...} } }
	t, err := dec.ReadToken()
	if err != nil {
		return fmt.Errorf("read opening curly bracket: %w", err)
	}

	for {
		//   ↓ (1)         ↓ (2)
		// { "other": ..., "packages": { "pkg1": {...}, "pkg2": {...} } }
		t, err = dec.ReadToken()
		if err != nil {
			return fmt.Errorf("read root key: %w", err)
		}

		if t.Kind() != '"' {
			return fmt.Errorf("expected root key as string, but got %s", t.Kind())
		}
		if t.String() == "packages" {
			break
		}

		dec.SkipValue()
	}

	//                             ↓ consume the opening '{'
	// { "other": ..., "packages": { "pkg1": {...}, "pkg2": {...} } }
	t, err = dec.ReadToken()
	if err != nil {
		return fmt.Errorf("read opening bracket for 'packages': %w", err)
	}

	var name string
	var content []byte

	for {
		//                               ↓ (1)          ↓ (2)
		// { "other": ..., "packages": { "pkg1": {...}, "pkg2": {...} } }
		t, err = dec.ReadToken()
		if err != nil {
			return fmt.Errorf("read package name: %w", err)
		}

		switch t.Kind() {
		case '"':
			name = t.String()
		case '}':
			//                                                            ↓
			// { "other": ..., "packages": { "pkg1": {...}, "pkg2": {...} } }
			return nil
		default:
			return fmt.Errorf("expected package name as string, but got %s", t.Kind())
		}

		content, err = dec.ReadValue()
		if err != nil {
			return fmt.Errorf("read package content: %w", err)
		}

		if err = cb(name, content); err != nil {
			return fmt.Errorf("callback failed: %w", err)
		}
	}
}
