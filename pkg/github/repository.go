package github

import "github.com/google/go-github/v53/github"

// TODO: generate PKUP report for all repos in org
func (gh *gh_client) listRepositories(org string) ([]*github.Repository, error) {

	gh.client.Repositories.ListByOrg(gh.ctx, org, &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: perPage,
		},
	})

	return nil, nil
}
