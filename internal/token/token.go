package token

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/user"

	"github.com/cli/oauth/device"
	"github.com/pterm/pterm"
	"github.com/zalando/go-keyring"
)

func Get(logger *pterm.Logger, clientID string) (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	tg := &tokenGetter{
		client:            http.DefaultClient,
		logger:            logger,
		serviceName:       "pkup-gen",
		username:          user.Username,
		githubHostname:    "https://github.com",
		githubAPIHostname: "https://api.github.com",
		clientID:          clientID,
	}
	return tg.do()
}

type tokenGetter struct {
	client            *http.Client
	logger            *pterm.Logger
	serviceName       string
	username          string
	githubHostname    string
	githubAPIHostname string
	clientID          string
}

func (tg *tokenGetter) do() (string, error) {
	token, err := keyring.Get(tg.serviceName, tg.username)
	if err == nil && isTokenValid(tg.client, tg.logger, tg.githubAPIHostname, token) {
		return token, nil
	}

	tg.logger.Trace("getting token from GitHub device")
	token, err = getGitHubDeviceToken(tg.client, tg.logger, tg.githubHostname, tg.clientID)
	if err != nil {
		return "", err
	}

	return token, keyring.Set(tg.serviceName, tg.username, token)
}

func getGitHubDeviceToken(httpClient *http.Client, logger *pterm.Logger, githubHostname, clientID string) (string, error) {
	scopes := []string{""}
	clientID = ensuresClientIDIfEmpty(clientID)
	code, err := device.RequestCode(
		httpClient,
		fmt.Sprintf("%s/login/device/code", githubHostname),
		clientID,
		scopes,
	)
	if err != nil {
		return "", err
	}

	logger.Trace("new code", logger.Args("code", code))
	logger.Warn("no valid token provided - grand access via pkup-gen GitHub app", logger.Args(
		"copy code", code.UserCode,
		"then open and paste the above code", code.VerificationURI,
	))

	accessToken, err := device.Wait(
		context.TODO(), httpClient,
		fmt.Sprintf("%s/login/oauth/access_token", githubHostname),
		device.WaitOptions{
			ClientID:   clientID,
			DeviceCode: code,
		})
	if err != nil {
		return "", err
	}

	return accessToken.Token, nil
}

func ensuresClientIDIfEmpty(clientID string) string {
	// clientID is empty when running app using `go run`
	// to support this scenario it's possible to get client from env
	if clientID != "" {
		return clientID
	}

	return os.Getenv("PKUP_GEN_CLIENT_ID")
}
