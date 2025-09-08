package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/indexes/indices"

	"github.com/urfave/cli/v3"
)

var Preview = &cli.Command{
	Name:      "preview",
	UsageText: "nix-search-tv preview [package_name]",
	Usage:     "Print preview for the package",
	Action:    NewPreviewAction(indices.Preview),
	Flags:     BaseFlags(),
}

type PreviewFunc func(index string, out io.Writer, pkg json.RawMessage) error

func NewPreviewAction(preview PreviewFunc) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		fullPkgName := strings.Join(cmd.Args().Slice(), " ")
		if fullPkgName == "" {
			return errors.New("package name is required")
		}

		conf, err := GetConfig(cmd)
		if err != nil {
			return fmt.Errorf("get config: %w", err)
		}
		if fullPkgName == waitingMessage {
			PreviewWaiting(Stdout, conf)
			return nil
		}

		if cmd.IsSet(IndexesFlag) {
			conf.Indexes = cmd.StringSlice(IndexesFlag)
		}

		_, err = SetupIndexes(conf)
		if err != nil {
			return err
		}

		var index, pkgName string

		if len(conf.Indexes) == 1 {
			index = conf.Indexes[0]
			pkgName = fullPkgName
		} else {
			var ok bool
			index, pkgName, ok = cutIndexPrefix(fullPkgName)
			if !ok {
				return errors.New("multiple indexes requested, but the package has no index prefix")
			}
		}

		pkg, err := indexer.LoadKey(conf.CacheDir, index, pkgName)
		if err != nil {
			return fmt.Errorf("load package content: %w", err)
		}

		pkg = injectKey(pkgName, pkg)
		return preview(index, Stdout, pkg)
	}
}

// injectKey appends the `_key` field into the json object.
//
// This thing saves about ~2.5s on my laptop when indexing 120k nix packages
func injectKey(key string, pkg json.RawMessage) json.RawMessage {
	return append([]byte(`{"_key":`+strconv.Quote(key)+`,`), pkg[1:]...)
}
