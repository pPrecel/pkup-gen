package github

import (
	"github.com/google/go-github/v53/github"
)

const (
	emptyDiff = ""
)

func (gh *gh_client) GetPRContentDiff(pr *github.PullRequest, org, repo string) (string, error) {
	diff, _, err := gh.client.Repositories.GetCommitRaw(
		gh.ctx,
		org,
		repo,
		pr.GetMergeCommitSHA(),
		github.RawOptions{
			Type: github.Diff,
		},
	)
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
