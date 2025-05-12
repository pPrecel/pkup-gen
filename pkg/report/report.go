package report

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pPrecel/PKUP/internal/file"
	"github.com/pPrecel/PKUP/pkg/github"
)

const (
	PeriodFormat = "02.01.2006"
)

type Result struct {
	Org  string
	Repo string
	// URL        string
	CommitList github.CommitList
}

type Options struct {
	OutputDir    string
	TemplatePath string
	PeriodFrom   time.Time
	PeriodTill   time.Time
	Results      []Result
	CustomValues map[string]string
}

func Render(opts Options) error {
	values := Values{
		PeriodFrom:   opts.PeriodFrom.Format(PeriodFormat),
		PeriodTill:   opts.PeriodTill.Format(PeriodFormat),
		ApprovalDate: opts.PeriodTill.Add(time.Hour * 24).Format(PeriodFormat),
		Result:       buildreportResult(opts),
		CustomValues: opts.CustomValues,
	}

	if opts.TemplatePath != "" {
		return newFromTemplate(opts.TemplatePath).RenderToFile(
			opts.OutputDir,
			filepath.Base(opts.TemplatePath),
			values,
		)
	}

	return newDefault().RenderToFile(
		opts.OutputDir,
		"report.txt",
		values,
	)
}

func buildreportResult(opts Options) []string {
	results := []string{}
	for _, result := range opts.Results {
		for i := range result.CommitList.Commits {
			org := result.Org
			repo := result.Repo
			commit := result.CommitList.Commits[i]
			results = append(
				results,
				fmt.Sprintf(
					"%s (%s)",
					strings.Split(commit.Commit.GetMessage(), "\n")[0],
					file.BuildDiffFilename(commit.GetSHA(), org, repo),
					// "<a href=\"%s/%s/%s/commit/%s\">%s</a> (%s)",
					// result.URL, org, repo, commit.GetSHA(), // commit link
					// strings.Split(commit.Commit.GetMessage(), "\n")[0], // commit message
					// file.BuildDiffFilename(commit.GetSHA(), org, repo), // file name
				),
			)
		}
	}
	return results
}
