package github

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pterm/pterm"
)

//go:generate mockery --name=GitRepo --output=automock --outpkg=automock --case=underscore
type RepoAgent interface {
}

type repoAgent struct {
	org   string
	repo  string
	url   string
	token string

	gitRepo   GitRepo
	apiClient ApiClient

	// lock does not allow to use any func before init ends
	initLock sync.WaitGroup
	// err that may be returned during init process
	initErr error
}

func NewRepoAgent(ctx context.Context, log *pterm.Logger, org, repo, url, token string) RepoAgent {
	apiUrl := url
	if apiUrl == "" {
		apiUrl = "https://github.com"
	}

	client := &repoAgent{
		org:   org,
		repo:  repo,
		url:   apiUrl,
		token: token,
	}

	client.initLock.Add(1)
	go func() {
		// lock until init process ends
		defer client.initLock.Done()
		var err error

		client.apiClient, err = NewApiClient(ctx, log, token, url)
		if err != nil {
			client.initErr = err
			return
		}

		client.gitRepo, err = newInMemoryRepo(ctx, log, org, repo, url, token)
		if err != nil {
			client.initErr = err
			return
		}
	}()

	return client
}

type UserCommit struct {
	AuthorLogin string
	AuthorName  string
	AuthorEmail string
	Hash        string
	Message     string
}

func (rc *repoAgent) ListUserCommits(username string, since, until *time.Time) ([]*UserCommit, error) {
	rc.initLock.Wait()
	if rc.initErr != nil {
		return nil, rc.initErr
	}

	userInfo, err := rc.apiClient.GetUserInfo(username)
	if err != nil {
		return nil, err
	}

	userCommits := []*UserCommit{}
	err = rc.gitRepo.ForEachCommit(since, until, func(c *commit) error {
		if isUserAuthor(c, userInfo) {

			fIter, err := c.Patch(c.Pa)
			if err != nil {
				return err
			}
			fIter.ForEach(func(f *object.File) error {

				return nil
			})

			userCommits = append(userCommits, &UserCommit{
				AuthorLogin: userInfo.Login,
				AuthorName:  userInfo.Name,
				AuthorEmail: userInfo.Email,
				Hash:        c.Hash.String(),
				Message:     strings.Split(c.Message, "\n")[0],
			})
		}

		return nil
	})

	return userCommits, err
}

func isUserAuthor(commit *commit, user *UserInfo) bool {
	// check if user is an author
	if commit.Author.Email == user.Email ||
		commit.Author.Name == user.Login ||
		commit.Author.Name == user.Name {
		return true
	}

	// check if user is an commiter
	if commit.Committer.Email == user.Email ||
		commit.Committer.Name == user.Login ||
		commit.Committer.Name == user.Name {
		return true
	}

	// check is user is an co-author
	payloadLines := strings.Split(commit.Message, "\n")
	for i := range payloadLines {
		line := payloadLines[i]
		// check if user is author of the commit based on the message fields
		// example message:
		// Fix links found by Link Checker (#415)
		// Co-authored-by: Natalia Sitko <80401180+nataliasitko@users.noreply.github.com>
		if strings.HasPrefix(line, fmt.Sprintf("Co-authored-by: %s ", user.Name)) ||
			strings.HasPrefix(line, fmt.Sprintf("Co-authored-by: %s ", user.Login)) {
			return true
		}
	}

	return false
}
