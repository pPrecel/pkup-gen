package generator

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pterm/pterm"
)

type LazyCommitsLister interface {
	List(*Config, *ComposeOpts) (*RepoCommitsList, error)
}

type RepoCommitsList struct {
	RepoCommits []RepoCommits
}

type RepoCommits struct {
	org           string
	repo          string
	enterpriseUrl string
	commits       *github.CommitList
}

type lazyRepoCommitsLister struct {
	mutex           sync.Mutex
	repoCommitsList *RepoCommitsList

	remoteClients *remoteClients
	logger        *pterm.Logger
}

func NewLazyRepoCommitsLister(logger *pterm.Logger, remoteClients *remoteClients) LazyCommitsLister {
	return &lazyRepoCommitsLister{
		remoteClients: remoteClients,
		logger:        logger,
	}
}

// list commits if were lister before
// if not then list them from remote
func (ll *lazyRepoCommitsLister) List(config *Config, opts *ComposeOpts) (*RepoCommitsList, error) {
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
				Since:      opts.Since,
				Until:      opts.Until,
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
				org:           orgName,
				repo:          repoName,
				enterpriseUrl: repo.EnterpriseUrl,
				commits:       commitList,
			}
		}()
	}

	wg.Wait()

	ll.repoCommitsList = &RepoCommitsList{
		RepoCommits: allRepoCommits,
	}
	return ll.repoCommitsList, err
}

func (ll *lazyRepoCommitsLister) listOrgRepos(remoteClients *remoteClients, config *Config) ([]Remote, error) {
	remotes := []Remote{}

	// resolve orgs
	for _, org := range config.Orgs {
		c := remoteClients.Get(org.EnterpriseUrl)

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

func containsOrgRepo(remotes []Remote, orgRepo string) bool {
	for _, remote := range remotes {
		if remote.Name == orgRepo {
			return true
		}
	}

	return false
}
