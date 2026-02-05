package cmd

import (
	"bufio"
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/3timeslazy/nix-search-tv/config"
	"github.com/3timeslazy/nix-search-tv/indexer"

	"github.com/urfave/cli/v3"
)

const OfflineFlag = "offline"

var Print = &cli.Command{
	Name:      "print",
	UsageText: "nix-search-tv print",
	Usage:     "Print indexed package names. If there is no indexed packages, they'll get indexed first",
	Action:    PrintAction,
	Flags: append(BaseFlags(), &cli.BoolFlag{
		Name:  OfflineFlag,
		Usage: "disable fetching new indexes",
	}),
}

func PrintAction(ctx context.Context, cmd *cli.Command) error {
	conf, err := GetConfig(cmd)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	available, err := SetupIndexes(conf)
	if err != nil {
		return fmt.Errorf("register fetchers: %w", err)
	}

	requested := available
	if cmd.IsSet(IndexesFlag) {
		flags := cmd.StringSlice(IndexesFlag)

		requested = slices.DeleteFunc(requested, func(index string) bool {
			return !slices.Contains(flags, index)
		})
	} else {
		requested = slices.DeleteFunc(requested, func(index string) bool {
			builtin := slices.Contains(conf.Indexes, index)
			_, renderDocs := conf.Experimental.RenderDocsIndexes[index]
			_, optionsFile := conf.Experimental.OptionsFile[index]
			return !builtin && !renderDocs && !optionsFile
		})
	}

	indexes, err := GetIndexes(conf.CacheDir, requested)
	if err != nil {
		return fmt.Errorf("get indexes: %w", err)
	}

	needIndexing, err := indexer.NeedIndexing(
		conf.CacheDir,
		time.Duration(conf.UpdateInterval),
		indexes,
	)
	if err != nil {
		return fmt.Errorf("check if indexing needed: %w", err)
	}

	if cmd.IsSet(OfflineFlag) {
		needIndexing = nil
	}

	if len(needIndexing) > 0 {
		if conf.EnableWaitingMessage {
			PrintWaiting(Stdout)
		}
	}

	withPrefix := len(indexes) > 1

	for _, index := range indexes {
		canPrint := !slices.ContainsFunc(needIndexing, func(need indexer.Index) bool {
			return need.Name == index.Name
		})
		if canPrint {
			err = PrintIndexKeys(conf, index.Name, withPrefix)
			if err != nil {
				return fmt.Errorf("%s: %w", index, err)
			}
		}
	}

	results := indexer.RunIndexing(ctx, conf.CacheDir, needIndexing)
	for result := range results {
		if result.Err != nil {
			msg := addIndexPrefix(
				result.Index,
				fmt.Sprintf("indexing failed: %s\n", result.Err),
			)
			Stdout.Write([]byte(msg))
			continue
		}

		err := PrintIndexKeys(conf, result.Index, withPrefix)
		if err != nil {
			return fmt.Errorf("%s: %w", result.Index, err)
		}
	}

	return nil
}

func PrintIndexKeys(conf config.Config, index string, withPrefix bool) error {
	keys, err := indexer.OpenKeysReader(conf.CacheDir, index)
	if err != nil {
		return fmt.Errorf("read keys file: %w", err)
	}
	defer keys.Close()

	allkeys := []string{}
	scanner := bufio.NewScanner(keys)
	for scanner.Scan() {
		allkeys = append(allkeys, scanner.Text())
	}

	prefix := []byte{}
	if withPrefix {
		prefix = []byte(index + "/ ")
	}

	slices.Sort(allkeys)

	for _, k := range allkeys {
		Stdout.Write(append([]byte(prefix), []byte(k)...))
		Stdout.Write([]byte{'\n'})
	}

	return nil
}
