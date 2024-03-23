package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pPrecel/PKUP/pkg/config"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pterm/pterm"
)

type LazyCommitsLister interface {
	List(*config.Config, time.Time, time.Time) (*RepoCommitsList, error)
}

type RepoCommitsList struct {
	RepoCommits []RepoCommits
}

type RepoCommits struct {
	Org           string
	Repo          string
	EnterpriseUrl string
	Commits       *github.CommitList
}

type lazyRepoCommitsLister struct {
	mutex           sync.Mutex
	repoCommitsList *RepoCommitsList

	remoteClients *RemoteClients
	logger        *pterm.Logger
}

func NewLazyRepoCommitsLister(logger *pterm.Logger, remoteClients *RemoteClients) LazyCommitsLister {
	return &lazyRepoCommitsLister{
		remoteClients: remoteClients,
		logger:        logger,
	}
}

// list commits if were lister before
// if not then list them from remote
func (ll *lazyRepoCommitsLister) List(config *config.Config, since, until time.Time) (*RepoCommitsList, error) {
	ll.mutex.Lock()
	defer ll.mutex.Unlock()

	// return if commits were lister before
	if ll.repoCommitsList != nil {
		return ll.repoCommitsList, nil
	}

	repos, err := ll.listOrgRepos(ll.remoteClients, config)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories for orgs: %s", err.Error())
	}

	wg := sync.WaitGroup{}
	allRepoCommits := make([]RepoCommits, len(repos))
	for i, r := range repos {
		iter := i
		repo := r

		wg.Add(1)
		go func() {
			defer wg.Done()

			orgName, repoName := SplitRemoteName(repo.Name)
			client := ll.remoteClients.Get(repo.EnterpriseUrl)

			ll.logger.Trace("listing commits for repo", ll.logger.Args("org", orgName, "repo", repoName))
			commitList, listErr := client.ListRepoCommits(github.ListRepoCommitsOpts{
				Org:        orgName,
				Repo:       repoName,
				Since:      since,
				Until:      until,
				Branches:   repo.Branches,
				UniqueOnly: repo.UniqueOnly,
			})
			if listErr != nil {
				ll.logger.Warn("failed to list commits", ll.logger.Args("org", orgName, "repo", repoName, "error", listErr.Error()))
				multierror.Append(err, listErr)
				return
			}

			ll.logger.Debug("found commits", ll.logger.Args("org", orgName, "repo", repoName, "count", len(commitList.Commits)))
			allRepoCommits[iter] = RepoCommits{
				Org:           orgName,
				Repo:          repoName,
				EnterpriseUrl: repo.EnterpriseUrl,
				Commits:       commitList,
			}
		}()
	}

	wg.Wait()

	ll.repoCommitsList = &RepoCommitsList{
		RepoCommits: allRepoCommits,
	}
	return ll.repoCommitsList, err
}

func (ll *lazyRepoCommitsLister) listOrgRepos(remoteClients *RemoteClients, cfg *config.Config) ([]config.Remote, error) {
	remotes := []config.Remote{}

	// resolve orgs
	for _, org := range cfg.Orgs {
		c := remoteClients.Get(org.EnterpriseUrl)

		repos, err := c.ListRepos(org.Name)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			name := fmt.Sprintf("%s/%s", org.Name, repo)

			if containsOrgRepo(cfg.Repos, name) {
				// skip if repo already is in config.Repos
				continue
			}

			remotes = append(remotes, config.Remote{
				Name:          name,
				EnterpriseUrl: org.EnterpriseUrl,
				Token:         org.Token,
				Branches:      org.Branches,
				AllBranches:   org.AllBranches,
				UniqueOnly:    org.UniqueOnly,
			})
		}
	}

	remotes = append(cfg.Repos, remotes...)

	// check if remote has AllBranches set
	for i, remote := range remotes {
		c := remoteClients.Get(remote.EnterpriseUrl)
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

func containsOrgRepo(remotes []config.Remote, orgRepo string) bool {
	for _, remote := range remotes {
		if remote.Name == orgRepo {
			return true
		}
	}

	return false
}
