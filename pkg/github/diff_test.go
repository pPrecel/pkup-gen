package github

import (
	"context"
	"io"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v53/github"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/require"
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

		diff, err := gh.GetPRContentDiff(testPullRequests[0], "pPrecel", "pkup-gen")
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
