package github

import (
	"context"
	"net/http"
	"os"

	"github.com/cli/oauth/device"
	"github.com/google/go-github/v53/github"
	"github.com/pterm/pterm"
)

//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore
type Client interface {
	GetFileDiffForPRs([]*github.PullRequest, string, string) (string, error)
	ListUserPRsForRepo(Options, []FilterFunc) ([]*github.PullRequest, error)
	GetLatestReleaseOrZero(string, string) (string, error)
}

type gh_client struct {
	ctx    context.Context
	log    *pterm.Logger
	client *github.Client
}

type ClientOpts struct {
	Token         string
	EnterpriseURL string
	AppClientID   string
}

func NewClient(ctx context.Context, logger *pterm.Logger, opts ClientOpts) (Client, error) {
	logger.Trace("building GitHub client")
	if opts.Token == "" && opts.EnterpriseURL == "" {
		var err error
		logger.Trace("getting token from GitHub device")
		opts.Token, err = getGitHubDeviceToken(logger, opts.AppClientID)
		if err != nil {
			return nil, err
		}
	}

	client := github.NewTokenClient(ctx, opts.Token)

	if opts.EnterpriseURL != "" {
		logger.Trace("building enterprise client", logger.Args(
			"url", opts.EnterpriseURL,
		))
		enterpriseClient, err := github.NewEnterpriseClient(
			opts.EnterpriseURL,
			"",
			client.Client(),
		)

		if err != nil {
			return nil, err
		}

		client = enterpriseClient
	}

	return &gh_client{
		ctx:    ctx,
		log:    logger,
		client: client,
	}, nil
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

	logger.Warn("no token provided - grand access via pkup-gen GitHub app", logger.Args(
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
