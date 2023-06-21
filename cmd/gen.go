package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	gh "github.com/google/go-github/v53/github"
	"github.com/pPrecel/PKUP/internal/file"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pPrecel/PKUP/pkg/period"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	logTimeFormat = time.DateTime
)

func NewGenCommand(opts *Options) *cli.Command {
	actionsOpts := &genActionOpts{
		Options: opts,
	}

	return &cli.Command{
		Name:  "gen",
		Usage: "Generates .patch files with all users merged content in the last PKUP period",
		UsageText: "pkup gen --token <personal-access-token> \\\n" +
			"\t\t--username <username> \\\n" +
			"\t\t--repo <org1>/<repo1> \\\n" +
			"\t\t--repo <org2>/<repo2>",
		Aliases: []string{"g", "generate", "get"},
		Flags: []cli.Flag{
			cli.HelpFlag,
			&cli.StringSliceFlag{
				Name:    "repo",
				Usage:   "<org>/<repo> slice",
				Aliases: []string{"r"},
				Action: func(ctx *cli.Context, args []string) error {
					repos, err := parseReposMap(opts.Log, args)
					actionsOpts.repos = repos
					return err
				},
			},
			&cli.StringFlag{
				Name:        "username",
				Usage:       "GitHub user name",
				Aliases:     []string{"u", "user"},
				Required:    true,
				Destination: &actionsOpts.username,
				Action: func(ctx *cli.Context, username string) error {
					if username == "" {
						return fmt.Errorf("username '%s' is empty", username)
					}

					return nil
				},
			},
			&cli.StringFlag{
				Name:        "token",
				Aliases:     []string{"t", "pta", "personalaccesstoken"},
				Usage:       "personal access token",
				Required:    true,
				Destination: &actionsOpts.token,
				Action: func(ctx *cli.Context, token string) error {
					if token == "" {
						return errors.New("The 'token' flag can't be empty")
					}

					return nil
				},
			},
			&cli.StringFlag{
				Name:        "dir",
				Usage:       "destination of .patch files",
				Aliases:     []string{"d"},
				Destination: &actionsOpts.dir,
				Action: func(_ *cli.Context, dir string) error {
					info, err := os.Stat(dir)
					if err != nil {
						return err
					}

					if !info.IsDir() {
						return fmt.Errorf("'%s' is not dir", dir)
					}

					actionsOpts.dir = dir
					return nil
				},
			},
			&cli.IntFlag{
				Name:        "period",
				Usage:       "pkup period to render from 0 to -n",
				Aliases:     []string{"p"},
				Destination: &actionsOpts.perdiod,
				Action: func(ctx *cli.Context, period int) error {
					if period > 0 {
						return fmt.Errorf("'%d' is not in range from 0 to -n", period)
					}

					return nil
				},
			},
			&cli.BoolFlag{
				Name:               "verbose",
				Aliases:            []string{"v"},
				Usage:              "verbose mode",
				DisableDefaultText: true,
				Action: func(_ *cli.Context, _ bool) error {
					actionsOpts.Log.Level = logrus.DebugLevel
					return nil
				},
			},
		},
		Action: func(ctx *cli.Context) error {
			return genCommandAction(ctx, actionsOpts)
		},
	}
}

func genCommandAction(ctx *cli.Context, opts *genActionOpts) error {
	if opts.dir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get pwd error: %s", err.Error())
		}
		opts.dir = pwd
	}

	client := github.NewClient(
		ctx.Context, opts.Log, opts.token,
	)

	mergedAfter, mergedBefore := period.GetLastPKUP(opts.perdiod)
	opts.Log.Infof("looking for changes beteen %s and %s",
		mergedAfter.Local().Format(logTimeFormat),
		mergedBefore.Local().Format(logTimeFormat))

	for org, repos := range opts.repos {
		for i := range repos {
			repo := repos[i]

			opts.Log.Infof("processing '%s/%s' repo", org, repo)
			prs, err := client.ListUserPRsForRepo(github.Options{
				Org:          org,
				Repo:         repo,
				Username:     opts.username,
				MergedAfter:  mergedAfter,
				MergedBefore: mergedBefore,
			})
			if err != nil {
				return fmt.Errorf("list users PRs in repo '%s/%s' error: %s",
					org,
					repo,
					err.Error(),
				)
			}

			printPRs(opts.Log, opts.username, prs)

			diff, err := client.GetFileDiffForPRs(prs, org, repo)
			if err != nil {
				return fmt.Errorf("get diff for repo '%s/%s' error: %s",
					org,
					repo,
					err.Error(),
				)
			}

			if diff == "" {
				opts.Log.Warnf("skipping '%s/%s' no user activity detected", org, repo)
				continue
			}

			filename := fmt.Sprintf("%s_%s.patch", org, repo)
			err = file.Create(opts.dir, filename, diff)
			if err != nil {
				return fmt.Errorf("save file '%s' error: %s",
					filename,
					err.Error(),
				)
			}

			opts.Log.Infof("patch saved to file '%s/%s'", opts.dir, filename)
		}
	}

	return nil
}

func printPRs(log *logrus.Logger, username string, prs []*gh.PullRequest) {
	for i := range prs {
		pr := *prs[i]
		log.Infof("\tuser '%s' is an author of '%s'", username, pr.GetTitle())
	}
}

func parseReposMap(log *logrus.Logger, args []string) (map[string][]string, error) {
	repos := map[string][]string{}
	for i := range args {
		arg := args[i]

		log.Debugf("parsing '%s' repo flag", arg)
		argSlice := strings.Split(arg, "/")
		if len(argSlice) != 2 {
			return nil, fmt.Errorf("repo '%s' must be in format <org>/<repo>", arg)
		}

		repos[argSlice[0]] = append(repos[argSlice[0]], argSlice[1])
	}

	return repos, nil
}
