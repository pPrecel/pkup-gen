package generator

import (
	"context"

	"github.com/pPrecel/PKUP/pkg/generator/config"
	"github.com/pPrecel/PKUP/pkg/generator/utils"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pterm/pterm"
)

//go:generate mockery --name=Generator --output=automock --outpkg=automock --case=underscore
type Generator interface {
	ForConfig(*config.Config, ComposeOpts) error
	ForArgs(*GeneratorArgs) error
}

type generator struct {
	ctx         context.Context
	logger      *pterm.Logger
	buildClient utils.BuildClientFunc

	repoCommitsLister utils.LazyCommitsLister
}

func New(ctx context.Context, logger *pterm.Logger) Generator {
	return &generator{
		ctx:         ctx,
		logger:      logger,
		buildClient: github.NewClient,
	}
}
