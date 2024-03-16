package compose

import (
	"fmt"

	"github.com/pPrecel/PKUP/pkg/github"
)

func buildUrlAuthors(remoteClients map[string]github.Client, user *User) (map[string][]string, error) {
	authorsMap := map[string][]string{}

	// get signatures for opensource if not empty
	if user.Username != "" {
		signatures, err := remoteClients[""].GetUserSignatures(user.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to list user signatures for opensource: %s", err.Error())
		}

		authorsMap[""] = signatures
	}

	// get signatures for every enterprise
	for url, username := range user.EnterpriseUsernames {
		c, ok := remoteClients[url]
		if !ok {
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
