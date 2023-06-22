package github

import (
	"context"

	"github.com/google/go-github/v53/github"
	"github.com/sirupsen/logrus"
)

type gh_client struct {
	ctx    context.Context
	log    *logrus.Logger
	client *github.Client
}

func NewClient(ctx context.Context, logger *logrus.Logger, token, enterpriseURL string) (*gh_client, error) {
	client := github.NewTokenClient(ctx, token)
	if enterpriseURL != "" {
		enterpriseClient, err := github.NewEnterpriseClient(
			enterpriseURL,
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
