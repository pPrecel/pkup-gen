package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	go_github "github.com/google/go-github/v53/github"
)

type CommitList struct {
	Commits []*go_github.RepositoryCommit
}

func (cl *CommitList) Append(from *CommitList) {
	cl.Commits = append(cl.Commits, from.Commits...)
}

type ListRepoCommitsOpts struct {
	Org        string
	Repo       string
	Authors    []string
	Branches   []string
	UniqueOnly bool
	Since      time.Time
	Until      time.Time
}

func (gh *gh_client) ListRepoCommits(opts ListRepoCommitsOpts) (*CommitList, error) {
	commits := &CommitList{
		Commits: []*go_github.RepositoryCommit{},
	}

	// default branches to HEAD if empty
	if len(opts.Branches) == 0 {
		opts.Branches = []string{""}
	}

	for _, branch := range opts.Branches {
		// get all repo commits in given period
		err := gh.listForPages(listCommitsPageFunc(gh.ctx, gh.client, commits, listForPageOpts{
			org:    opts.Org,
			repo:   opts.Repo,
			branch: branch,
			since:  opts.Since,
			until:  opts.Until,
		}))
		if err != nil {
			return nil, fmt.Errorf("failed to list commits for repo '%s/%s': %s", opts.Org, opts.Repo, err.Error())
		}
	}

	// filter out not user commits
	if len(opts.Authors) > 0 {
		commits.Commits = GetUserCommits(commits.Commits, opts.Authors)
	}

	// remove same commits from different branches
	if opts.UniqueOnly {
		removeDuplicates(commits)
	}

	return commits, nil
}

func GetUserCommits(commits []*go_github.RepositoryCommit, authors []string) []*go_github.RepositoryCommit {
	userCommits := []*go_github.RepositoryCommit{}

	for _, commit := range commits {
		for _, author := range authors {
			if isVerifiedCommitAuthor(commit, author) ||
				isRepositoryCommitAuthor(commit, author) ||
				isCommitAuthor(commit.Commit, author) {

				userCommits = append(userCommits, commit)
				break
			}
		}
	}

	return userCommits
}

type listForPageOpts struct {
	org    string
	repo   string
	branch string
	since  time.Time
	until  time.Time
}

func listCommitsPageFunc(ctx context.Context, client *go_github.Client, dest *CommitList, opts listForPageOpts) pageListFunc {
	return func(page int) (bool, error) {
		perPage := 100
		commits, resp, listErr := client.Repositories.ListCommits(ctx, opts.org, opts.repo, &go_github.CommitsListOptions{
			SHA:   opts.branch,
			Since: opts.since,
			Until: opts.until,
			ListOptions: go_github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		})
		// return error only when statusCode is not 409 (repo is empty)
		if listErr != nil && resp.StatusCode != 409 {
			return false, listErr
		}

		dest.Commits = append(dest.Commits, commits...)
		return len(commits) == perPage, nil
	}
}

func isVerifiedCommitAuthor(commit *go_github.RepositoryCommit, author string) bool {
	if commit.Commit == nil ||
		commit.Commit.Verification == nil ||
		commit.Commit.Verification.Verified == nil ||
		!*commit.Commit.Verification.Verified {
		return false
	}

	payloadLines := strings.Split(*commit.Commit.Verification.Payload, "\n")
	for i := range payloadLines {
		line := payloadLines[i]
		// check if user is author of the commit based on the payload fields
		// example payload:
		// tree d880fdb45b81e17eb18c270b8ac835d8de3e92e0
		// parent 1c1b51c12888f2e8275aa92a48d6fb96fb70d4f3
		// author Filip Strózik <filip.strozik@outlook.com> 1697452255 +0200
		// committer GitHub <noreply@github.com> 1697452255 +0200
		//
		// Reflect used presets in status (#351)
		//
		// Co-authored-by: Marcin Dobrochowski <anoip@o2.pl>"
		if strings.HasPrefix(line, fmt.Sprintf("author %s ", author)) ||
			strings.HasPrefix(line, fmt.Sprintf("Co-authored-by: %s ", author)) {
			return true
		}
	}

	return false
}

func isCommitAuthor(commit *go_github.Commit, author string) bool {
	if commit == nil || commit.Author == nil {
		return false
	}

	if commit.Author.Login != nil &&
		*commit.Author.Login == author {
		return true
	}

	if commit.Author.Name != nil &&
		*commit.Author.Name == author {
		return true
	}

	return false
}

func isRepositoryCommitAuthor(commit *go_github.RepositoryCommit, author string) bool {
	if commit == nil || commit.Author == nil {
		return false
	}

	if commit.Author.Login != nil &&
		*commit.Author.Login == author {
		return true
	}

	if commit.Author.Name != nil &&
		*commit.Author.Name == author {
		return true
	}

	return false
}

func removeDuplicates(commitList *CommitList) {
	commits := []*go_github.RepositoryCommit{}
	for _, commit := range commitList.Commits {
		if !isInCommits(commits, commit) {
			commits = append(commits, commit)
		}
	}

	commitList.Commits = commits
}

func isInCommits(where []*go_github.RepositoryCommit, what *go_github.RepositoryCommit) bool {
	for _, commit := range where {
		if commit.GetSHA() == what.GetSHA() {
			return true
		}
	}
	return false
}

type pageListFunc func(page int) (nextPage bool, err error)

func (gh *gh_client) listForPages(fn pageListFunc) error {
	page := 1
	nextPage := true
	for nextPage {
		var err error
		nextPage, err = gh.execListFuncWithRetry(fn, page)
		if err != nil {
			return err
		}

		page++
	}

	return nil
}

func (gh *gh_client) execListFuncWithRetry(fn pageListFunc, page int) (bool, error) {
	var nextPage bool
	var err error
	for i := 0; i < 5; i++ {
		nextPage, err = fn(page)
		if err == nil {
			break
		}

		if isRateLimitErr(err) {
			// rate limit reached
			// wait for the reset time
			// https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api?apiVersion=2022-11-28
			d := getRateLimitResetDuration(err)
			gh.log.Warn("Rate limit exceeded, waiting: %s", gh.log.Args("duration", d))
			time.Sleep(d)
			continue
		}

		// return any other error
		return false, err
	}

	return nextPage, err
}

func isRateLimitErr(err error) bool {
	if err == nil {
		return false
	}

	switch e := err.(type) {
	case *go_github.ErrorResponse:
		// common error response when reaching the same endpoint too many times
		return e.Response.StatusCode == 403
	case *go_github.RateLimitError:
		// specific error response when reaching the rate limit
		return true
	default:
		return false
	}
}

func getRateLimitResetDuration(err error) time.Duration {
	switch e := err.(type) {
	case *go_github.RateLimitError:
		return time.Until(e.Rate.Reset.Time)
	default:
		// default to 1 minute if we can't determine the reset time
		return time.Minute
	}
}
