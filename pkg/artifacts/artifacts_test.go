package artifacts

import (
	"errors"
	"os"
	"path"
	"testing"
	"time"

	go_github "github.com/google/go-github/v53/github"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pPrecel/PKUP/pkg/github/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestGenUserArtifactsToDir(t *testing.T) {
	t.Run("generate diff", func(t *testing.T) {
		tmpDir := t.TempDir()
		diff := "+ anything"
		testCommits := &github.CommitList{
			Commits: []*go_github.RepositoryCommit{
				{
					SHA: ptr.To("sha1"),
					Commit: &go_github.Commit{
						Message: ptr.To("test PR 1 (#123)"),
					},
				},
				{
					SHA: ptr.To("sha2"),
					Commit: &go_github.Commit{
						Message: ptr.To("test PR 2 (#123)"),
					},
				},
			},
		}

		clientMock := automock.NewClient(t)
		clientMock.On("GetCommitContentDiff", mock.Anything, "test-org", "test-repo").Return(diff, nil).Twice()
		clientMock.On("ListRepoCommits", github.ListRepoCommitsOpts{
			Org:     "test-org",
			Repo:    "test-repo",
			Authors: []string{"test-username"},
			Since:   time.Time{},
			Until:   time.Time{},
		}, mock.Anything).Return(testCommits, nil).Once()

		commitList, err := GenUserArtifactsToDir(clientMock, Options{
			Org:     "test-org",
			Repo:    "test-repo",
			Authors: []string{"test-username"},
			Since:   time.Time{},
			Until:   time.Time{},
			Dir:     tmpDir,
		})

		require.NoError(t, err)
		require.ElementsMatch(t, testCommits.Commits, commitList.Commits)

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
		clientMock.On("ListRepoCommits", github.ListRepoCommitsOpts{
			Org:     "test-org",
			Repo:    "test-repo",
			Authors: []string{"test-username"},
			Since:   time.Time{},
			Until:   time.Time{},
		}, mock.Anything).Return(nil, errors.New("test error")).Once()

		prs, err := GenUserArtifactsToDir(clientMock, Options{
			Org:     "test-org",
			Repo:    "test-repo",
			Authors: []string{"test-username"},
			Since:   time.Time{},
			Until:   time.Time{},
			Dir:     "/test/dir",
		})

		require.Error(t, err)
		require.Empty(t, prs)
	})
	t.Run("diff error", func(t *testing.T) {
		clientMock := automock.NewClient(t)
		clientMock.On("GetCommitContentDiff", mock.Anything, "test-org", "test-repo").Return("", errors.New("test error")).Once()
		clientMock.On("ListRepoCommits", github.ListRepoCommitsOpts{
			Org:     "test-org",
			Repo:    "test-repo",
			Authors: []string{"test-username"},
			Since:   time.Time{},
			Until:   time.Time{},
		}, mock.Anything).Return(&github.CommitList{
			Commits: []*go_github.RepositoryCommit{
				{
					// empty commit
				},
			},
		}, nil).Once()

		prs, err := GenUserArtifactsToDir(clientMock, Options{
			Org:     "test-org",
			Repo:    "test-repo",
			Authors: []string{"test-username"},
			Since:   time.Time{},
			Until:   time.Time{},
			Dir:     "/test/dir",
		})

		require.Error(t, err)
		require.Empty(t, prs)
	})
}
