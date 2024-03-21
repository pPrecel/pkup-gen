package generator

import (
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
)

const (
	logTimeFormat = "02.01.2006 15:04:05"
)

type ComposeOpts struct {
	Since time.Time
	Until time.Time
	Ci    bool
}

func (c *generator) ForConfig(config *Config, opts ComposeOpts) error {
	view := view.NewMultiTaskView(c.logger, opts.Ci)
	viewLogger := c.logger.WithWriter(view.NewWriter())

	remoteClients, err := buildClients(c.ctx, c.logger, config, c.buildClient)
	if err != nil {
		return err
	}

	c.repoCommitsLister = NewLazyRepoCommitsLister(c.logger, remoteClients)

	for i := range config.Users {
		user := config.Users[i]

		valChan := make(chan *github.CommitList)
		errChan := make(chan error)
		view.Add(getUsernames(user), valChan, errChan)

		go func() {
			viewLogger.Debug("compose for user", viewLogger.Args("user", user.Username))
			commitList, err := c.composeForUser(remoteClients, &user, config, &opts)
			if err != nil {
				errChan <- err
				return
			}

			valChan <- commitList
		}()
	}

	return view.Run()
}

func (c *generator) composeForUser(remoteClients *remoteClients, user *User, config *Config, opts *ComposeOpts) (*github.CommitList, error) {
	outputDir, err := sanitizeOutputDir(user.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize path '%s': %s", user.OutputDir, err.Error())
	}

	urlAuthors, err := buildUrlAuthors(remoteClients, user)
	if err != nil {
		return nil, fmt.Errorf("failed to list user signatures: %s", err.Error())
	}

	repoCommits, err := c.repoCommitsLister.List(config, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %s", err.Error())
	}

	wg := sync.WaitGroup{}
	var errors error
	commitList := github.CommitList{}
	results := []report.Result{}
	for i := range repoCommits.RepoCommits {
		repo := repoCommits.RepoCommits[i]
		wg.Add(1)
		go func() {
			authors := urlAuthors.GetAuthors(repo.enterpriseUrl)
			userCommits := github.CommitList{
				Commits: github.GetUserCommits(repo.commits.Commits, authors),
			}

			saveErr := artifacts.SaveDiffToFiles(remoteClients.Get(repo.enterpriseUrl), &userCommits, artifacts.Options{
				Org:     repo.org,
				Repo:    repo.repo,
				Authors: authors,
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

func getUsernames(user User) string {
	users := []string{}
	if user.Username != "" {
		users = append(users, user.Username)
	}

	for _, enterpriseUsername := range user.EnterpriseUsernames {
		users = append(users, enterpriseUsername)
	}

	return strings.Join(users, ", ")
}

func sanitizeOutputDir(dir string) (string, error) {
	outputDir, err := filepath.Abs(dir)
	if err != nil {
		return outputDir, err
	}

	return outputDir, os.MkdirAll(outputDir, os.ModePerm)
}

func SplitRemoteName(remote string) (string, string) {
	repoOrg := strings.Split(remote, "/")
	return repoOrg[0], repoOrg[1]
}
