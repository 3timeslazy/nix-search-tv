package integrations_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/jubnzv/go-tmux"
)

func TestIntegrations(t *testing.T) {
	binPath := os.Getenv("NIX_SEARCH_TV_BIN")
	if binPath == "" {
		t.Skip()
	}
	dataDir, err := filepath.Abs("./testdata")
	assert.NoError(t, err)
	cacheDir := filepath.Join(dataDir, "cache")

	expected := func(t *testing.T, name string) string {
		expectedPath := filepath.Join(dataDir, "cases", name+".txt")
		data, err := os.ReadFile(expectedPath)
		assert.NoError(t, err)
		return string(data)
	}
	fzfPreview := func(searchFlags string) string {
		return fmt.Sprintf("%s preview %s {1}{2}", binPath, searchFlags)
	}
	tvPreview := func(searchFlags string) string {
		return fmt.Sprintf("echo {} | awk \"{ print $1$2 }\" | xargs %s preview %s", binPath, searchFlags)
	}

	t.Run("single", func(t *testing.T) {
		searchFlags := "--indexes nixpkgs --cache-dir " + cacheDir

		t.Run("fzf", func(t *testing.T) {
			name := "fzf-single"
			testIntegration(t,
				CmdArgs{
					Name:        name,
					Tool:        "fzf",
					ToolPreview: fzfPreview(searchFlags),
					SearchBin:   binPath,
					SearchFlags: searchFlags,
				},
				expected(t, name),
			)
		})
		t.Run("tv", func(t *testing.T) {
			name := "tv-single"
			testIntegration(t,
				CmdArgs{
					Name:        name,
					Tool:        "tv",
					ToolPreview: tvPreview(searchFlags),
					SearchBin:   binPath,
					SearchFlags: searchFlags,
				},
				expected(t, name),
			)
		})
	})

	t.Run("miltiple", func(t *testing.T) {
		searchFlags := "--indexes nixpkgs,home-manager --cache-dir " + cacheDir

		t.Run("fzf", func(t *testing.T) {
			name := "fzf-multiple"
			testIntegration(t,
				CmdArgs{
					Name:        name,
					Tool:        "fzf",
					ToolPreview: fzfPreview(searchFlags),
					SearchBin:   binPath,
					SearchFlags: searchFlags,
				},
				expected(t, name),
			)
		})
		t.Run("tv", func(t *testing.T) {
			name := "tv-multiple"
			testIntegration(t,
				CmdArgs{
					Name:        name,
					Tool:        "tv",
					ToolPreview: tvPreview(searchFlags),
					SearchBin:   binPath,
					SearchFlags: searchFlags,
				},
				expected(t, name),
			)
		})
	})
}

type CmdArgs struct {
	Name string

	// Template fields
	Tool        string
	ToolPreview string
	SearchBin   string
	SearchFlags string
}

var cmdTmpl = template.Must(template.New("").Parse(
	"{{ .SearchBin }} print {{ .SearchFlags }} | {{ .Tool }} --preview '{{ .ToolPreview }}'",
))

func testIntegration(t *testing.T, cmdArgs CmdArgs, expected string) {
	t.Parallel()

	srv := new(tmux.Server)
	pane := newSession(t, srv, cmdArgs.Name)

	cmd := bytes.NewBuffer(nil)
	err := cmdTmpl.Execute(cmd, cmdArgs)
	assert.NoError(t, err)

	err = pane.RunCommand(cmd.String())
	assert.NoError(t, err)

	// A good candidate for a flaky test,
	// but we have to wait for the fuzzy tool
	// to print something and I don't know a better
	// way yet
	time.Sleep(2 * time.Second)

	output, err := pane.Capture()
	assert.NoError(t, err)

	assert.Equal(t, expected, output, "Command: ", cmd.String())
}

func newSession(
	t *testing.T,
	srv *tmux.Server,
	sessName string,
) *tmux.Pane {
	sess := tmux.Session{
		Name: sessName,
		Windows: []tmux.Window{{
			Id:   1,
			Name: sessName + "-window",
		}},
	}
	srv.AddSession(sess)
	t.Cleanup(func() {
		srv.KillSession(sess.Name)
	})

	conf := tmux.Configuration{
		Server:        srv,
		Sessions:      []*tmux.Session{&sess},
		ActiveSession: nil,
	}
	err := conf.Apply()
	assert.NoError(t, err)

	panes, err := sess.ListPanes()
	assert.NoError(t, err)
	assert.True(t, len(panes) == 1)

	return &panes[0]
}
