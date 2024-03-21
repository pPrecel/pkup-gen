package generator

import (
	"context"
	"fmt"

	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pterm/pterm"
)

type remoteClients map[string]github.Client

func (rc remoteClients) Set(url string, client github.Client) {
	rc[url] = client
}

func (rc remoteClients) Get(url string) github.Client {
	return rc[url]
}

func buildClients(ctx context.Context, logger *pterm.Logger, config *Config, buildClient buildClientFunc) (*remoteClients, error) {
	remoteClients := &remoteClients{}

	err := appendRemoteClients(remoteClients, ctx, logger, config.Orgs, buildClient)
	if err != nil {
		return nil, err
	}

	err = appendRemoteClients(remoteClients, ctx, logger, config.Repos, buildClient)

	return remoteClients, err
}

func appendRemoteClients(dest *remoteClients, ctx context.Context, logger *pterm.Logger, remotes []Remote, buildClient buildClientFunc) error {
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

			dest.Set(remotes[i].EnterpriseUrl, client)
		}
	}

	return nil
}
