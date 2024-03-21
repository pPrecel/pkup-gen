package utils

import (
	"fmt"

	"github.com/pPrecel/PKUP/pkg/compose/config"
)

// use to get GitHub username
const DefaultGitHubAuthor = ""

type UrlAuthors map[string][]string

func (ua UrlAuthors) GetAuthors(url string) []string {
	return ua[url]
}

func (ua UrlAuthors) set(url string, authors []string) {
	ua[url] = authors
}

func BuildUrlAuthors(remoteClients *RemoteClients, user *config.User) (*UrlAuthors, error) {
	authorsMap := &UrlAuthors{}

	// get signatures for opensource if not empty
	if user.Username != "" {
		signatures, err := remoteClients.Get(DefaultGitHubURL).GetUserSignatures(user.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to list user signatures for opensource: %s", err.Error())
		}

		authorsMap.set(DefaultGitHubAuthor, signatures)
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

		authorsMap.set(url, signatures)
	}

	return authorsMap, nil
}
