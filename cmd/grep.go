package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/3timeslazy/nix-search-tv/indexer"
	"github.com/3timeslazy/nix-search-tv/indexer/x/jsonstream"

	"github.com/urfave/cli/v3"
)

var Grep = &cli.Command{
	Name:      "grep",
	UsageText: "nix-search-tv grep [package_name]",
	Usage:     "Print the the link to the source code",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		wordsStr := cmd.Args().First()
		if wordsStr == "" {
			return errors.New("search words required")
		}

		conf, err := GetConfig(cmd)
		if err != nil {
			return fmt.Errorf("get config: %w", err)
		}

		ind, err := indexer.NewBadger(indexer.BadgerConfig{
			Dir:      conf.CacheDir + "/nixpkgs/badger",
			ReadOnly: true,
		})
		if err != nil {
			return fmt.Errorf("new badger: %w", err)
		}
		defer ind.Close()

		wordsStr = strings.ToLower(wordsStr)
		words := strings.Fields(wordsStr)

		err = ind.IterAll(func(v []byte) bool {
			desc, _ := jsonstream.FindPath(bytes.NewReader(v), "meta.description")
			desc = strings.ToLower(desc)

			longDesc, _ := jsonstream.FindPath(bytes.NewReader(v), "meta.longDescription")
			longDesc = strings.ToLower(longDesc)

			haystack := desc + "\n" + longDesc

			all := true
			for _, w := range words {
				all = all && strings.Contains(haystack, w)
			}
			return all
		})
		if err != nil {
			return fmt.Errorf("all has: %w", err)
		}

		return nil
	},
	Flags: BaseFlags(),
}
