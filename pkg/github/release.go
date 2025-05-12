package github

import (
	"fmt"

	"github.com/google/go-github/v53/github"
)

func (gh *gh_client) GetLatestReleaseOrZero(org, repo string) (string, error) {
	var release *github.RepositoryRelease
	var err error
	err = gh.callWithRateLimitRetry(func() error {
		release, _, err = gh.client.Repositories.GetLatestRelease(gh.ctx, org, repo)
		return err
	})
	return fmt.Sprintf("v%s", release.GetTagName()), err
}
