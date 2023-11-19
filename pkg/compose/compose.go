package compose

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pPrecel/PKUP/internal/view"
	"github.com/pPrecel/PKUP/pkg/artifacts"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pPrecel/PKUP/pkg/report"
	"github.com/pterm/pterm"
)

//go:generate mockery --name=Compose --output=automock --outpkg=automock --case=underscore
type Compose interface {
	ForConfig(*Config, ComposeOpts) error
}

type buildClientFunc func(context.Context, *pterm.Logger, github.ClientOpts) (github.Client, error)

type compose struct {
	ctx         context.Context
	logger      *pterm.Logger
	buildClient buildClientFunc
}

func New(ctx context.Context, logger *pterm.Logger) Compose {
	return &compose{
		ctx:         ctx,
		logger:      logger,
		buildClient: github.NewClient,
	}
}

type ComposeOpts struct {
	Since time.Time
	Until time.Time
	Ci    bool
}

func (c *compose) ForConfig(config *Config, opts ComposeOpts) error {
	remoteClients, err := buildClients(c.ctx, c.logger, config, c.buildClient)
	if err != nil {
		return err
	}

	view := view.NewMultiTaskView(c.logger, opts.Ci)

	for i := range config.Users {
		user := config.Users[i]

		valChan := make(chan *github.CommitList)
		errChan := make(chan error)
		view.Add(user.Username, valChan, errChan)

		go func() {
			commitList, err := c.composeForUser(remoteClients, user, config, opts)
			if err != nil {
				errChan <- err
				return
			}

			valChan <- commitList
		}()
	}

	return view.Run()
}

func (c *compose) composeForUser(remoteClients map[string]github.Client, user User, config *Config, opts ComposeOpts) (*github.CommitList, error) {
	outputDir, err := sanitizeOutputDir(user.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize path '%s': %s", user.OutputDir, err.Error())
	}

	orgRepos, err := c.listOrgRepos(remoteClients, config.Orgs)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories for orgs: %s", err.Error())
	}

	authorsMap, err := buildAuthors(remoteClients, user)
	if err != nil {
		return nil, fmt.Errorf("failed to list user signatures: %s", err.Error())
	}

	repos := append(config.Repos, orgRepos...)
	wg := sync.WaitGroup{}
	var errors error
	commitList := github.CommitList{}
	results := []report.Result{}
	for i := range repos {
		repo := repos[i]
		wg.Add(1)
		go func() {
			orgName, repoName := splitRemoteName(repo.Name)
			commits, genErr := artifacts.GenUserArtifactsToDir(remoteClients[repo.EnterpriseUrl], artifacts.Options{
				Org:     orgName,
				Repo:    repoName,
				Authors: authorsMap[repo.EnterpriseUrl],
				Dir:     outputDir,
				Since:   opts.Since,
				Until:   opts.Until,
			})
			if genErr != nil {
				errors = multierror.Append(errors, fmt.Errorf(
					"failed to generate artifacts for repo '%s': %s", repo.Name, genErr.Error(),
				))
			}

			if commits != nil {
				commitList.Append(commits)
				results = append(results, report.Result{
					Org:        orgName,
					Repo:       repoName,
					CommitList: commits,
				})
			}

			wg.Done()
		}()
	}

	wg.Wait()
	if errors != nil {
		return nil, errors
	}

	templatePath, err := filepath.Abs(config.Template)
	if err != nil {
		return nil, err
	}

	err = report.Render(report.Options{
		OutputDir:    outputDir,
		TemplatePath: templatePath,
		PeriodFrom:   opts.Since,
		PeriodTill:   opts.Until,
		Results:      results,
		CustomValues: user.ReportFields,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render report: %s", err.Error())
	}

	return &commitList, nil
}

func (c *compose) listOrgRepos(remoteClients map[string]github.Client, orgs []Remote) ([]Remote, error) {
	remotes := []Remote{}
	for _, org := range orgs {
		c := remoteClients[org.EnterpriseUrl]

		repos, err := c.ListRepos(org.Name)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			remotes = append(remotes, Remote{
				Name:          fmt.Sprintf("%s/%s", org.Name, repo),
				EnterpriseUrl: org.EnterpriseUrl,
				Token:         org.Token,
			})
		}
	}

	return remotes, nil
}

func sanitizeOutputDir(dir string) (string, error) {
	outputDir, err := filepath.Abs(dir)
	if err != nil {
		return outputDir, err
	}

	return outputDir, os.MkdirAll(outputDir, os.ModePerm)
}

func splitRemoteName(remote string) (string, string) {
	repoOrg := strings.Split(remote, "/")
	return repoOrg[0], repoOrg[1]
}
