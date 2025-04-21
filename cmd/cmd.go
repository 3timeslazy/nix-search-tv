package cmd

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/3timeslazy/nix-search-tv/config"
	"github.com/3timeslazy/nix-search-tv/indexes/indices"
	"github.com/3timeslazy/nix-search-tv/indexes/nuschtos"
	"github.com/3timeslazy/nix-search-tv/indexes/renderdocs"

	"github.com/urfave/cli/v3"
)

func BaseFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  ConfigFlag,
			Usage: "path to the configuration file",
			Validator: func(path string) error {
				if path == "" {
					return errors.New("config path cannot be empty")
				}
				return nil
			},
		},
		&cli.StringSliceFlag{
			Name:  IndexesFlag,
			Usage: "what packages to index",
		},
		&cli.StringFlag{
			Name:   CacheDirFlag,
			Hidden: true,
			Usage:  "Path to the indexes cache directory",
		},
	}
}

const (
	ConfigFlag   = "config"
	IndexesFlag  = "indexes"
	CacheDirFlag = "cache-dir"
)

var Stdout io.ReadWriter = os.Stdout

func GetConfig(cmd *cli.Command) (config.Config, error) {
	var conf config.Config
	var err error

	if cmd.IsSet(ConfigFlag) {
		conf, err = config.LoadPath(cmd.String(ConfigFlag))
	} else {
		conf, err = config.LoadDefault()
	}
	if err != nil {
		return config.Config{}, fmt.Errorf("load config: %w", err)
	}

	if cmd.IsSet(IndexesFlag) {
		conf.Indexes = cmd.StringSlice(IndexesFlag)
	}
	if cmd.IsSet(CacheDirFlag) {
		conf.CacheDir = cmd.String(CacheDirFlag)
	}

	if err = validateIndexes(conf.Indexes); err != nil {
		return config.Config{}, err
	}

	if err := os.MkdirAll(conf.CacheDir, 0755); err != nil {
		return conf, fmt.Errorf("cannot create cache directory: %w", err)
	}

	return conf, nil
}

func validateIndexes(indexNames []string) error {
	for _, ind := range indexNames {
		if !indices.BuiltinIndexes[ind] {
			avail := slices.Collect(maps.Keys(indices.BuiltinIndexes))
			str := strings.Join(avail, "\n  ")
			return fmt.Errorf("unknown index %q. Valid values are:\n  %s", ind, str)
		}
	}
	return nil
}

func RegisterRenderDocs(conf config.Config) ([]string, error) {
	const nixvim = "nixvim"
	err := indices.Register(
		nixvim,
		nuschtos.NewFetcher("https://nix-community.github.io/nixvim/search/meta/"),
		func() indices.Pkg {
			return &nuschtos.Package{}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("register nixvim: %w", err)
	}

	for index, indexHTML := range conf.Experimental.RenderDocsIndexes {
		err := indices.Register(
			index,
			renderdocs.NewFetcher(indexHTML),
			func() indices.Pkg {
				return &renderdocs.Package{
					PageURL: indexHTML,
				}
			},
		)
		if err != nil {
			return nil, fmt.Errorf("register index %q: %w", index, err)
		}
	}

	indexes := slices.Collect(maps.Keys(conf.Experimental.RenderDocsIndexes))
	indexes = append(indexes, nixvim)
	return indexes, nil
}
