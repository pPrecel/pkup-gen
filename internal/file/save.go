package file

import (
	"fmt"
	"os"

	gh "github.com/google/go-github/v53/github"
)

func BuildDiffFilename(pr *gh.PullRequest, org, repo string) string {
	return fmt.Sprintf("%s_%s_%s.diff", org, repo, cutSHA(pr.GetMergeCommitSHA()))
}

func Create(dir, filename, content string) error {
	file, err := os.Create(fmt.Sprintf("%s/%s", dir, filename))
	if err != nil {
		return err
	}

	_, err = file.WriteString(content)
	return err
}

func cutSHA(fullSHA string) string {
	if len(fullSHA) < 8 {
		return fullSHA
	}

	return fullSHA[0:8]
}
