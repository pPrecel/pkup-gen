package generator

import (
	"context"
	"fmt"

	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pterm/pterm"
)

func buildClients(ctx context.Context, logger *pterm.Logger, config *Config, buildClient buildClientFunc) (map[string]github.Client, error) {
	urlClients := map[string]github.Client{}

	var err error
	urlClients, err = appendRemoteClients(urlClients, ctx, logger, config.Orgs, buildClient)
	if err != nil {
		return nil, err
	}

	urlClients, err = appendRemoteClients(urlClients, ctx, logger, config.Repos, buildClient)

	return urlClients, err
}

func appendRemoteClients(dest map[string]github.Client, ctx context.Context, logger *pterm.Logger, remotes []Remote, buildClient buildClientFunc) (map[string]github.Client, error) {
	for i := range remotes {
		if _, ok := dest[remotes[i].EnterpriseUrl]; !ok {
			var err error
			dest[remotes[i].EnterpriseUrl], err = buildClient(
				ctx,
				logger,
				github.ClientOpts{
					EnterpriseURL: remotes[i].EnterpriseUrl,
					Token:         remotes[i].Token,
				},
			)
			if err != nil {
				return nil, fmt.Errorf("failed to build client for '%s': %s", remotes[i].Name, err.Error())
			}
		}
	}

	return dest, nil
}
