package style

import (
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestNoColor(t *testing.T) {
	styler := TextStyle

	// Test with NO_COLOR not set
	assert.NoError(t, os.Unsetenv("NO_COLOR"))
	styler = detectStyle()

	assert.Equal(t, "\x1b[1mtest\x1b[22m", styler.Bold("test"))
	assert.Equal(t, "\x1b[2mtest\x1b[22m", styler.Dim("test"))
	assert.Equal(t, "\x1b[31mtest\x1b[0m", styler.Red("test"))

	// Test with NO_COLOR set to any value
	assert.NoError(t, os.Setenv("NO_COLOR", "1"))
	styler = detectStyle()

	assert.Equal(t, "test", styler.Bold("test"))
	assert.Equal(t, "test", styler.Dim("test"))
	assert.Equal(t, "test", styler.Red("test"))

	// Test with NO_COLOR set to "0" or "false" (should still disable colors)
	for _, val := range []string{"0", "false", "anything"} {
		assert.NoError(t, os.Setenv("NO_COLOR", val))
		styler = detectStyle()

		assert.Equal(t, "test", styler.Bold("test"), "NO_COLOR=%s should disable colors", val)
		assert.Equal(t, "test", styler.Dim("test"), "NO_COLOR=%s should disable colors", val)
	}

	// Reset to empty string (colors enabled again)
	assert.NoError(t, os.Setenv("NO_COLOR", ""))
	styler = detectStyle()

	assert.Equal(t, "\x1b[1mtest\x1b[22m", styler.Bold("test"))
	assert.Equal(t, "\x1b[2mtest\x1b[22m", styler.Dim("test"))
}

func TestNoColorStyleTextBlock(t *testing.T) {
	styler := TextStyle

	// Test with NO_COLOR not set
	assert.NoError(t, os.Unsetenv("NO_COLOR"))

	styler = detectStyle()

	assert.Equal(t, "\x1b[2mline1\x1b[22m\n\x1b[2mline2\x1b[22m", styler.Dim("line1\nline2"))

	// Test with NO_COLOR set
	assert.NoError(t, os.Setenv("NO_COLOR", "1"))

	styler = detectStyle()

	assert.Equal(t, "line1\nline2", styler.Dim("line1\nline2"))

	// Reset
	assert.NoError(t, os.Setenv("NO_COLOR", ""))

	styler = detectStyle()

}
