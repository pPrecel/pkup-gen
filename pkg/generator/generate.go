package generator

import (
	"fmt"
	"time"

	"github.com/pPrecel/PKUP/internal/view"
	"github.com/pPrecel/PKUP/pkg/artifacts"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pPrecel/PKUP/pkg/report"
)

type GeneratorArgs struct {
	Username      string
	Orgs          []string
	Repos         map[string][]string
	Token         string
	EnterpriseURL string
	OutputDir     string
	TemplatePath  string
	ReportFields  map[string]string
	Ci            bool
	Since         *time.Time
	Until         *time.Time

	ProjectMeta
}

// TODO: extract this
type ProjectMeta struct {
	ProjectOwner string
	ProjectRepo  string
	BuildVersion string
}

func (g *generator) ForArgs(args *GeneratorArgs) error {
	client, err := github.NewClient(g.ctx, g.logger, github.ClientOpts{
		Token:         args.Token,
		EnterpriseURL: args.EnterpriseURL,
	})
	if err != nil {
		return fmt.Errorf("create Github client error: %s", err.Error())
	}

	multiView := view.NewMultiTaskView(g.logger, args.Ci)
	log := g.logger.WithWriter(multiView.NewWriter())

	if !args.Ci {
		g.warnOnNewRelease(client, args)
	}

	log.Info("generating artifacts for the PKUP period", log.Args(
		"since", args.Since.Local().Format(logTimeFormat),
		"until", args.Until.Local().Format(logTimeFormat),
	))

	authors, err := client.GetUserSignatures(args.Username)
	if err != nil {
		return fmt.Errorf("failed to resolve '%s' user: %s", args.Username, err.Error())
	}

	reportResults := []report.Result{}

	for _, org := range args.Orgs {
		log.Info("looking for repos in org", log.Args("org", org))
		args.Repos[org], err = client.ListRepos(org)
		if err != nil {
			return fmt.Errorf("failed to list repos for org '%s': %s", org, err)
		}
	}

	for org, repos := range args.Repos {
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
					Dir:     args.OutputDir,
					Since:   *args.Since,
					Until:   *args.Until,
				}

				commitList, processErr := artifacts.GenUserArtifactsToDir(client, config)
				if processErr != nil {
					errChan <- processErr
					return
				}

				// url := opts.enterpriseURL
				// if url == "" {
				// 	url = "https://github.com"
				// }
				reportResults = append(reportResults, report.Result{
					Org:  org,
					Repo: repo,
					// URL:        url,
					CommitList: commitList,
				})
				valChan <- commitList
			}()
		}
	}

	multiView.Run()

	err = report.Render(report.Options{
		OutputDir:    args.OutputDir,
		TemplatePath: args.TemplatePath,
		PeriodFrom:   *args.Since,
		PeriodTill:   *args.Until,
		Results:      reportResults,
		CustomValues: args.ReportFields,
	})
	if err != nil {
		return err
	}

	g.logger.Info("all files saved to dir", g.logger.Args("dir", args.OutputDir))
	return nil
}

// TODO: extract this func from the package
func (g *generator) warnOnNewRelease(client github.Client, args *GeneratorArgs) {
	latestVersion, err := client.GetLatestReleaseOrZero(args.ProjectOwner, args.ProjectRepo)
	if err != nil {
		g.logger.Trace("failed to check the latest available release", g.logger.Args(
			"error", err.Error(),
		))
		return
	}
	if args.BuildVersion != "local" && args.BuildVersion != latestVersion {
		g.logger.Warn("new pkup-gen release detected - upgrade to get all features and fixes", g.logger.Args(
			"buildVersion", args.BuildVersion,
			"latestVersion", latestVersion,
		))
	}
}
