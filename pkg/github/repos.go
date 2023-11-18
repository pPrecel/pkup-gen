package github

import (
	"context"
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

	err := listForPages(listReposPageFunc(gh.ctx, gh.client, repoList, org))
	if err != nil {
		return nil, fmt.Errorf("failed to list branches for org '%s': %s", org, err)
	}

	repos := []string{}
	for _, repo := range repoList.repos {
		repos = append(repos, repo.GetName())
	}

	return repos, nil
}

func listReposPageFunc(ctx context.Context, client *go_github.Client, dest *repoList, org string) pageListFunc {
	return func(page int) (bool, error) {
		perPage := 100
		repos, _, err := client.Repositories.ListByOrg(ctx, org, &go_github.RepositoryListByOrgOptions{
			ListOptions: go_github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		})
		if err != nil {
			return false, err
		}

		dest.repos = append(dest.repos, repos...)
		return len(repos) == perPage, nil
	}
}
