package github

func (gh *gh_client) GetUserSignatures(username string) ([]string, error) {
	user, _, err := gh.client.Users.Get(gh.ctx, username)
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
