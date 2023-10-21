package cmd

import (
	"errors"
	"fmt"
	"time"

	gh "github.com/google/go-github/v53/github"
	"github.com/pPrecel/PKUP/internal/logo"
	"github.com/pPrecel/PKUP/internal/token"
	"github.com/pPrecel/PKUP/internal/view"
	"github.com/pPrecel/PKUP/pkg/artifacts"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pPrecel/PKUP/pkg/period"
	"github.com/pPrecel/PKUP/pkg/report"
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
		Usage: "Generates .diff files with all users merged content in the last PKUP period",
		UsageText: "pkup gen --token <personal-access-token> \\\n" +
			"\t\t--username <username> \\\n" +
			"\t\t--repo <org1>/<repo1> \\\n" +
			"\t\t--repo <org2>/<repo2>",
		Aliases: []string{"g", "generate", "get"},
		Flags:   getGenFlags(actionsOpts),
		Before: func(_ *cli.Context) error {
			// print logo before any action
			if !actionsOpts.ci {
				fmt.Printf("%s\n\n", logo.Build(opts.BuildVersion))
			}

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

	client, err := github.NewClient(ctx.Context, opts.Log, github.ClientOpts{
		Token:         opts.token,
		EnterpriseURL: opts.enterpriseURL,
		AppClientID:   opts.PkupClientID,
	})
	if err != nil {
		return fmt.Errorf("create Github client error: %s", err.Error())
	}

	multiView := view.NewMultiTaskView(opts.Log, opts.ci)
	log := opts.Log.WithWriter(multiView.NewWriter())

	warnOnNewRelease(client, opts)

	mergedAfter, mergedBefore := period.GetLastPKUP(opts.perdiod)
	log.Info("generating artifacts for the actual PKUP period", log.Args(
		"after", mergedAfter.Local().Format(logTimeFormat),
		"before", mergedBefore.Local().Format(logTimeFormat),
	))

	reportResults := []report.Result{}
	for org, repos := range opts.repos {
		for i := range repos {
			org := org
			repo := repos[i]

			valChan := make(chan []*gh.PullRequest)
			errChan := make(chan error)
			multiView.Add(fmt.Sprintf("%s/%s", org, repo), valChan, errChan)
			go func() {
				defer close(errChan)
				defer close(valChan)

				config := artifacts.Options{
					Org:          org,
					Repo:         repo,
					Username:     opts.username,
					Dir:          opts.outputDir,
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

				prs, processErr := artifacts.GenUserArtifactsToDir(client, config)

				log.Debug("ending process for repo", log.Args(
					"org", config.Org,
					"repo", config.Repo,
					"prs", prs,
					"error", processErr,
				))
				if processErr != nil {
					errChan <- processErr
					return
				}

				valChan <- prs
				reportResults = append(reportResults, report.Result{
					Org:          org,
					Repo:         repo,
					PullRequests: prs,
				})
			}()
		}
	}

	multiView.Run()

	err = report.Render(report.Options{
		OutputDir:    opts.outputDir,
		TemplatePath: opts.templatePath,
		PeriodFrom:   mergedAfter,
		PeriodTill:   mergedBefore,
		Results:      reportResults,
	})
	if err != nil {
		return err
	}

	opts.Log.Info("all files saved to dir", opts.Log.Args("dir", opts.outputDir))
	return nil
}

func warnOnNewRelease(client github.Client, opts *genActionOpts) {
	latestVersion, err := client.GetLatestReleaseOrZero(opts.ProjectOwner, opts.ProjectRepo)
	if err != nil {
		opts.Log.Trace("failed to check the latest available release", opts.Log.Args(
			"error", err.Error(),
		))
		return
	}
	if opts.BuildVersion != "local" && opts.BuildVersion != latestVersion {
		opts.Log.Warn("new pkup-gen release detected - upgrade to get all features and fixes", opts.Log.Args(
			"buildVersion", opts.BuildVersion,
			"latestVersion", latestVersion,
		))
	}
}
