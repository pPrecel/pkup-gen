package github

import (
	"fmt"

	"github.com/google/go-github/v53/github"
)

func (gh *gh_client) GetLatestReleaseOrZero(org, repo string) (string, error) {
	release, _, err := retryOnRateLimit(gh.log, func() (*github.RepositoryRelease, *github.Response, error) {
		return gh.client.Repositories.GetLatestRelease(gh.ctx, org, repo)
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("v%s", release.GetTagName()), err
}
