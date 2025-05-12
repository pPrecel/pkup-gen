package github

import "github.com/google/go-github/v53/github"

func (gh *gh_client) GetUserSignatures(username string) ([]string, error) {
	var user *github.User
	var err error
	err = gh.callWithRateLimitRetry(func() error {
		user, _, err = gh.client.Users.Get(gh.ctx, username)
		return err
	})
	if err != nil {
		return nil, err
	}

	signatures := []string{}
	if user.Name != nil {
		signatures = append(signatures, user.GetName())
	}

	if user.Login != nil {
		signatures = append(signatures, user.GetLogin())
	}

	return signatures, nil
}
