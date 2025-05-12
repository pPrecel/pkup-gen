package github

import (
	"fmt"

	go_github "github.com/google/go-github/v53/github"
)

type repoList struct {
	repos []*go_github.Repository
}

func (gh *gh_client) ListRepos(org string) ([]string, error) {
	repoList := &repoList{
		repos: []*go_github.Repository{},
	}

	err := listForPages(gh.listReposPageFunc(repoList, org))
	if err != nil {
		return nil, fmt.Errorf("failed to list branches for org '%s': %s", org, err)
	}

	repos := []string{}
	for _, repo := range repoList.repos {
		repos = append(repos, repo.GetName())
	}

	return repos, nil
}

func (gh *gh_client) listReposPageFunc(dest *repoList, org string) pageListFunc {
	return func(page int) (bool, error) {
		perPage := 100
		var repos []*go_github.Repository
		var err error
		err = gh.callWithRateLimitRetry(func() error {
			repos, _, err = gh.client.Repositories.ListByOrg(gh.ctx, org, &go_github.RepositoryListByOrgOptions{
				ListOptions: go_github.ListOptions{
					Page:    page,
					PerPage: perPage,
				},
			})
			return err
		})
		if err != nil {
			return false, err
		}

		dest.repos = append(dest.repos, repos...)
		return len(repos) == perPage, nil
	}
}
