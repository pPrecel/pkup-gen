package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/pointer"
)

const (
	diffMessage = "test diff"
)

var (
	testMergedBefore      = time.Now()
	testMergedAfter       = time.Now().AddDate(0, -1, 0)
	testWrongPullRequests = []*github.PullRequest{
		{
			// wrong period
			MergedAt: &github.Timestamp{
				Time: testMergedBefore.AddDate(-1, 0, -2),
			},
		},
	}
	testPullRequests = []*github.PullRequest{
		{
			MergeCommitSHA: pointer.String("9c3af6972080f7b8e20948bfd61f755cf3e74ab8"),
			MergedAt: &github.Timestamp{
				Time: testMergedBefore.AddDate(0, 0, -1),
			},
		},
		{
			MergeCommitSHA: pointer.String("9c3af6972080f7b8e20948bfd61f755cf3e74ab8"),
			MergedAt: &github.Timestamp{
				Time: testMergedBefore.AddDate(0, 0, -3),
			},
		},
	}
	testCommits = []*github.RepositoryCommit{
		{
			Author: &github.User{
				Login: pointer.String("anyone"),
			},
			Committer: &github.User{
				Login: pointer.String("anyone"),
			},
		},
		{
			Author: &github.User{
				Login: pointer.String("pPrecel"),
			},
			Committer: &github.User{
				Login: pointer.String("pPrecel"),
			},
		},
	}
)

func Test_gh_client_ListUserPRsForRepo(t *testing.T) {
	t.Run("list user PRs", func(t *testing.T) {
		server := fixTestServer(t,
			append(testPullRequests, testWrongPullRequests...), testCommits)
		defer server.Close()

		gh := gh_client{
			ctx:    context.Background(),
			log:    fixLogger(),
			client: fixTestClient(t, server),
		}

		PRs, err := gh.ListUserPRsForRepo(Options{
			Org:          "pPrecel",
			Repo:         "pkup-gen",
			Username:     "pPrecel",
			MergedBefore: testMergedBefore,
			MergedAfter:  testMergedAfter,
		}, []FilterFunc{FilterPRsByMergedAt})
		require.NoError(t, err)
		require.Len(t, PRs, 2)
	})
}

func fixTestServer(t *testing.T, PRs []*github.PullRequest, commits []*github.RepositoryCommit) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// diff
		if strings.Contains(r.URL.String(), "/commits/") {
			w.Write([]byte(diffMessage))
			return
		}

		// commits
		if strings.Contains(r.URL.String(), "/commits") {
			bytes, err := json.Marshal(&commits)
			require.NoError(t, err)
			w.Write(bytes)
			return
		}

		// pull requests
		bytes, err := json.Marshal(&PRs)
		require.NoError(t, err)
		w.Write(bytes)
	}))
}
