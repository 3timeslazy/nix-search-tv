package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/3timeslazy/nix-search-tv/config"
	"github.com/3timeslazy/nix-search-tv/indexes/indices"

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
	JsonFlag     = "json"
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

	if cmd.IsSet(CacheDirFlag) {
		conf.CacheDir = cmd.String(CacheDirFlag)
	}

	if err = validateIndexes(conf, conf.Indexes); err != nil {
		return config.Config{}, err
	}

	if err := os.MkdirAll(conf.CacheDir, 0755); err != nil {
		return conf, fmt.Errorf("cannot create cache directory: %w", err)
	}

	return conf, nil
}

func validateIndexes(conf config.Config, indexNames []string) error {
	for index := range conf.Experimental.RenderDocsIndexes {
		if indices.BuiltinIndexes[index] {
			return fmt.Errorf("experimental %[1]q conflicts with builtin %[1]q", index)
		}
	}

	for _, index := range indexNames {
		if indices.BuiltinIndexes[index] {
			continue
		}

		_, parseHTML := conf.Experimental.RenderDocsIndexes[index]
		if !parseHTML {
			valid := strings.Join(indexNames, "\n")
			return fmt.Errorf("unknown index %q. Valid values are:\n %s", index, valid)
		}
	}

	return nil
}
