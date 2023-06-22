package github

import (
	"fmt"

	"github.com/google/go-github/v53/github"
)

const (
	emptyDiff = ""
)

func (gh *gh_client) GetFileDiffForPRs(prs []*github.PullRequest, org, repo string) (string, error) {
	diff := emptyDiff

	for i := range prs {
		pr := *prs[i]

		d, _, err := gh.client.Repositories.GetCommitRaw(
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

		diff += fmt.Sprintf("%s\n", d)
	}

	return diff, nil
}
