package github

import (
	"sort"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/sirupsen/logrus"
)

const (
	perPage = 100

	// there is no sense to list more than perPage * maxPage pages
	// for example kyma has more than 12k closed PRs
	maxPage = 15
)

type Options struct {
	Org, Repo, Username       string
	MergedBefore, MergedAfter time.Time
}

func (gh *gh_client) ListUserPRsForRepo(opts Options) ([]*github.PullRequest, error) {
	userPullRequests := []*github.PullRequest{}
	page := 1

	for page <= maxPage {
		prs, wasLast, err := gh.listUserPRsForRepo(opts, page)
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

func (gh *gh_client) listUserPRsForRepo(opts Options, page int) ([]*github.PullRequest, bool, error) {
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

	gh.log.Debugf("listed %d PRs on the page", len(pagePRs))

	sort.Slice(pagePRs, func(i, j int) bool {
		return pagePRs[i].GetMergedAt().After(
			pagePRs[j].GetMergedAt().Time,
		)
	})

	filtered := filterPRsByMergedAt(gh.log, pagePRs, opts)
	pullRequests, err := gh.listUserPRs(filtered, opts)

	gh.log.Debugf("\t%d PRs are related with user %s", len(pullRequests), opts.Username)
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
			gh.log.Debugf("\t%s is NOT one of the authors of the '%s'", opts.Username, pr.GetTitle())
			continue
		}

		gh.log.Debugf("\t%s is one of the authors of the '%s'", opts.Username, pr.GetTitle())
		userPRs = append(userPRs, pr)
	}

	return userPRs, nil
}

func filterPRsByMergedAt(log *logrus.Logger, prs []*github.PullRequest, opts Options) []*github.PullRequest {
	filtered := []*github.PullRequest{}
	for i := range prs {
		pr := *prs[i]

		if pr.GetMergedAt().Before(opts.MergedBefore) && pr.GetMergedAt().After(opts.MergedAfter) {
			filtered = append(filtered, &pr)
		}

	}

	log.Debugf("\t%d PRs in the period on this page", len(filtered))
	return filtered
}

func isAuthorOrCommitter(log *logrus.Logger, commits []*github.RepositoryCommit, userName string) bool {
	for i := range commits {
		commit := commits[i]

		log.Debugf("\t\t'%s' is author and '%s' is committer of '%s'",
			commit.Author.GetLogin(),
			commit.Committer.GetLogin(),
			commit.Commit.GetMessage(),
		)

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
