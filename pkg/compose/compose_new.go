package compose

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pPrecel/PKUP/internal/view"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pterm/pterm"
)

type compose_new struct {
	ctx         context.Context
	logger      *pterm.Logger
	buildClient buildClientFunc
}

func NewNew(ctx context.Context, logger *pterm.Logger) Compose {
	return &compose_new{
		ctx:         ctx,
		logger:      logger,
		buildClient: github.NewClient,
	}
}

func (c *compose_new) ForConfig(config *Config, opts ComposeOpts) error {
	view := view.NewMultiTaskView(c.logger, opts.Ci)

	for i := range config.Users {
		user := config.Users[i]

		valChan := make(chan *github.CommitList)
		errChan := make(chan error)
		view.Add(user.Username, valChan, errChan)
		go func() {

		}()
	}

	viewWg := sync.WaitGroup{}
	viewWg.Add(1)
	var runErr error
	go func() {
		runErr = view.Run()
		viewWg.Done()
	}()

	viewWg.Wait()
	if runErr != nil {
		return runErr
	}

	return runErr
}

type OrgRepoInfo struct {
	Org        string
	Repo       string
	URL        string
	CommitIter object.CommitIter
	ApiClient  github.Client
}

func mock() {
	r, _ := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: "https://github.com/pPrecel/pkup-gen-reports",
		Auth: &http.BasicAuth{
			Username: "pkup-bot",
			Password: "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",
		},
		Progress: os.Stdout,
	})

	cIter, _ := r.Log(&git.LogOptions{})

	_ = cIter.ForEach(func(c *object.Commit) error {
		fmt.Println(c)
		return nil
	})
}
