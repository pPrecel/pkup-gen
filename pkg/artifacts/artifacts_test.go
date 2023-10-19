package artifacts

import (
	"errors"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	gh "github.com/google/go-github/v53/github"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pPrecel/PKUP/pkg/github/automock"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/pointer"
)

func TestGenUserArtifactsToDir(t *testing.T) {
	t.Run("generate diff", func(t *testing.T) {
		tmpDir := t.TempDir()
		diff := "+ anything"
		testPRs := []*gh.PullRequest{
			{
				Title:          pointer.String("test PR 1"),
				MergeCommitSHA: pointer.String("sha1"),
				Number:         pointer.Int(123),
				ClosedAt:       &gh.Timestamp{},
			},
			{
				Title:          pointer.String("test PR 2"),
				MergeCommitSHA: pointer.String("sha2"),
				Number:         pointer.Int(124),
				MergedAt:       &gh.Timestamp{Time: time.Now()},
			},
		}

		clientMock := automock.NewClient(t)
		clientMock.On("GetPRContentDiff", mock.Anything, "test-org", "test-repo").Return(diff, nil).Twice()
		clientMock.On("ListUserPRsForRepo", github.Options{
			Org:          "test-org",
			Repo:         "test-repo",
			Username:     "test-username",
			MergedBefore: time.Time{},
			MergedAfter:  time.Time{},
		}, mock.Anything).Return(testPRs, nil).Once()

		prs, err := GenUserArtifactsToDir(clientMock, Options{
			Org:          "test-org",
			Repo:         "test-repo",
			Username:     "test-username",
			WithClosed:   true,
			MergedBefore: time.Time{},
			MergedAfter:  time.Time{},
			Dir:          tmpDir,
		})

		require.NoError(t, err)

		expectedPRs := []string{
			fmt.Sprint(" ", pterm.Red("[C]"), " (#123) test PR 1"),
			fmt.Sprint(" ", pterm.Magenta("[M]"), " (#124) test PR 2"),
		}
		require.ElementsMatch(t, expectedPRs, prs)

		expectedDiffFile := path.Join(tmpDir, "test-org_test-repo_sha1.diff")
		require.FileExists(t, expectedDiffFile)
		diffBody, err := os.ReadFile(expectedDiffFile)
		require.NoError(t, err)
		require.Equal(t, diff, string(diffBody))

		expectedDiffFile = path.Join(tmpDir, "test-org_test-repo_sha2.diff")
		require.FileExists(t, expectedDiffFile)
		diffBody, err = os.ReadFile(expectedDiffFile)
		require.NoError(t, err)
		require.Equal(t, diff, string(diffBody))
	})
	t.Run("list error", func(t *testing.T) {
		clientMock := automock.NewClient(t)
		clientMock.On("ListUserPRsForRepo", github.Options{
			Org:          "test-org",
			Repo:         "test-repo",
			Username:     "test-username",
			MergedBefore: time.Time{},
			MergedAfter:  time.Time{},
		}, mock.Anything).Return(nil, errors.New("test error")).Once()

		prs, err := GenUserArtifactsToDir(clientMock, Options{
			Org:          "test-org",
			Repo:         "test-repo",
			Username:     "test-username",
			WithClosed:   true,
			MergedBefore: time.Time{},
			MergedAfter:  time.Time{},
			Dir:          "/test/dir",
		})

		require.Error(t, err)
		require.Empty(t, prs)
	})
	t.Run("diff error", func(t *testing.T) {
		clientMock := automock.NewClient(t)
		clientMock.On("GetPRContentDiff", mock.Anything, "test-org", "test-repo").Return("", errors.New("test error")).Once()
		clientMock.On("ListUserPRsForRepo", github.Options{
			Org:          "test-org",
			Repo:         "test-repo",
			Username:     "test-username",
			MergedBefore: time.Time{},
			MergedAfter:  time.Time{},
		}, mock.Anything).Return([]*gh.PullRequest{
			{
				//empty PR
			},
		}, nil).Once()

		prs, err := GenUserArtifactsToDir(clientMock, Options{
			Org:          "test-org",
			Repo:         "test-repo",
			Username:     "test-username",
			WithClosed:   true,
			MergedBefore: time.Time{},
			MergedAfter:  time.Time{},
			Dir:          "/test/dir",
		})

		require.Error(t, err)
		require.Empty(t, prs)
	})
}
