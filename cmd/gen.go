package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	gh "github.com/google/go-github/v53/github"
	"github.com/pPrecel/PKUP/internal/file"
	"github.com/pPrecel/PKUP/internal/view"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pPrecel/PKUP/pkg/period"
	"github.com/pterm/pterm"
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
			&cli.StringFlag{
				Name:        "enterprise-url",
				Usage:       "enterprise URL for calling other instances than github.com",
				Aliases:     []string{"e", "enterprise"},
				Destination: &actionsOpts.enterpriseURL,
				Action: func(ctx *cli.Context, url string) error {
					if url == "" {
						return fmt.Errorf("'%s' enterprise url is empty", url)
					}

					return nil
				},
			},
			&cli.BoolFlag{
				Name:               "v",
				Usage:              "verbose log mode",
				DisableDefaultText: true,
				Action: func(_ *cli.Context, _ bool) error {
					actionsOpts.Log.Level = pterm.LogLevelDebug
					return nil
				},
			},
			&cli.BoolFlag{
				Name:               "vv",
				Usage:              "trace log mode",
				DisableDefaultText: true,
				Action: func(_ *cli.Context, _ bool) error {
					actionsOpts.Log.Level = pterm.LogLevelTrace
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

	multiView := view.NewMultiTaskView()
	log := opts.Log.WithWriter(multiView.NewWriter())
	client, err := github.NewClient(
		ctx.Context, log, opts.token, opts.enterpriseURL,
	)
	if err != nil {
		return fmt.Errorf("create Github client error: %s", err.Error())
	}

	mergedAfter, mergedBefore := period.GetLastPKUP(opts.perdiod)
	log.Info("generating artifacts for the actual PKUP period", log.Args(
		"after", mergedAfter.Local().Format(logTimeFormat),
		"before", mergedBefore.Local().Format(logTimeFormat),
	))

	for org, repos := range opts.repos {
		for i := range repos {
			org := org
			repo := repos[i]

			valChan := make(chan []string)
			errChan := make(chan error)
			multiView.Add(fmt.Sprintf("%s/%s", org, repo), valChan, errChan)
			go func() {
				defer close(errChan)
				defer close(valChan)

				prs, err := listUserPRsToFile(client, &listToFileOpts{
					org:          org,
					repo:         repo,
					username:     opts.username,
					dir:          opts.dir,
					mergedAfter:  mergedAfter,
					mergedBefore: mergedBefore,
				})
				if err != nil {
					errChan <- err
					return
				}

				valChan <- prs
			}()
		}
	}

	multiView.Run()

	log.Info("all patch files saved to dir", log.Args("dir", opts.dir))
	return nil
}

type listToFileOpts struct {
	org          string
	repo         string
	username     string
	dir          string
	mergedAfter  time.Time
	mergedBefore time.Time
}

func listUserPRsToFile(client github.Client, opts *listToFileOpts) ([]string, error) {
	prs, err := client.ListUserPRsForRepo(github.Options{
		Org:          opts.org,
		Repo:         opts.repo,
		Username:     opts.username,
		MergedAfter:  opts.mergedAfter,
		MergedBefore: opts.mergedBefore,
	})
	if err != nil {
		return nil, fmt.Errorf("list users PRs in repo '%s/%s' error: %s", opts.org, opts.repo, err.Error())
	}

	diff, err := client.GetFileDiffForPRs(prs, opts.org, opts.repo)
	if err != nil {
		return nil, fmt.Errorf("get diff for repo '%s/%s' error: %s", opts.org, opts.repo, err.Error())
	}

	if diff != "" {
		filename := fmt.Sprintf("%s_%s.patch", opts.org, opts.repo)
		err = file.Create(opts.dir, filename, diff)
		if err != nil {
			return nil, fmt.Errorf("save file '%s' error: %s", filename, err.Error())
		}
	}

	return prsToStringList(prs), nil
}

func prsToStringList(prs []*gh.PullRequest) []string {
	list := []string{}
	for i := range prs {
		pr := *prs[i]
		list = append(list, *pr.Title)
	}

	return list
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
