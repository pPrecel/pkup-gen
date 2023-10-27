package github

import (
	"sort"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/pterm/pterm"
)

const (
	perPage = 100

	// there is no sense to list more than perPage * maxPage pages
	// for example kyma has more than 12k closed PRs
	maxPage = 15
)

type ListUserPRsOpts struct {
	Org          string
	Repo         string
	Username     string
	MergedBefore time.Time
	MergedAfter  time.Time
}

// depricated
func (gh *gh_client) ListUserPRsForRepo(opts ListUserPRsOpts, filters []FilterFunc) ([]*github.PullRequest, error) {
	allUserPRs, err := gh.listLastPRsForRepo(opts, filters)
	if err != nil {
		return nil, err
	}

	sorted := sortPRsByMergedAt(allUserPRs)
	filtered := gh.fireFilters(sorted, opts, filters)

	pullRequests, err := gh.listUserPRs(filtered, opts)
	if err != nil {
		return nil, err
	}

	gh.log.Trace("PRs related with user", gh.log.Args(
		"org", opts.Org,
		"repo", opts.Repo,
		"username", opts.Username,
		"count", len(pullRequests),
	))

	return pullRequests, nil
}

func (gh *gh_client) listLastPRsForRepo(opts ListUserPRsOpts, filters []FilterFunc) ([]*github.PullRequest, error) {
	userPullRequests := []*github.PullRequest{}
	page := 1

	for page <= maxPage {
		pagePRs, _, err := gh.client.PullRequests.List(gh.ctx, opts.Org, opts.Repo, &github.PullRequestListOptions{
			State: "closed",
			ListOptions: github.ListOptions{
				PerPage: perPage, // max
				Page:    page,
			},
		})
		if err != nil {
			return nil, err
		}

		gh.log.Trace("prs on the page", gh.log.Args(
			"org", opts.Org,
			"repo", opts.Repo,
			"page", page,
			"prs", len(pagePRs),
		))

		userPullRequests = append(userPullRequests, pagePRs...)

		if len(pagePRs) < perPage {
			break
		}
		page++
	}

	return userPullRequests, nil
}

func (gh *gh_client) listUserPRs(prs []*github.PullRequest, opts ListUserPRsOpts) ([]*github.PullRequest, error) {
	userPRs := []*github.PullRequest{}
	for i := range prs {
		pr := prs[i]
		commits, _, err := gh.client.PullRequests.ListCommits(gh.ctx, opts.Org, opts.Repo, pr.GetNumber(), &github.ListOptions{
			PerPage: 100,
		})
		if err != nil {
			return nil, err
		}

		if !isAuthorOrCommitter(gh.log, commits, opts.Username) {
			gh.log.Trace("user is NOT one of the authors of the pr", gh.log.Args(
				"org", opts.Org,
				"repo", opts.Repo,
				"username", opts.Username,
				"pr", pr.GetTitle(),
				"prURL", pr.GetHTMLURL(),
			))
			continue
		}

		gh.log.Debug("user is one of the authors of the pr", gh.log.Args(
			"org", opts.Org,
			"repo", opts.Repo,
			"username", opts.Username,
			"pr", pr.GetTitle(),
			"prURL", pr.GetHTMLURL(),
		))
		userPRs = append(userPRs, pr)
	}

	return userPRs, nil
}

func isAuthorOrCommitter(log *pterm.Logger, commits []*github.RepositoryCommit, userName string) bool {
	for i := range commits {
		commit := commits[i]

		log.Trace("user and author on the commit", log.Args(
			"author", commit.Author.GetLogin(),
			"commiter", commit.Committer.GetLogin(),
			"commitSHA", commit.GetSHA(),
			"commitMsg", commit.Commit.GetMessage(),
		))

		if commit != nil &&
			((commit.Author != nil &&
				commit.Author.GetLogin() == userName) ||
				(commit.Committer != nil &&
					commit.Committer.GetLogin() == userName)) {

			return true
		}
	}

	return false
}

func sortPRsByMergedAt(prs []*github.PullRequest) []*github.PullRequest {
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].GetMergedAt().After(
			prs[j].GetMergedAt().Time,
		)
	})
	return prs
}
