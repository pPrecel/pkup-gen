package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	go_github "github.com/google/go-github/v53/github"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func Test_gh_client_ListRepos(t *testing.T) {
	t.Run("list repos", func(t *testing.T) {
		testRepos := []string{
			"test-repo-1",
			"test-repo-2",
			"test-repo-3",
		}

		server := fixTestServer(t, nil, fixTestRepos(testRepos...))
		defer server.Close()

		gh := gh_client{
			ctx:    context.Background(),
			log:    fixLogger(),
			client: fixTestClient(t, server),
		}

		repos, err := gh.ListRepos("test-org")

		require.NoError(t, err)
		require.Equal(t, testRepos, repos)
	})

	t.Run("client error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(404)
		}))
		defer server.Close()

		gh := gh_client{
			ctx:    context.Background(),
			log:    fixLogger(),
			client: fixTestClient(t, server),
		}

		repos, err := gh.ListRepos("test-org")

		require.Error(t, err)
		require.Empty(t, repos)
	})
}

func fixTestRepos(names ...string) []*go_github.Repository {
	repos := []*go_github.Repository{}
	for _, name := range names {
		repos = append(repos, &go_github.Repository{
			Name: ptr.To[string](name),
		})
	}

	return repos
}
