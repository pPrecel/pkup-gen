package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	go_github "github.com/google/go-github/v53/github"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

var (
	testWrongCommits = []*go_github.RepositoryCommit{
		{
			SHA: ptr.To[string]("test-sha1"),
			Author: &go_github.User{
				Login: ptr.To[string]("test-wronglogin"),
			},
		},
		{
			SHA: ptr.To[string]("test-sha2"),
			Author: &go_github.User{
				Name: ptr.To[string]("test-wrong-name"),
			},
		},
		{
			SHA: ptr.To[string]("test-sha3"),
			Commit: &go_github.Commit{
				Author: &go_github.CommitAuthor{
					Login: ptr.To[string]("test-wrong-login"),
				},
			},
		},
		{
			SHA: ptr.To[string]("test-sha4"),
			Commit: &go_github.Commit{
				Author: &go_github.CommitAuthor{
					Name: ptr.To[string]("test-wrong-name"),
				},
			},
		},
		{
			Commit: &go_github.Commit{
				Verification: &go_github.SignatureVerification{
					Verified: ptr.To[bool](true),
					Payload:  ptr.To[string]("\n\n\nauthor test-name-wrong <email>\n\n\n"),
				},
			},
		},
		{
			Commit: &go_github.Commit{
				Verification: &go_github.SignatureVerification{
					Verified: ptr.To[bool](true),
					Payload:  ptr.To[string]("\n\nCo-authored-by: test-name-wrong <email>\n"),
				},
			},
		},
	}
	testVerifiedCommit = []*go_github.RepositoryCommit{
		{
			Commit: &go_github.Commit{
				Verification: &go_github.SignatureVerification{
					Verified: ptr.To[bool](true),
					Payload:  ptr.To[string]("\n\n\nauthor test-name <email>\n\n\n"),
				},
			},
		},
		{
			Commit: &go_github.Commit{
				Verification: &go_github.SignatureVerification{
					Verified: ptr.To[bool](true),
					Payload:  ptr.To[string]("\n\nCo-authored-by: test-name <email>\n"),
				},
			},
		},
	}
	testCommits = []*go_github.RepositoryCommit{
		{
			SHA: ptr.To[string]("test-sha1"),
			Author: &go_github.User{
				Login: ptr.To[string]("test-login"),
			},
		},
		{
			SHA: ptr.To[string]("test-sha2"),
			Author: &go_github.User{
				Name: ptr.To[string]("test-name"),
			},
		},
		{
			SHA: ptr.To[string]("test-sha3"),
			Commit: &go_github.Commit{
				Author: &go_github.CommitAuthor{
					Login: ptr.To[string]("test-login"),
				},
			},
		},
		{
			SHA: ptr.To[string]("test-sha4"),
			Commit: &go_github.Commit{
				Author: &go_github.CommitAuthor{
					Name: ptr.To[string]("test-name"),
				},
			},
		},
	}
)

func Test_gh_client_ListRepoCommits(t *testing.T) {
	t.Run("list user commits", func(t *testing.T) {
		server := fixTestServer(t, &testServerArgs{
			commits: append(
				append(testCommits, testWrongCommits...),
				testVerifiedCommit...,
			),
		})
		defer server.Close()

		gh := gh_client{
			ctx:    context.Background(),
			log:    fixLogger(),
			client: fixTestClient(t, server),
		}

		commitList, err := gh.ListRepoCommits(ListRepoCommitsOpts{
			Org:      "test-org",
			Repo:     "test-repo",
			Branches: []string{"main"},
			Authors:  []string{"test-name", "test-login"},
			Since:    time.Time{},
			Until:    time.Time{},
		})

		require.NoError(t, err)
		require.NotNil(t, commitList)
		require.ElementsMatch(t, append(testCommits, testVerifiedCommit...), commitList.Commits)
	})

	t.Run("wait until rate limits end", func(t *testing.T) {
		callsIter := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if callsIter > 3 {
				handleTestRequest(t, &testServerArgs{
					commits: append(
						append(testCommits, testWrongCommits...),
						testVerifiedCommit...,
					),
				})(w, r)

				return
			}
			w.Header().Set("X-RateLimit-Remaining", "0") //fmt.Sprintf("%d", time.Now().Unix()))
			w.Header().Set("X-RateLimit-Reset", "22")
			w.WriteHeader(403)
			w.Write([]byte(`{"message":"API rate limit exceeded"}`))
			callsIter++
		}))
		defer server.Close()

		gh := gh_client{
			ctx:    context.Background(),
			log:    fixLogger(),
			client: fixTestClient(t, server),
		}

		commitList, err := gh.ListRepoCommits(ListRepoCommitsOpts{
			Org:      "test-org",
			Repo:     "test-repo",
			Branches: []string{"main"},
			Authors:  []string{"test-name", "test-login"},
			Since:    time.Time{},
			Until:    time.Time{},
		})

		require.NoError(t, err)
		require.NotNil(t, commitList)
		require.ElementsMatch(t, append(testCommits, testVerifiedCommit...), commitList.Commits)
	})
}
