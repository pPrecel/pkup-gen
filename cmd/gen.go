package cmd

import (
	"fmt"
	"time"

	"github.com/pPrecel/PKUP/internal/view"
	"github.com/pPrecel/PKUP/pkg/artifacts"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pPrecel/PKUP/pkg/period"
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
		Flags:   getGenFlags(actionsOpts),
		Action: func(ctx *cli.Context) error {
			if err := actionsOpts.setDefaults(); err != nil {
				return err
			}
			return genCommandAction(ctx, actionsOpts)
		},
	}
}

func genCommandAction(ctx *cli.Context, opts *genActionOpts) error {
	multiView := view.NewMultiTaskView(opts.Log, opts.ci)
	log := opts.Log.WithWriter(multiView.NewWriter())

	client, err := github.NewClient(ctx.Context, log, opts.token, opts.enterpriseURL)
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

				config := artifacts.GenerateOpts{
					Org:          org,
					Repo:         repo,
					Username:     opts.username,
					Dir:          opts.dir,
					WithClosed:   opts.withClosed,
					MergedAfter:  mergedAfter,
					MergedBefore: mergedBefore,
				}

				log.Debug("starting process for repo with config", log.Args(
					"org", config.Org,
					"repo", config.Repo,
					"username", config.Username,
					"dir", config.Dir,
					"withClosed", config.WithClosed,
					"mergedAfter", config.MergedAfter.String(),
					"mergedBefore", config.MergedBefore.String(),
				))

				prs, err := artifacts.GenUserArtifactsToFile(client, &config)

				log.Debug("ending process for repo", log.Args(
					"prs", prs,
					"error", err,
				))
				if err != nil {
					errChan <- err
					return
				}

				valChan <- prs
			}()
		}
	}

	multiView.Run()

	opts.Log.Info("all patch files saved to dir", opts.Log.Args("dir", opts.dir))
	return nil
}
