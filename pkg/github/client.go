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

func NewClient(ctx context.Context, logger *logrus.Logger, token string) *gh_client {
	return &gh_client{
		ctx:    ctx,
		log:    logger,
		client: github.NewTokenClient(ctx, token),
	}
}
