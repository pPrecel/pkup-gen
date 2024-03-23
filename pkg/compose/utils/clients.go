package utils

import (
	"context"
	"fmt"

	"github.com/pPrecel/PKUP/pkg/config"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pterm/pterm"
)

// use to get default GitHub client
const DefaultGitHubURL = ""

type RemoteClients map[string]github.Client

func (rc RemoteClients) Get(url string) github.Client {
	return rc[url]
}

func (rc RemoteClients) set(url string, client github.Client) {
	rc[url] = client
}

type BuildClientFunc func(context.Context, *pterm.Logger, github.ClientOpts) (github.Client, error)

func BuildClients(ctx context.Context, logger *pterm.Logger, config *config.Config, buildClient BuildClientFunc) (*RemoteClients, error) {
	remoteClients := &RemoteClients{}

	err := appendRemoteClients(remoteClients, ctx, logger, config.Orgs, buildClient)
	if err != nil {
		return nil, err
	}

	err = appendRemoteClients(remoteClients, ctx, logger, config.Repos, buildClient)

	return remoteClients, err
}

func appendRemoteClients(dest *RemoteClients, ctx context.Context, logger *pterm.Logger, remotes []config.Remote, buildClient BuildClientFunc) error {
	for i := range remotes {
		if c := dest.Get(remotes[i].EnterpriseUrl); c == nil {
			var err error
			client, err := buildClient(
				ctx,
				logger,
				github.ClientOpts{
					EnterpriseURL: remotes[i].EnterpriseUrl,
					Token:         remotes[i].Token,
				},
			)
			if err != nil {
				return fmt.Errorf("failed to build client for '%s': %s", remotes[i].Name, err.Error())
			}

			dest.set(remotes[i].EnterpriseUrl, client)
		}
	}

	return nil
}
