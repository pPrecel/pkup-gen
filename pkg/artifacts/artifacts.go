package artifacts

import (
	"fmt"
	"time"

	gh "github.com/google/go-github/v53/github"
	"github.com/pPrecel/PKUP/internal/file"
	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pterm/pterm"
)

type Options struct {
	Org          string
	Repo         string
	Username     string
	Dir          string
	WithClosed   bool
	MergedAfter  time.Time
	MergedBefore time.Time
}

func GenUserArtifactsToDir(client github.Client, opts Options) ([]string, error) {
	filters := []github.FilterFunc{github.FilterPRsByMergedAt}
	if opts.WithClosed {
		filters = append(filters, github.FilterPRsByClosedAt)
	}
	prs, err := client.ListUserPRsForRepo(github.Options{
		Org:          opts.Org,
		Repo:         opts.Repo,
		Username:     opts.Username,
		MergedAfter:  opts.MergedAfter,
		MergedBefore: opts.MergedBefore,
	}, filters)
	if err != nil {
		return nil, fmt.Errorf("list users PRs in repo '%s/%s' error: %s", opts.Org, opts.Repo, err.Error())
	}

	err = savePRsDiffToFiles(client, prs, opts)
	if err != nil {
		return nil, err
	}

	return prsToStringList(prs), nil
}

func savePRsDiffToFiles(client github.Client, prs []*gh.PullRequest, opts Options) error {
	for i := range prs {
		pr := prs[i]
		diff, err := client.GetPRContentDiff(pr, opts.Org, opts.Repo)
		if err != nil {
			return fmt.Errorf("get diff for repo '%s/%s' error: %s", opts.Org, opts.Repo, err.Error())
		}

		if diff != "" {
			filename := fmt.Sprintf("%s_%s_%s.diff", opts.Org, opts.Repo, cutSHA(pr.GetMergeCommitSHA()))
			err = file.Create(opts.Dir, filename, diff)
			if err != nil {
				return fmt.Errorf("save file '%s' error: %s", filename, err.Error())
			}
		}
	}

	return nil
}

func cutSHA(fullSHA string) string {
	if len(fullSHA) < 8 {
		return fullSHA
	}

	return fullSHA[0:8]
}

func prsToStringList(prs []*gh.PullRequest) []string {
	list := []string{}
	for i := range prs {
		pr := *prs[i]

		title := fmt.Sprintf(" %s (#%d) %s", getStatePrefix(pr), pr.GetNumber(), pr.GetTitle())
		list = append(list, title)
	}

	return list
}

func getStatePrefix(pr gh.PullRequest) string {
	// pull request - 
	if !pr.GetMergedAt().IsZero() {
		return pterm.Magenta("[M]")
	}

	return pterm.Red("[C]")
}
