package github

import "fmt"

func (gh *gh_client) GetLatestReleaseOrZero(org, repo string) (string, error) {
	release, _, err := gh.client.Repositories.GetLatestRelease(gh.ctx, org, repo)
	return fmt.Sprintf("v%s", release.GetTagName()), err
}
