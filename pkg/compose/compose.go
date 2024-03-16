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

const (
	logTimeFormat = "02.01.2006 15:04:05"
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
	view := view.NewMultiTaskView(c.logger, opts.Ci)
	viewLogger := c.logger.WithWriter(view.NewWriter())

	remoteClients, err := buildClients(c.ctx, c.logger, config, c.buildClient)
	if err != nil {
		return err
	}

	var repoCommits []*repoCommits
	var listErr error
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		viewLogger.Debug("listing all commits")
		repoCommits, listErr = c.listAllCommits(remoteClients, config, &opts)
		viewLogger.Debug("found commits for repos", viewLogger.Args("repos count", len(repoCommits)))
		wg.Done()
	}()

	for i := range config.Users {
		user := config.Users[i]

		valChan := make(chan *github.CommitList)
		errChan := make(chan error)
		view.Add(user.Username, valChan, errChan)

		go func() {
			wg.Wait()
			if listErr != nil {
				errChan <- listErr
			}

			viewLogger.Debug("compose for user", viewLogger.Args("user", user.Username))
			commitList, err := c.composeForUser(remoteClients, repoCommits, &user, config, &opts)
			if err != nil {
				errChan <- err
				return
			}

			valChan <- commitList
		}()
	}

	return view.Run()
}

func (c *compose) composeForUser(remoteClients map[string]github.Client, repoCommits []*repoCommits, user *User, config *Config, opts *ComposeOpts) (*github.CommitList, error) {
	outputDir, err := sanitizeOutputDir(user.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize path '%s': %s", user.OutputDir, err.Error())
	}

	urlAuthors, err := buildUrlAuthors(remoteClients, user)
	if err != nil {
		return nil, fmt.Errorf("failed to list user signatures: %s", err.Error())
	}

	wg := sync.WaitGroup{}
	var errors error
	commitList := github.CommitList{}
	results := []report.Result{}
	for i := range repoCommits {
		repo := repoCommits[i]
		wg.Add(1)
		go func() {
			userCommits := github.CommitList{
				Commits: github.GetUserCommits(repo.commits.Commits, urlAuthors[repo.enterpriseUrl]),
			}

			saveErr := artifacts.SaveDiffToFiles(remoteClients[repo.enterpriseUrl], &userCommits, artifacts.Options{
				Org:     repo.org,
				Repo:    repo.repo,
				Authors: urlAuthors[repo.enterpriseUrl],
				Dir:     outputDir,
				Since:   opts.Since,
				Until:   opts.Until,
			})
			if saveErr != nil {
				errors = multierror.Append(errors, fmt.Errorf(
					"failed to generate artifacts for repo '%s': %s", repo.repo, saveErr.Error(),
				))
			} else {
				// url := repo.enterpriseUrl
				// if url == "" {
				// 	url = "https://github.com"
				// }
				commitList.Append(&userCommits)
				results = append(results, report.Result{
					Org:  repo.org,
					Repo: repo.repo,
					// URL:        url,
					CommitList: &userCommits,
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

type repoCommits struct {
	org           string
	repo          string
	enterpriseUrl string
	commits       *github.CommitList
}

func (c *compose) listAllCommits(remoteClients map[string]github.Client, config *Config, opts *ComposeOpts) ([]*repoCommits, error) {
	repos, err := c.listOrgRepos(remoteClients, config)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories for orgs: %s", err.Error())
	}

	wg := sync.WaitGroup{}
	allRepoCommits := make([]*repoCommits, len(repos))
	for i, r := range repos {
		iter := i
		repo := r

		wg.Add(1)
		go func() {
			defer wg.Done()

			orgName, repoName := splitRemoteName(repo.Name)
			client := remoteClients[repo.EnterpriseUrl]

			c.logger.Debug("list commits for repo", c.logger.Args("org", orgName, "repo", repoName))
			commitList, listErr := client.ListRepoCommits(github.ListRepoCommitsOpts{
				Org:        orgName,
				Repo:       repoName,
				Since:      opts.Since,
				Until:      opts.Until,
				Branches:   repo.Branches,
				UniqueOnly: repo.UniqueOnly,
			})
			if listErr != nil {
				c.logger.Warn("failed to list commits", c.logger.Args("org", orgName, "repo", repoName, "error", listErr.Error()))
				multierror.Append(err, listErr)
				return
			}

			c.logger.Debug("found commits", c.logger.Args("org", orgName, "repo", repoName, "count", len(commitList.Commits)))
			allRepoCommits[iter] = &repoCommits{
				org:           orgName,
				repo:          repoName,
				enterpriseUrl: repo.EnterpriseUrl,
				commits:       commitList,
			}
		}()
	}

	wg.Wait()
	return allRepoCommits, err
}

func (c *compose) listOrgRepos(remoteClients map[string]github.Client, config *Config) ([]Remote, error) {
	remotes := []Remote{}

	// resolve orgs
	for _, org := range config.Orgs {
		c := remoteClients[org.EnterpriseUrl]

		repos, err := c.ListRepos(org.Name)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			name := fmt.Sprintf("%s/%s", org.Name, repo)

			if containsOrgRepo(config.Repos, name) {
				// skip if repo already is in config.Repos
				continue
			}

			remotes = append(remotes, Remote{
				Name:          name,
				EnterpriseUrl: org.EnterpriseUrl,
				Token:         org.Token,
				Branches:      org.Branches,
				AllBranches:   org.AllBranches,
				UniqueOnly:    org.UniqueOnly,
			})
		}
	}

	remotes = append(config.Repos, remotes...)

	// check if remote has AllBranches set
	for i, remote := range remotes {
		c := remoteClients[remote.EnterpriseUrl]
		repoOrg := strings.Split(remote.Name, "/")

		if remote.AllBranches {
			branchList, listError := c.ListRepoBranches(repoOrg[0], repoOrg[1])
			if listError != nil {
				return nil, listError
			}

			remotes[i].Branches = branchList.Branches
		}
	}

	return remotes, nil
}

func containsOrgRepo(remotes []Remote, orgRepo string) bool {
	for _, remote := range remotes {
		if remote.Name == orgRepo {
			return true
		}
	}

	return false
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
