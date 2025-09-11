package jsonstream

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestUnmarshalNixpkgs(t *testing.T) {
	input := bytes.NewBufferString(`
	{
	  "version": 1,
	  "packages": {
	    "pkg1": { "mainProgram": "pkg1" },
		"pkg2": { "mainProgram": "pkg2" }
	  }
	}
	`)

	expected := map[string]any{
		"pkg1": map[string]any{"mainProgram": "pkg1"},
		"pkg2": map[string]any{"mainProgram": "pkg2"},
	}
	actual := map[string]any{}

	err := ParsePackages(input, func(k string, v []byte) error {
		pkg := map[string]any{}
		err := json.Unmarshal(v, &pkg)
		if err != nil {
			return err
		}

		actual[k] = pkg
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
