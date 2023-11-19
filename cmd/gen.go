package cmd

import (
	"errors"
	"fmt"

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
	logTimeFormat = "02.01.2006 15:04:05"
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
		Flags:   getGenFlags(actionsOpts),
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

	if !opts.ci {
		warnOnNewRelease(client, opts)
	}

	log.Info("generating artifacts for the actual PKUP period", log.Args(
		"since", opts.since.Value().Local().Format(logTimeFormat),
		"until", opts.until.Value().Local().Format(logTimeFormat),
	))

	authors, err := client.GetUserSignatures(opts.username)
	if err != nil {
		return fmt.Errorf("failed to resolve '%s' user: %s", opts.username, err.Error())
	}

	reportResults := []report.Result{}

	for _, org := range opts.orgs {
		log.Info("looking for repos in org", log.Args("org", org))
		opts.repos[org], err = client.ListRepos(org)
		if err != nil {
			return fmt.Errorf("failed to list repos for org '%s': %s", org, err)
		}
	}

	for org, repos := range opts.repos {
		for i := range repos {
			org := org
			repo := repos[i]

			valChan := make(chan *github.CommitList)
			errChan := make(chan error)
			multiView.Add(fmt.Sprintf("%s/%s", org, repo), valChan, errChan)
			go func() {
				defer close(errChan)
				defer close(valChan)

				config := artifacts.Options{
					Org:     org,
					Repo:    repo,
					Authors: authors,
					Dir:     opts.outputDir,
					Since:   *opts.since.Value(),
					Until:   *opts.until.Value(),
				}

				commitList, processErr := artifacts.GenUserArtifactsToDir(client, config)
				if processErr != nil {
					errChan <- processErr
					return
				}

				reportResults = append(reportResults, report.Result{
					Org:        org,
					Repo:       repo,
					CommitList: commitList,
				})
				valChan <- commitList
			}()
		}
	}

	multiView.Run()

	err = report.Render(report.Options{
		OutputDir:    opts.outputDir,
		TemplatePath: opts.templatePath,
		PeriodFrom:   *opts.since.Value(),
		PeriodTill:   *opts.until.Value(),
		Results:      reportResults,
		CustomValues: opts.reportFields,
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
