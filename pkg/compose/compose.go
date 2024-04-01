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
	"github.com/pPrecel/PKUP/pkg/compose/utils"
	"github.com/pPrecel/PKUP/pkg/config"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pPrecel/PKUP/pkg/report"
	"github.com/pterm/pterm"
)

//go:generate mockery --name=Compose --output=automock --outpkg=automock --case=underscore
type Compose interface {
	ForConfig(*config.Config, Options) error
}

type compose struct {
	ctx         context.Context
	logger      *pterm.Logger
	buildClient utils.BuildClientFunc

	repoCommitsLister utils.LazyCommitsLister
}

func New(ctx context.Context, logger *pterm.Logger) Compose {
	return &compose{
		ctx:         ctx,
		logger:      logger,
		buildClient: github.NewClient,
	}
}

type Options struct {
	Since time.Time
	Until time.Time
	Ci    bool
}

func (c *compose) ForConfig(config *config.Config, opts Options) error {
	taskView := view.NewMultiTaskView(c.logger, opts.Ci)
	viewLogger := c.logger.WithWriter(taskView.NewWriter())

	remoteClients, err := utils.BuildClients(c.ctx, c.logger, config, c.buildClient)
	if err != nil {
		return err
	}

	c.repoCommitsLister = utils.NewLazyRepoCommitsLister(c.logger, remoteClients)

	for i := range config.Reports {
		user := config.Reports[i]

		valChan := make(chan []*view.RepoCommit)
		errChan := make(chan error)
		taskView.Add(getUsernames(user), valChan, errChan)

		go func() {
			viewLogger.Debug("compose for user", viewLogger.Args("user", getUsernames(user)))
			commitList, err := c.composeForUser(remoteClients, &user, config, &opts)
			if err != nil {
				errChan <- err
				return
			}

			valChan <- commitList
		}()
	}

	return taskView.Run()
}

func (c *compose) composeForUser(remoteClients *utils.RemoteClients, user *config.Report, config *config.Config, opts *Options) ([]*view.RepoCommit, error) {
	outputDir, err := sanitizeOutputDir(user.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize path '%s': %s", user.OutputDir, err.Error())
	}

	urlAuthors, err := utils.BuildUrlAuthors(remoteClients, user.Signatures)
	if err != nil {
		return nil, fmt.Errorf("failed to list user signatures: %s", err.Error())
	}

	repoCommits, err := c.repoCommitsLister.List(config, opts.Since, opts.Until)
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %s", err.Error())
	}

	wg := sync.WaitGroup{}
	var errors error
	commitList := []*view.RepoCommit{}
	results := []report.Result{}
	for i := range repoCommits.RepoCommits {
		repo := repoCommits.RepoCommits[i]
		wg.Add(1)
		go func() {
			authors := urlAuthors.GetAuthors(repo.EnterpriseUrl)
			userCommits := github.CommitList{
				Commits: github.GetUserCommits(repo.Commits.Commits, authors),
			}

			saveErr := artifacts.SaveDiffToFiles(remoteClients.Get(repo.EnterpriseUrl), &userCommits, artifacts.Options{
				Org:     repo.Org,
				Repo:    repo.Repo,
				Authors: authors,
				Dir:     outputDir,
				Since:   opts.Since,
				Until:   opts.Until,
			})
			if saveErr != nil {
				errors = multierror.Append(errors, fmt.Errorf(
					"failed to generate artifacts for repo '%s': %s", repo.Repo, saveErr.Error(),
				))
			} else {
				// url := repo.enterpriseUrl
				// if url == "" {
				// 	url = "https://github.com"
				// }

				for _, commit := range userCommits.Commits {
					repoCommit := &view.RepoCommit{
						Org:     repo.Org,
						Repo:    repo.Repo,
						Message: strings.Split(commit.Commit.GetMessage(), "\n")[0],
						SHA:     commit.GetSHA(),
					}
					commitList = append(commitList, repoCommit)

					c.logger.Trace(
						fmt.Sprintf("found commit for user %s", getUsernames(*user)),
						c.logger.Args(
							"org/repo", fmt.Sprintf("%s/%s", repoCommit.Org, repoCommit.Repo),
							"sha", repoCommit.SHA,
							"message", repoCommit.Message,
						),
					)
				}

				results = append(results, report.Result{
					Org:  repo.Org,
					Repo: repo.Repo,
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

	if config.Template != "" {
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
			CustomValues: user.ExtraFields,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to render report: %s", err.Error())
		}
	}

	return commitList, nil
}

func getUsernames(user config.Report) string {
	users := []string{}
	for _, u := range user.Signatures {
		users = append(users, u.Username)
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
