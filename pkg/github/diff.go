package github

import (
	"github.com/google/go-github/v53/github"
)

const (
	emptyDiff = ""
)

func (gh *gh_client) GetCommitContentDiff(commit *github.RepositoryCommit, org, repo string) (string, error) {
	return gh.getContentDiff(commit.GetSHA(), org, repo)
}

func (gh *gh_client) getContentDiff(sha, org, repo string) (string, error) {
	diff, _, err := retryOnRateLimit(gh.log, func() (string, *github.Response, error) {
		return gh.client.Repositories.GetCommitRaw(
			gh.ctx,
			org,
			repo,
			sha,
			github.RawOptions{
				Type: github.Diff,
			},
		)
	})
	if err != nil {
		return emptyDiff, err
	}

	gh.log.Trace("got diff for commit", gh.log.Args(
		"org", org,
		"repo", repo,
		"diffLen", len(diff),
	))

	return diff, nil
}
