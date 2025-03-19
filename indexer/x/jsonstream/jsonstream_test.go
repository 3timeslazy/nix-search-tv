package jsonstream

import (
	"bytes"
	"encoding/json"
	"strings"
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

func TestFindPath(t *testing.T) {
	t.Run("top-level property", func(t *testing.T) {
		input := strings.NewReader(`{ "version": "v1" }`)
		v, err := FindPath(input, "version")
		assert.NoError(t, err)
		assert.Equal(t, "v1", v)
	})

	t.Run("nested property", func(t *testing.T) {
		input := strings.NewReader(`{ "meta": { "version": "v1" } }`)
		v, err := FindPath(input, "meta.version")
		assert.NoError(t, err)
		assert.Equal(t, "v1", v)
	})

	t.Run("not exist", func(t *testing.T) {
		input := strings.NewReader(`{ "version": "v1" }`)
		v, err := FindPath(input, "meta")
		assert.Equal(t, "", v)
		assert.NoError(t, err)
	})

	t.Run("not a string", func(t *testing.T) {
		input := strings.NewReader(`{ "version": 1 }`)
		_, err := FindPath(input, "version")
		assert.Error(t, err)
	})
}
