package report

import (
	"fmt"
	"path/filepath"
	"time"

	gh "github.com/google/go-github/v53/github"
	"github.com/pPrecel/PKUP/internal/file"
)

const (
	periodFormat = "02.01.2006"
)

type Result struct {
	Org          string
	Repo         string
	PullRequests []*gh.PullRequest
}

type Options struct {
	OutputDir    string
	TemplatePath string
	PeriodFrom   time.Time
	PeriodTill   time.Time
	Results      []Result
}

func Render(opts Options) error {
	values := Values{
		PeriodFrom:   opts.PeriodFrom.Format(periodFormat),
		PeriodTill:   opts.PeriodTill.Format(periodFormat),
		ApprovalDate: opts.PeriodTill.Add(time.Hour * 24).Format(periodFormat),
		Result:       buildreportResult(opts),
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
		for i := range result.PullRequests {
			org := result.Org
			repo := result.Repo
			pr := result.PullRequests[i]
			results = append(
				results,
				fmt.Sprintf("%s (%s)", pr.GetTitle(), file.BuildDiffFilename(pr, org, repo)),
			)
		}
	}
	return results
}
