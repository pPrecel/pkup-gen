package github

import (
	"context"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pterm/pterm"
)

//go:generate mockery --name=GitRepo --output=automock --outpkg=automock --case=underscore
type GitRepo interface {
	ForEachCommit(*time.Time, *time.Time, func(*commit) error) error
}

type inMemoryRepo struct {
	ctx        context.Context
	log        *pterm.Logger
	repository *git.Repository
}

func newInMemoryRepo(ctx context.Context, log *pterm.Logger, org, repo, url, token string) (GitRepo, error) {
	repoUrl := fmt.Sprintf("%s/%s/%s", url, org, repo)
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: repoUrl,
		Auth: &http.BasicAuth{
			Username: "pkup-bot",
			Password: token,
		},
	})
	if err != nil {
		return nil, err
	}

	return &inMemoryRepo{
		ctx:        ctx,
		log:        log,
		repository: r,
	}, nil
}

type commit = object.Commit

func (r *inMemoryRepo) ForEachCommit(since, until *time.Time, iterFn func(*commit) error) error {

	cIter, err := r.repository.Log(&git.LogOptions{
		All:   true,
		Since: since,
		Until: until,
	})
	if err != nil {
		return err
	}

	return cIter.ForEach(iterFn)
}
