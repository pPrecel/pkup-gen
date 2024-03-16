package github

import (
	"context"
	"fmt"

	go_github "github.com/google/go-github/v53/github"
)

type BranchList struct {
	Branches []string
}

func (gh *gh_client) ListRepoBranches(org, repo string) (*BranchList, error) {
	branchList := &BranchList{}
	err := listForPages(listBranchesForPage(gh.ctx, gh.client, branchList, org, repo))
	if err != nil {
		return nil, fmt.Errorf("failed to list branches for repo '%s/%s': %s", org, repo, err.Error())
	}

	return branchList, nil
}

func listBranchesForPage(ctx context.Context, client *go_github.Client, dest *BranchList, org, repo string) pageListFunc {
	return func(page int) (nextPage bool, err error) {
		perPage := 100
		branches, resp, listErr := client.Repositories.ListBranches(ctx, org, repo, &go_github.BranchListOptions{
			ListOptions: go_github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		})
		// return error only when statusCode is not 409 (repo is empty)
		if listErr != nil && resp.StatusCode != 409 {
			return false, listErr
		}

		for _, branch := range branches {
			dest.Branches = append(dest.Branches, branch.GetName())
		}

		return len(branches) == perPage, nil
	}
}
