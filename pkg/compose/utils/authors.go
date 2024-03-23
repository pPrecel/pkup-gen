package utils

import (
	"fmt"

	"github.com/pPrecel/PKUP/pkg/config"
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

func BuildUrlAuthors(remoteClients *RemoteClients, usernames []config.Signature) (*UrlAuthors, error) {
	authorsMap := &UrlAuthors{}

	for _, u := range usernames {
		c := remoteClients.Get(u.EnterpriseUrl)
		if c == nil {
			// client is not specified for this enterpriseUrl
			continue
		}

		signatures, err := c.GetUserSignatures(u.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to list user signatures for '%s': %s", u.Username, err.Error())
		}

		authorsMap.set(u.EnterpriseUrl, signatures)
	}

	return authorsMap, nil
}
