package artifacts

import (
	"fmt"
	"time"

	gh "github.com/google/go-github/v53/github"
	"github.com/pPrecel/PKUP/internal/file"
	"github.com/pPrecel/PKUP/pkg/github"
)

type GenerateOpts struct {
	Org          string
	Repo         string
	Username     string
	Dir          string
	WithClosed   bool
	MergedAfter  time.Time
	MergedBefore time.Time
}

func GenUserArtifactsToFile(client github.Client, opts *GenerateOpts) ([]string, error) {
	prs, err := client.ListUserPRsForRepo(github.Options{
		Org:          opts.Org,
		Repo:         opts.Repo,
		Username:     opts.Username,
		WithClosed:   opts.WithClosed,
		MergedAfter:  opts.MergedAfter,
		MergedBefore: opts.MergedBefore,
	})
	if err != nil {
		return nil, fmt.Errorf("list users PRs in repo '%s/%s' error: %s", opts.Org, opts.Repo, err.Error())
	}

	diff, err := client.GetFileDiffForPRs(prs, opts.Org, opts.Repo)
	if err != nil {
		return nil, fmt.Errorf("get diff for repo '%s/%s' error: %s", opts.Org, opts.Repo, err.Error())
	}

	if diff != "" {
		filename := fmt.Sprintf("%s_%s.patch", opts.Org, opts.Repo)
		err = file.Create(opts.Dir, filename, diff)
		if err != nil {
			return nil, fmt.Errorf("save file '%s' error: %s", filename, err.Error())
		}
	}

	return prsToStringList(prs, opts.WithClosed), nil
}

func prsToStringList(prs []*gh.PullRequest, signState bool) []string {
	list := []string{}
	for i := range prs {
		pr := *prs[i]

		title := *pr.Title
		if signState {
			title = fmt.Sprintf("%s %s", getStatePrefix(pr), pr.GetTitle())
		}
		list = append(list, title)
	}

	return list
}

func getStatePrefix(pr gh.PullRequest) string {
	if !pr.GetMergedAt().IsZero() {
		return "[MERGED]"
	}

	return "[CLOSED]"
}
