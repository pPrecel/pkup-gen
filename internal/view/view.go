package view

import (
	"io"

	gh "github.com/google/go-github/v53/github"
	"github.com/pterm/pterm"
)

type MultiTaskView interface {
	Run() error
	Add(string, chan []*gh.PullRequest, chan error)
	NewWriter() io.Writer
}

type taskChannels struct {
	valuesChan chan []*gh.PullRequest
	errorChan  chan error
}

func NewMultiTaskView(log *pterm.Logger, useStatic bool) MultiTaskView {
	if useStatic {
		return newStatic(log)
	}
	return newDynamic(log)
}
