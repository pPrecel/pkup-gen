package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/pPrecel/PKUP/internal/logo"
	"github.com/pPrecel/PKUP/pkg/config"
	"github.com/pPrecel/PKUP/pkg/period"
	"github.com/pPrecel/PKUP/pkg/report"
	"github.com/pPrecel/PKUP/pkg/send"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

func NewSendCommand(opts *Options) *cli.Command {
	_, until := period.GetCurrentPKUP()
	actionOpts := &sendActionOpts{
		Options:         opts,
		reportTimestamp: *cli.NewTimestamp(until),
	}
	return &cli.Command{
		Name:      "send",
		Usage:     "Send emails with generated reports based on the config",
		UsageText: "pkup send --config .pkupcompose.yaml",
		Aliases:   []string{"s"},
		Before: func(_ *cli.Context) error {
			// print logo before any action
			fmt.Printf("%s\n\n", logo.Build(opts.BuildVersion))

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Value:       ".pkupcompose.yaml",
				Destination: &actionOpts.config,
				Action: func(_ *cli.Context, path string) error {
					path, err := filepath.Abs(path)
					if err != nil {
						return err
					}

					actionOpts.config = path
					return nil
				},
			},
			&cli.TimestampFlag{
				Name:     "timestamp",
				Usage:    "timestamp used to create zip file suffix base on month and year" + report.PeriodFormat,
				Layout:   report.PeriodFormat,
				Timezone: time.Local,
				Action: func(_ *cli.Context, t *time.Time) error {
					actionOpts.reportTimestamp.SetTimestamp(t.Add(time.Hour*24 - time.Second))
					return nil
				},
			},
			&cli.BoolFlag{
				Name:               "v",
				Usage:              "verbose log mode",
				DisableDefaultText: true,
				Category:           loggingCategory,
				Action: func(_ *cli.Context, _ bool) error {
					opts.Log.Level = pterm.LogLevelDebug
					return nil
				},
			},
			&cli.BoolFlag{
				Name:               "vv",
				Usage:              "trace log mode",
				DisableDefaultText: true,
				Category:           loggingCategory,
				Action: func(_ *cli.Context, _ bool) error {
					opts.Log.Level = pterm.LogLevelTrace
					return nil
				},
			},
		},
		Action: func(_ *cli.Context) error {
			return sendCommandAction(actionOpts)
		},
	}
}

func sendCommandAction(opts *sendActionOpts) error {
	opts.Log.Info("sending reports for", opts.Log.Args(
		"config", opts.config,
	))

	config, err := config.Read(opts.config)
	if err != nil {
		return fmt.Errorf("failed to read config from path '%s': %s", opts.config, err.Error())
	}

	zipPrefix := opts.reportTimestamp.Value().Format("200601_")
	return send.New(opts.Log).ForConfig(config, zipPrefix)
}
