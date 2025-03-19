package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/3timeslazy/nix-search-tv/cmd"

	"github.com/urfave/cli/v3"
)

var root = &cli.Command{
	Name:      "nix-search-tv",
	UsageText: `nix-search-tv [options] [command]`,
	Usage:     "Nix-related television channel",
	Commands: []*cli.Command{
		cmd.Print,
		cmd.Preview,
		cmd.Source,
		cmd.Homepage,
		cmd.Grep,
	},
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := root.Run(ctx, os.Args); err != nil {
		code := 1

		var codeError cli.ExitCoder
		if errors.As(err, &codeError) {
			code = codeError.ExitCode()
		}

		fmt.Println(err)

		os.Exit(code)
	}
}
