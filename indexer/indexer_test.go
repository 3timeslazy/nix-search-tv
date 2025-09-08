package indexer

import (
	"testing"
)

func TestInjectKey(t *testing.T) {

	// t.Run("simple", func(t *testing.T) {
	// 	pkg := []byte(`{ "version": "v1.0.0" }`)
	// 	pkg = injectKey("nix-search-tv", pkg)
	// 	assert.Equal(t, []byte(`{"_key":"nix-search-tv", "version": "v1.0.0" }`), pkg)
	// })

	// t.Run("with quotes", func(t *testing.T) {
	// 	pkg := []byte(`{ "version": "v1.0.0" }`)
	// 	pkg = injectKey(`package."with quotes"`, pkg)
	// 	assert.Equal(t, []byte(`{"_key":"package.\"with quotes\"", "version": "v1.0.0" }`), pkg)
	// })
}
