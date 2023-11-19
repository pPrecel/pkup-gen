package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/pPrecel/PKUP/internal/logo"
	"github.com/pPrecel/PKUP/pkg/compose"
	"github.com/pPrecel/PKUP/pkg/period"
	"github.com/pPrecel/PKUP/pkg/report"
	"github.com/urfave/cli/v2"
)

func NewComposeCommand(opts *Options) *cli.Command {
	since, until := period.GetCurrentPKUP()
	actionsOpts := composeActionOpts{
		Options: opts,
		since:   *cli.NewTimestamp(since),
		until:   *cli.NewTimestamp(until),
	}

	return &cli.Command{
		Name:      "compose",
		Usage:     "Generates .diff and report files for many users and based on the .yaml config file",
		UsageText: "pkup gen --config .pkupcompose.yaml",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Value:       ".pkupcompose.yaml",
				Destination: &actionsOpts.config,
				Action: func(_ *cli.Context, path string) error {
					path, err := filepath.Abs(path)
					if err != nil {
						return err
					}

					actionsOpts.config = path
					return nil
				},
			},
			&cli.TimestampFlag{
				Name:     "since",
				Usage:    "timestamp used to get commits and render report - foramt " + report.PeriodFormat,
				Layout:   report.PeriodFormat,
				Timezone: time.Local,
				Action: func(_ *cli.Context, time *time.Time) error {
					actionsOpts.since.SetTimestamp(*time)
					return nil
				},
			},
			&cli.TimestampFlag{
				Name:     "until",
				Usage:    "timestamp used to get commits and render report - foramt " + report.PeriodFormat,
				Layout:   report.PeriodFormat,
				Timezone: time.Local,
				Action: func(_ *cli.Context, t *time.Time) error {
					actionsOpts.until.SetTimestamp(t.Add(time.Hour*24 - time.Second))
					return nil
				},
			},
			&cli.BoolFlag{
				Name:        "ci",
				Usage:       "print output using standard log",
				Category:    loggingCategory,
				Destination: &actionsOpts.ci,
			},
		},
		Before: func(_ *cli.Context) error {
			// print logo before any action
			fmt.Printf("%s\n\n", logo.Build(opts.BuildVersion))

			return nil
		},
		Action: func(ctx *cli.Context) error {
			return composeCommandAction(ctx, &actionsOpts)
		},
	}
}

func composeCommandAction(ctx *cli.Context, opts *composeActionOpts) error {
	opts.Log.Info("generating artifacts for the PKUP period", opts.Log.Args(
		"config", opts.config,
		"since", opts.since.Value().Local().Format(logTimeFormat),
		"until", opts.until.Value().Local().Format(logTimeFormat),
	))

	cfg, err := compose.ReadConfig(opts.config)
	if err != nil {
		return fmt.Errorf("failed to read config from path '%s': %s", opts.config, err.Error())
	}

	return compose.New(ctx.Context, opts.Log).ForConfig(cfg, compose.ComposeOpts{
		Since: *opts.since.Value(),
		Until: *opts.until.Value(),
		Ci:    opts.ci,
	})
}
