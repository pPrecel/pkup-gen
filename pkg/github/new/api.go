package github

import (
	"context"
	"fmt"

	go_github "github.com/google/go-github/v53/github"
	"github.com/pterm/pterm"
)

type tokenUrl = string

var (
	// saves all clients in map to return existing one
	// format: "url@token"
	allClients = map[tokenUrl]*apiClient{}
)

//go:generate mockery --name=ApiClient --output=automock --outpkg=automock --case=underscore
type ApiClient interface {
	GetUserInfo(username string) (*UserInfo, error)
}

type apiClient struct {
	ctx      context.Context
	log      *pterm.Logger
	ghClient *go_github.Client
}

func NewApiClient(ctx context.Context, log *pterm.Logger, token, enterpriseURL string) (ApiClient, error) {
	tokenUrl := fmt.Sprintf("%s@%s", enterpriseURL, token)
	if client, ok := allClients[tokenUrl]; ok {
		return client, nil
	}

	client := go_github.NewTokenClient(ctx, token)

	if enterpriseURL != "" {
		log.Trace("building enterprise client", log.Args(
			"url", enterpriseURL,
		))

		enterpriseClient, err := go_github.NewEnterpriseClient(
			enterpriseURL,
			"",
			client.Client())
		if err != nil {
			return nil, err
		}

		client = enterpriseClient
	}

	allClients[tokenUrl] = &apiClient{
		ctx:      ctx,
		log:      log,
		ghClient: client,
	}
	return allClients[tokenUrl], nil
}

type UserInfo struct {
	Name  string
	Login string
	Email string
}

func (ac *apiClient) GetUserInfo(username string) (*UserInfo, error) {
	user, _, err := ac.ghClient.Users.Get(ac.ctx, username)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		Name:  user.GetName(),
		Login: user.GetLogin(),
		Email: user.GetEmail(),
	}, nil
}
