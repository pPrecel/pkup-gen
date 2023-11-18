package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pPrecel/PKUP/pkg/report"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

const (
	loggingCategory = "logging config:"
)

func getGenFlags(opts *genActionOpts) []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "repo",
			Usage: "<org>/<repo> slice - use this flag to look for user activity in specified repos",
			Action: func(_ *cli.Context, args []string) error {
				repos, err := parseReposMap(opts.Log, args)
				opts.repos = repos
				return err
			},
		},
		&cli.StringSliceFlag{
			Name:  "org",
			Usage: "<org> slice - use this flag to look for user activity in all organization repos",
			Action: func(ctx *cli.Context, s []string) error {
				opts.orgs = s
				return nil
			},
		},
		&cli.StringFlag{
			Name:        "username",
			Usage:       "GitHub user name",
			Required:    true,
			Destination: &opts.username,
		},
		&cli.TimestampFlag{
			Name:     "since",
			Usage:    "timestamp used to get commits and render report - foramt " + report.PeriodFormat,
			Layout:   report.PeriodFormat,
			Timezone: time.Local,
			Action: func(_ *cli.Context, time *time.Time) error {
				opts.since.SetTimestamp(*time)
				return nil
			},
		},
		&cli.TimestampFlag{
			Name:     "until",
			Usage:    "timestamp used to get commits and render report - foramt " + report.PeriodFormat,
			Layout:   report.PeriodFormat,
			Timezone: time.Local,
			Action: func(_ *cli.Context, t *time.Time) error {
				opts.until.SetTimestamp(t.Add(time.Hour*24 - time.Second))
				return nil
			},
		},
		&cli.StringFlag{
			Name:        "enterprise-url",
			Usage:       "enterprise URL for calling other instances than github.com",
			Destination: &opts.enterpriseURL,
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
			Destination: &opts.token,
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

				opts.outputDir = dir
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

				opts.templatePath = path
				return nil
			},
		},
		&cli.StringSliceFlag{
			Name:  "report-field",
			Usage: "custom field that will be replace in the output report - in format FIELD=VALUE",
			Action: func(_ *cli.Context, fields []string) error {
				reportFields, err := parseReportFields(fields)
				opts.reportFields = reportFields
				return err
			},
		},
		&cli.BoolFlag{
			Name:     "ci",
			Usage:    "print output using standard log",
			Category: loggingCategory,
			Action: func(_ *cli.Context, b bool) error {
				opts.ci = b
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
	}
}

func parseReportFields(args []string) (map[string]string, error) {
	reportFields := map[string]string{}
	for _, field := range args {
		vals := strings.Split(field, "=")
		if len(vals) != 2 {
			return nil, fmt.Errorf("failed to parse '%s' report field", field)
		}

		reportFields[vals[0]] = vals[1]
	}

	return reportFields, nil
}

func parseReposMap(log *pterm.Logger, args []string) (map[string][]string, error) {
	repos := map[string][]string{}
	for i := range args {
		arg := args[i]

		log.Debug("parsing flag", log.Args("argument", arg))
		argSlice := strings.Split(arg, "/")
		if len(argSlice) != 2 {
			return nil, fmt.Errorf("repo '%s' must be in format <org>/<repo>", arg)
		}

		repos[argSlice[0]] = append(repos[argSlice[0]], argSlice[1])
	}

	return repos, nil
}
