package view

import (
	"io"

	"github.com/pterm/pterm"
)

type MultiTaskView interface {
	Run() error
	Add(string, chan []*RepoCommit, chan error)
	NewWriter() io.Writer
}

type RepoCommit struct {
	Org     string
	Repo    string
	Message string
	SHA     string
}

type taskChannels struct {
	valuesChan chan []*RepoCommit
	errorChan  chan error
}

func NewMultiTaskView(log *pterm.Logger, useStatic bool) MultiTaskView {
	if useStatic {
		return newStatic(log)
	}
	return newDynamic(log)
}
