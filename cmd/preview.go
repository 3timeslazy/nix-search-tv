package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/3timeslazy/nix-search-tv/indexes/indices"

	"github.com/urfave/cli/v3"
)

var Preview = &cli.Command{
	Name:      "preview",
	UsageText: "nix-search-tv preview [package_name]",
	Usage:     "Print preview for the package",
	Action:    PreviewAction,
	Flags:     BaseFlags(),
}

func PreviewAction(ctx context.Context, cmd *cli.Command) error {
	fullPkgName := strings.Join(cmd.Args().Slice(), " ")
	if fullPkgName == "" {
		return errors.New("package name is required")
	}

	conf, err := GetConfig(cmd)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}
	if fullPkgName == waitingMessage {
		PreviewWaiting(os.Stdout, conf)
		return nil
	}

	if len(conf.Indexes) == 1 {
		preview := indices.Previews[conf.Indexes[0]]
		return preview(conf, fullPkgName)
	}

	ind, pkgName, ok := cutIndexPrefix(fullPkgName)
	if !ok {
		return errors.New("multiple indexes requested, but the package has no index prefix")
	}

	preview := indices.Previews[ind]
	return preview(conf, pkgName)
}
