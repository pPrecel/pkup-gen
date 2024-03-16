package github

import (
	"context"
	"testing"

	go_github "github.com/google/go-github/v53/github"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

var (
	testBranches = []*go_github.Branch{
		{
			Name: ptr.To[string]("main"),
		},
		{
			Name: ptr.To[string]("release-1"),
		},
		{
			Name: ptr.To[string]("release-2"),
		},
	}

	testBranchesSlice = []string{
		"main", "release-1", "release-2",
	}
)

func Test_gh_client_ListRepoBranches(t *testing.T) {
	t.Run("list branches", func(t *testing.T) {
		server := fixTestServer(t, &testServerArgs{
			branches: testBranches,
		})
		defer server.Close()

		gh := gh_client{
			ctx:    context.Background(),
			log:    fixLogger(),
			client: fixTestClient(t, server),
		}

		branchList, err := gh.ListRepoBranches("test-org", "test-repo")
		require.NoError(t, err)
		require.NotNil(t, branchList)
		require.ElementsMatch(t, testBranchesSlice, branchList.Branches)
	})
}
