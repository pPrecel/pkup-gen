package artifacts

import (
	"fmt"
	"time"

	"github.com/pPrecel/PKUP/internal/file"
	"github.com/pPrecel/PKUP/pkg/github"
)

type Options struct {
	Org     string
	Repo    string
	Authors []string
	Dir     string
	Since   time.Time
	Until   time.Time
}

func GenUserArtifactsToDir(client github.Client, opts Options) (*github.CommitList, error) {
	commits, err := client.ListRepoCommits(github.ListRepoCommitsOpts{
		Org:     opts.Org,
		Repo:    opts.Repo,
		Authors: opts.Authors,
		Since:   opts.Since,
		Until:   opts.Until,
	})
	if err != nil {
		return nil, fmt.Errorf("list users commits in repo '%s/%s' error: %s", opts.Org, opts.Repo, err.Error())
	}

	err = SaveDiffToFiles(client, commits, opts)
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func SaveDiffToFiles(client github.Client, commits *github.CommitList, opts Options) error {
	for i := range commits.Commits {
		commit := commits.Commits[i]
		diff, err := client.GetCommitContentDiff(commit, opts.Org, opts.Repo)
		if err != nil {
			return fmt.Errorf("get diff for repo '%s/%s' error: %s", opts.Org, opts.Repo, err.Error())
		}

		if diff != "" {
			filename := file.BuildDiffFilename(commit.GetSHA(), opts.Org, opts.Repo)
			err = file.Create(opts.Dir, filename, diff)
			if err != nil {
				return fmt.Errorf("save file '%s' error: %s", filename, err.Error())
			}
		}
	}

	return nil
}
