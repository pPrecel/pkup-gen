package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

const (
	loggingCategory = "logging config:"
)

func getGenFlags(opts *genActionOpts) []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "repo",
			Usage:   "<org>/<repo> slice",
			Aliases: []string{"r"},
			Action: func(_ *cli.Context, args []string) error {
				repos, err := parseReposMap(opts.Log, args)
				opts.repos = repos
				return err
			},
		},
		&cli.StringFlag{
			Name:        "username",
			Usage:       "GitHub user name",
			Aliases:     []string{"u", "user"},
			Required:    true,
			Destination: &opts.username,
			Action: func(_ *cli.Context, username string) error {
				if username == "" {
					return fmt.Errorf("username '%s' is empty", username)
				}

				return nil
			},
		},
		&cli.IntFlag{
			Name:        "period",
			Usage:       "pkup period to render from 0 to -n",
			Aliases:     []string{"p"},
			Destination: &opts.perdiod,
			Action: func(_ *cli.Context, period int) error {
				if period > 1 {
					return fmt.Errorf("'%d' is not in range from 1 to -n", period)
				}

				return nil
			},
		},
		&cli.StringFlag{
			Name:        "enterprise-url",
			Usage:       "enterprise URL for calling other instances than github.com",
			Aliases:     []string{"e", "enterprise"},
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
			Aliases:     []string{"t", "pat"},
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
			Name:    "output",
			Usage:   "directory path where pkup-gen put generated files",
			Aliases: []string{"o"},
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
		&cli.BoolFlag{
			Name:        "with-closed",
			Usage:       "count closed (not merged) PullRequests",
			Aliases:     []string{"wc", "closed"},
			Destination: &opts.withClosed,
		},
		&cli.BoolFlag{
			Name:     "ci",
			Usage:    "print output using standard log and JSON format",
			Category: loggingCategory,
			Action: func(_ *cli.Context, b bool) error {
				opts.ci = b
				opts.Log = opts.Log.WithFormatter(pterm.LogFormatterJSON)
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
