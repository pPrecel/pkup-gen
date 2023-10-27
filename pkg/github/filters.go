package github

import (
	"fmt"

	"github.com/google/go-github/v53/github"
	"github.com/pterm/pterm"
)

type FilterFunc func(*pterm.Logger, []*github.PullRequest, ListUserPRsOpts) []*github.PullRequest

func FilterPRsByClosedAt(log *pterm.Logger, prs []*github.PullRequest, opts ListUserPRsOpts) []*github.PullRequest {
	filtered := []*github.PullRequest{}
	for i := range prs {
		pr := *prs[i]

		if pr.GetMergedAt().IsZero() && pr.GetClosedAt().Before(opts.MergedBefore) && pr.GetClosedAt().After(opts.MergedAfter) {
			filtered = append(filtered, &pr)
		}

	}

	log.Debug(fmt.Sprintf("%d PRs closed in the period on this page", len(filtered)), log.Args(
		"org", opts.Org,
		"repo", opts.Repo,
	))
	return filtered
}

func FilterPRsByMergedAt(log *pterm.Logger, prs []*github.PullRequest, opts ListUserPRsOpts) []*github.PullRequest {
	filtered := []*github.PullRequest{}
	for i := range prs {
		pr := *prs[i]

		if pr.GetMergedAt().Before(opts.MergedBefore) && pr.GetMergedAt().After(opts.MergedAfter) {
			filtered = append(filtered, &pr)
		}

	}

	log.Debug(fmt.Sprintf("%d PRs merged in the period on this page", len(filtered)), log.Args(
		"org", opts.Org,
		"repo", opts.Repo,
	))
	return filtered
}

func (gh *gh_client) fireFilters(prs []*github.PullRequest, opts ListUserPRsOpts, filters []FilterFunc) []*github.PullRequest {
	filtered := []*github.PullRequest{}
	for i := range filters {
		filtered = append(filtered, filters[i](gh.log, prs, opts)...)
	}

	return filtered
}
