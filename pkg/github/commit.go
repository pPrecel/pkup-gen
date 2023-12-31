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
	Authors []string
	Org     string
	Repo    string
	Branch  string
	Since   time.Time
	Until   time.Time
}

func (gh *gh_client) ListRepoCommits(opts ListRepoCommitsOpts) (*CommitList, error) {
	commits := &CommitList{
		Commits: []*go_github.RepositoryCommit{},
	}

	// get all repo commits in given period
	err := listForPages(listCommitsPageFunc(gh.ctx, gh.client, commits, opts))
	if err != nil {
		return nil, fmt.Errorf("failed to list commits for repo '%s/%s': %s", opts.Org, opts.Repo, err.Error())
	}

	// filter out not user commits
	if len(opts.Authors) > 0 {
		commits.Commits = GetUserCommits(commits.Commits, opts.Authors)
	}

	// client.Repositories.GetCommit(ctx, org, repo, "sha", &go_github.ListOptions{} )

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

func listCommitsPageFunc(ctx context.Context, client *go_github.Client, dest *CommitList, opts ListRepoCommitsOpts) pageListFunc {
	return func(page int) (bool, error) {
		perPage := 100
		commits, resp, listErr := client.Repositories.ListCommits(ctx, opts.Org, opts.Repo, &go_github.CommitsListOptions{
			SHA:   opts.Branch,
			Since: opts.Since,
			Until: opts.Until,
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

type pageListFunc func(page int) (nextPage bool, err error)

func listForPages(fn pageListFunc) error {
	page := 1
	nextPage := true
	for nextPage {
		var err error
		nextPage, err = fn(page)
		if err != nil {
			return err
		}

		page++
	}

	return nil
}
