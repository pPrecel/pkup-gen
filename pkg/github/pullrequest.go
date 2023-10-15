package github

import (
	"fmt"
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

type Options struct {
	Org          string
	Repo         string
	Username     string
	WithClosed   bool
	MergedBefore time.Time
	MergedAfter  time.Time
}

func (gh *gh_client) ListUserPRsForRepo(opts Options) ([]*github.PullRequest, error) {
	userPullRequests := []*github.PullRequest{}
	page := 1

	filters := []filterFunc{filterPRsByMergedAt}
	if opts.WithClosed {
		filters = append(filters, filterPRsByClosedAt)
	}

	for page <= maxPage {
		prs, wasLast, err := gh.listUserPRsForRepo(opts, page, filters)
		if err != nil {
			return nil, err
		}

		userPullRequests = append(userPullRequests, prs...)

		if wasLast {
			break
		}
		page++
	}

	return userPullRequests, nil
}

type filterFunc func(*pterm.Logger, []*github.PullRequest, Options) []*github.PullRequest

func (gh *gh_client) listUserPRsForRepo(opts Options, page int, filters []filterFunc) ([]*github.PullRequest, bool, error) {
	pagePRs, _, err := gh.client.PullRequests.List(gh.ctx, opts.Org, opts.Repo, &github.PullRequestListOptions{
		State: "closed",
		ListOptions: github.ListOptions{
			PerPage: perPage, // max
			Page:    page,
		},
	})
	if err != nil {
		return nil, true, err
	}

	gh.log.Trace(fmt.Sprintf("listed %d PRs on the page", len(pagePRs)), gh.log.Args(
		"org", opts.Org,
		"repo", opts.Repo,
	))

	sort.Slice(pagePRs, func(i, j int) bool {
		return pagePRs[i].GetMergedAt().After(
			pagePRs[j].GetMergedAt().Time,
		)
	})

	filtered := []*github.PullRequest{}
	for i := range filters {
		filtered = append(filtered, filters[i](gh.log, pagePRs, opts)...)
	}

	pullRequests, err := gh.listUserPRs(filtered, opts)

	gh.log.Trace(fmt.Sprintf("%d PRs are related with user %s", len(pullRequests), opts.Username), gh.log.Args(
		"org", opts.Org,
		"repo", opts.Repo,
	))
	return pullRequests,
		len(pagePRs) < perPage,
		err
}

func (gh *gh_client) listUserPRs(prs []*github.PullRequest, opts Options) ([]*github.PullRequest, error) {
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
			))
			continue
		}

		gh.log.Trace("user is one of the authors of the pr", gh.log.Args(
			"org", opts.Org,
			"repo", opts.Repo,
			"username", opts.Username,
			"pr", pr.GetTitle(),
		))
		userPRs = append(userPRs, pr)
	}

	return userPRs, nil
}

func isAuthorOrCommitter(log *pterm.Logger, commits []*github.RepositoryCommit, userName string) bool {
	for i := range commits {
		commit := commits[i]

		log.Trace("user and author on the commit", log.Args(
			"commit", commit.Commit.GetMessage(),
			"author", commit.Author.GetLogin(),
			"commiter", commit.Committer.GetLogin(),
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
