package token

import (
	"context"
	"net/http"
	"os"
	"os/user"

	"github.com/cli/oauth/device"
	"github.com/pterm/pterm"
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "pkup-gen"
)

func Get(logger *pterm.Logger, clientID string) (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	token, err := keyring.Get(serviceName, user.Username)
	if err == nil && isTokenValid(logger, token) {
		return token, nil
	}

	logger.Trace("getting token from GitHub device")
	token, err = getGitHubDeviceToken(logger, clientID)
	if err != nil {
		return "", err
	}

	return token, keyring.Set(serviceName, user.Username, token)
}

func getGitHubDeviceToken(logger *pterm.Logger, clientID string) (string, error) {
	scopes := []string{""}
	clientID = ensuresClientIDIfEmpty(clientID)
	httpClient := http.DefaultClient
	code, err := device.RequestCode(
		httpClient, "https://github.com/login/device/code",
		clientID, scopes)
	if err != nil {
		return "", err
	}

	logger.Warn("no valid token provided - grand access via pkup-gen GitHub app", logger.Args(
		"copy code", code.UserCode,
		"then open and paste the above code", code.VerificationURI,
	))

	accessToken, err := device.Wait(
		context.TODO(), httpClient,
		"https://github.com/login/oauth/access_token",
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
