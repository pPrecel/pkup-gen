package generator

import (
	"fmt"
)

func buildUrlAuthors(remoteClients *remoteClients, user *User) (map[string][]string, error) {
	authorsMap := map[string][]string{}

	// get signatures for opensource if not empty
	if user.Username != "" {
		signatures, err := remoteClients.Get("").GetUserSignatures(user.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to list user signatures for opensource: %s", err.Error())
		}

		authorsMap[""] = signatures
	}

	// get signatures for every enterprise
	for url, username := range user.EnterpriseUsernames {
		c := remoteClients.Get(url)
		if c == nil {
			// enterprise user is specified by not enterpriese org/repo
			continue
		}
		signatures, err := c.GetUserSignatures(username)
		if err != nil {
			return nil, fmt.Errorf("failed to list user signatures for '%s': %s", url, err.Error())
		}

		authorsMap[url] = signatures
	}

	return authorsMap, nil
}
