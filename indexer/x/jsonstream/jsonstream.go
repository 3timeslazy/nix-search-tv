package jsonstream

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
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
func ParsePackages(pkgs io.Reader, cb func(pkgName string, pkgContent []byte) error) error {
	dec := json.NewDecoder(pkgs)

	// We're here
	// ↓
	// { "other": ..., "packages": { "pkg1": {...}, "pkg2": {...} } }
	token, err := dec.Token()
	if err != nil {
		return fmt.Errorf("get first token: %w", err)
	}
	if token != json.Delim('{') {
		return errors.New("an object expected")
	}

	for {
		//   ↓ (1)         ↓ (2)
		// { "other": ..., "packages": { "pkg1": {...}, "pkg2": {...} } }
		token, err := dec.Token()
		if err != nil {
			return err
		}

		key, ok := token.(string)
		if ok && key == "packages" {
			break
		}

		s := json.RawMessage{}
		err = dec.Decode(&s)
		if err != nil {
			return err
		}
	}

	//                             ↓ consume the opening '{'
	// { "other": ..., "packages": { "pkg1": {...}, "pkg2": {...} } }
	token, err = dec.Token()
	if err != nil {
		return err
	}

	for {
		//                               ↓ (1)          ↓ (2)
		// { "other": ..., "packages": { "pkg1": {...}, "pkg2": {...} } }
		token, err := dec.Token()
		if err != nil {
			return err
		}
		//                                                            ↓
		// { "other": ..., "packages": { "pkg1": {...}, "pkg2": {...} } }
		if token == json.Delim('}') {
			break
		}

		key, ok := token.(string)
		if !ok {
			return fmt.Errorf("key is not a string: %v", token)
		}

		b := json.RawMessage{}
		err = dec.Decode(&b)
		if err != nil {
			return fmt.Errorf("decode package definition: %w", err)
		}

		err = cb(key, b)
		if err != nil {
			return fmt.Errorf("callback failed: %w", err)
		}
	}

	return nil
}

func FindPath(pkg io.Reader, path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// { a: 1, b: { c: 2 } }
	dec := json.NewDecoder(pkg)

	steps := strings.Split(path, ".")
	for i := range steps {
		// ↓ (1)      ↓ (2)
		// { a: 1, b: { c: 2 } }
		token, err := dec.Token()
		if err != nil {
			return "", fmt.Errorf("input is not a valid json object: %w", err)
		}

		for {
			//   ↓ (1)   ↓ (2)  ↓ (3)
			// { a: 1,   b: {   c: 2 } }
			token, err = dec.Token()
			if err != nil {
				return "", fmt.Errorf("get next key: %w", err)
			}
			if token == json.Delim('}') {
				break
			}

			key, ok := token.(string)
			if !ok {
				return "", fmt.Errorf("unexpected token type %T", token)
			}
			if key != steps[i] {
				err = dec.Decode(&json.RawMessage{})
				if err != nil {
					return "", fmt.Errorf("consume value: %w", err)
				}
				continue
			}

			if i == len(steps)-1 {
				v := ""
				if err = dec.Decode(&v); err != nil {
					return "", fmt.Errorf("consume string: %w", err)
				}
				return v, nil
			}

			break
		}
	}

	return "", nil
}
