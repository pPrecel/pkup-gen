package github

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/v53/github"
)

const (
	diffUrlFormat = "https://github.com/%s/%s/commit/%s.diff"
	emptyDiff     = ""
)

func (gh *gh_client) GetFileDiffForPRs(prs []*github.PullRequest, org, repo string) (string, error) {
	diff := emptyDiff

	for i := range prs {
		pr := *prs[i]
		url := fmt.Sprintf(diffUrlFormat, org, repo, pr.GetMergeCommitSHA())

		req, err := gh.client.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return emptyDiff, err
		}

		resp, err := gh.client.Client().Do(req)
		if err != nil {
			return emptyDiff, err
		}

		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return emptyDiff, err
		}

		diff = fmt.Sprintf("%s\n\n%s\n", diff, string(bytes))
	}

	return diff, nil
}
