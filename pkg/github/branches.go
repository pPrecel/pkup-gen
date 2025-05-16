package github

import (
	"fmt"

	go_github "github.com/google/go-github/v53/github"
)

type BranchList struct {
	Branches []string
}

func (gh *gh_client) ListRepoBranches(org, repo string) (*BranchList, error) {
	branchList := &BranchList{}
	err := listForPages(gh.listBranchesForPage(branchList, org, repo))
	if err != nil {
		return nil, fmt.Errorf("failed to list branches for repo '%s/%s': %s", org, repo, err.Error())
	}

	return branchList, nil
}

func (gh *gh_client) listBranchesForPage(dest *BranchList, org, repo string) pageListFunc {
	return func(page int) (bool, error) {
		perPage := 100
		branches, resp, err := retryOnRateLimit(gh.log, func() ([]*go_github.Branch, *go_github.Response, error) {
			return gh.client.Repositories.ListBranches(gh.ctx, org, repo, &go_github.BranchListOptions{
				ListOptions: go_github.ListOptions{
					Page:    page,
					PerPage: perPage,
				},
			})
		})
		// return error only when statusCode is not 409 (repo is empty)
		if err != nil && resp.StatusCode != 409 {
			return false, err
		}

		for _, branch := range branches {
			dest.Branches = append(dest.Branches, branch.GetName())
		}

		return len(branches) == perPage, nil
	}
}
