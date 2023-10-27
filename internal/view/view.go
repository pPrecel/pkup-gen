package view

import (
	"io"

	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pterm/pterm"
)

type MultiTaskView interface {
	Run() error
	Add(string, chan *github.CommitList, chan error)
	NewWriter() io.Writer
}

type taskChannels struct {
	valuesChan chan *github.CommitList
	errorChan  chan error
}

func NewMultiTaskView(log *pterm.Logger, useStatic bool) MultiTaskView {
	if useStatic {
		return newStatic(log)
	}
	return newDynamic(log)
}
