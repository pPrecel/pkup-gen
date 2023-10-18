package github

import (
	"context"

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
