package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pPrecel/PKUP/internal/logo"
	"github.com/pPrecel/PKUP/internal/token"
	"github.com/pPrecel/PKUP/pkg/generator"
	"github.com/pPrecel/PKUP/pkg/period"
	"github.com/pPrecel/PKUP/pkg/report"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

const (
	logTimeFormat   = "02.01.2006 15:04:05"
	loggingCategory = "logging config:"
)

func NewGenCommand(opts *Options) *cli.Command {
	since, until := period.GetCurrentPKUP()
	actionsOpts := &genActionOpts{
		Options: opts,
		since:   *cli.NewTimestamp(since),
		until:   *cli.NewTimestamp(until),
		repos:   map[string][]string{},
	}

	return &cli.Command{
		Name:  "gen",
		Usage: "Generates .diff and report files with all users merged content in the last PKUP period",
		UsageText: "pkup gen \\\n" +
			"\t\t--username <username> \\\n" +
			"\t\t--repo <org1>/<repo1> \\\n" +
			"\t\t--repo <org2>/<repo2>",
		Aliases: []string{"g", "generate", "get"},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "repo",
				Usage: "<org>/<repo> slice - use this flag to look for user activity in specified repos",
				Action: func(_ *cli.Context, args []string) error {
					repos, err := parseReposMap(opts.Log, args)
					actionsOpts.repos = repos
					return err
				},
			},
			&cli.StringSliceFlag{
				Name:  "org",
				Usage: "<org> slice - use this flag to look for user activity in all organization repos",
				Action: func(_ *cli.Context, s []string) error {
					actionsOpts.orgs = s
					return nil
				},
			},
			&cli.StringFlag{
				Name:        "username",
				Usage:       "GitHub user name",
				Required:    true,
				Destination: &actionsOpts.username,
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
			&cli.StringFlag{
				Name:        "enterprise-url",
				Usage:       "enterprise URL for calling other instances than github.com",
				Destination: &actionsOpts.enterpriseURL,
				Action: func(_ *cli.Context, url string) error {
					if url == "" {
						return fmt.Errorf("'%s' enterprise url is empty", url)
					}

					return nil
				},
			},
			&cli.StringFlag{
				Name:        "token",
				Usage:       "personal access token",
				Destination: &actionsOpts.token,
				Action: func(_ *cli.Context, token string) error {
					if token == "" {
						return errors.New("'token' flag can't be empty")
					}

					return nil
				},
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "directory path where pkup-gen put generated files",
				Action: func(_ *cli.Context, dir string) error {
					dir, err := filepath.Abs(filepath.Clean(dir))
					if err != nil {
						return err
					}

					info, err := os.Stat(dir)
					if err != nil {
						return err
					}

					if !info.IsDir() {
						return fmt.Errorf("'%s' is not a dir", dir)
					}

					actionsOpts.outputDir = dir
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "template",
				Usage:   "full path to the docx template - if not set program generates .txt data file",
				Aliases: []string{"tmpl"},
				Action: func(_ *cli.Context, path string) error {
					path, err := filepath.Abs(filepath.Clean(path))
					if err != nil {
						return err
					}

					if path == "" {
						return fmt.Errorf("'%s' template path is empty", path)
					}

					actionsOpts.templatePath = path
					return nil
				},
			},
			&cli.StringSliceFlag{
				Name:  "report-field",
				Usage: "custom field that will be replace in the output report - in format FIELD=VALUE",
				Action: func(_ *cli.Context, fields []string) error {
					reportFields, err := parseReportFields(fields)
					actionsOpts.reportFields = reportFields
					return err
				},
			},
			&cli.BoolFlag{
				Name:     "ci",
				Usage:    "print output using standard log",
				Category: loggingCategory,
				Action: func(_ *cli.Context, b bool) error {
					actionsOpts.ci = b
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
		Before: func(_ *cli.Context) error {
			// print logo before any action
			fmt.Printf("%s\n\n", logo.Build(opts.BuildVersion))

			// default
			if err := actionsOpts.setDefaults(); err != nil {
				return err
			}

			// validate
			if actionsOpts.enterpriseURL != "" && actionsOpts.token == "" {
				return errors.New("specify token when using enterprise url")
			}

			return nil
		},
		Action: func(ctx *cli.Context) error {

			return genCommandAction(ctx, actionsOpts)
		},
	}
}

func genCommandAction(ctx *cli.Context, opts *genActionOpts) error {
	if opts.token == "" {
		var err error
		opts.token, err = token.Get(opts.Log, opts.PkupClientID)
		if err != nil {
			return fmt.Errorf("failed to provide token: %s", err.Error())
		}
	}

	return generator.New(ctx.Context, opts.Log).ForArgs(&generator.GeneratorArgs{
		Username:      opts.username,
		Orgs:          opts.orgs,
		Repos:         opts.repos,
		Token:         opts.token,
		EnterpriseURL: opts.enterpriseURL,
		Since:         opts.since.Value(),
		Until:         opts.until.Value(),
		OutputDir:     opts.outputDir,
		TemplatePath:  opts.templatePath,
		ReportFields:  opts.reportFields,
		Ci:            opts.ci,
		ProjectMeta: generator.ProjectMeta{
			ProjectOwner: opts.ProjectOwner,
			ProjectRepo:  opts.ProjectRepo,
			BuildVersion: opts.BuildVersion,
		},
	})
}
