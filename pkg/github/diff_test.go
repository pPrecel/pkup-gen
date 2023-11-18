package github

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-github/v53/github"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

const (
	diffMessage = "test diff"
)

func Test_gh_client_GetFileDiffForPRs(t *testing.T) {
	t.Run("get diff", func(t *testing.T) {
		testDiff := "test diff"
		server := fixTestServer(t, nil, nil)
		defer server.Close()

		gh := gh_client{
			ctx:    context.Background(),
			log:    fixLogger(),
			client: fixTestClient(t, server),
		}

		diff, err := gh.GetCommitContentDiff(&github.RepositoryCommit{
			SHA: ptr.To[string]("test-sha-1"),
		}, "pPrecel", "pkup-gen")
		require.NoError(t, err)
		require.Equal(t, testDiff, diff)
	})
}

func fixLogger() *pterm.Logger {
	log := &pterm.DefaultLogger
	log.Writer = io.Discard
	return log
}

func fixTestClient(t *testing.T, server *httptest.Server) *github.Client {
	client := github.NewClient(server.Client())
	baseURL, err := url.Parse(server.URL + "/")
	require.NoError(t, err)

	client.BaseURL = baseURL
	return client
}

func fixTestServer(t *testing.T, commits []*github.RepositoryCommit, repos []*github.Repository) *httptest.Server {
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

		// repos
		if strings.Contains(r.URL.String(), "/repos") {
			bytes, err := json.Marshal(&repos)
			require.NoError(t, err)
			w.Write(bytes)
			return
		}
	}))
}
