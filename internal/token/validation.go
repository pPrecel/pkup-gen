package token

import (
	"fmt"
	"net/http"

	"github.com/pterm/pterm"
)

func isTokenValid(httpClient *http.Client, logger *pterm.Logger, githubApiHostname, token string) bool {
	// do request to the test github endpoint to validate if token is up-to-date
	url := fmt.Sprintf("%s/octocat", githubApiHostname)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Trace("can't call github", logger.Args(
			"error", err,
		))
		return false
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Add("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Trace("token is not valid", logger.Args(
			"status code", resp.StatusCode,
			"error", err,
		))
		return false
	}

	return true
}
