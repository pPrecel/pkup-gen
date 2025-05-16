package github

import (
	"context"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/pterm/pterm"
)

//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore
type Client interface {
	ListRepoCommits(ListRepoCommitsOpts) (*CommitList, error)
	ListRepos(string) ([]string, error)
	ListRepoBranches(string, string) (*BranchList, error)
	GetCommitContentDiff(*github.RepositoryCommit, string, string) (string, error)
	GetLatestReleaseOrZero(string, string) (string, error)
	GetUserSignatures(string) ([]string, error)
}

type gh_client struct {
	ctx    context.Context
	log    *pterm.Logger
	client *github.Client
}

type ClientOpts struct {
	Token         string
	EnterpriseURL string
}

func NewClient(ctx context.Context, logger *pterm.Logger, opts ClientOpts) (Client, error) {
	client := github.NewTokenClient(ctx, opts.Token)

	if opts.EnterpriseURL != "" {
		logger.Trace("building enterprise client", logger.Args(
			"url", opts.EnterpriseURL,
		))
		enterpriseClient, err := github.NewEnterpriseClient(
			opts.EnterpriseURL,
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

func retryOnRateLimit[T any](log *pterm.Logger, fn func() (T, *github.Response, error)) (T, *github.Response, error) {
	var value T
	var resp *github.Response
	var err error

	for i := 0; i < 5; i++ {
		log.Trace("Ralling GH API", log.Args("iteration", i))
		value, resp, err = fn()
		if isRateLimitErr(err) && i < 5 {
			// rate limit reached
			// wait for the reset time
			// https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api?apiVersion=2022-11-28
			d := getRateLimitResetDuration(err)
			log.Warn("Rate limit exceeded, waiting", log.Args("duration", d, "error", err.Error()))
			time.Sleep(d)
			continue
		}

		break
	}

	return value, resp, err
}

func isRateLimitErr(err error) bool {
	if err == nil {
		return false
	}

	switch e := err.(type) {
	case *github.ErrorResponse:
		// common error response when reaching the same endpoint too many times
		return e.Response.StatusCode == 403
	case *github.RateLimitError:
		// specific error response when reaching the rate limit
		return true
	default:
		return false
	}
}

func getRateLimitResetDuration(err error) time.Duration {
	switch e := err.(type) {
	case *github.RateLimitError:
		return time.Until(e.Rate.Reset.Time)
	default:
		// default to 1 minute if we can't determine the reset time
		return time.Minute
	}
}
