package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func NewVersionCommand(opts *Options) *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Shows tool version",
		Action: func(_ *cli.Context) error {
			fmt.Printf("pkup-gen version: %s\n", opts.Version)
			return nil
		},
	}
}
