package generator

import (
	"context"

	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pterm/pterm"
)

//go:generate mockery --name=Generator --output=automock --outpkg=automock --case=underscore
type Generator interface {
	ForConfig(*Config, ComposeOpts) error
	ForArgs(*GeneratorArgs) error
}

type buildClientFunc func(context.Context, *pterm.Logger, github.ClientOpts) (github.Client, error)

type generator struct {
	ctx         context.Context
	logger      *pterm.Logger
	buildClient buildClientFunc
}

func New(ctx context.Context, logger *pterm.Logger) Generator {
	return &generator{
		ctx:         ctx,
		logger:      logger,
		buildClient: github.NewClient,
	}
}
