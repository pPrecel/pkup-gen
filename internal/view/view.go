package view

import (
	"io"

	"github.com/pterm/pterm"
)

type MultiTaskView interface {
	Run() error
	Add(string, chan []string, chan error)
	NewWriter() io.Writer
}

type taskChannels struct {
	valuesChan chan []string
	errorChan  chan error
}

func NewMultiTaskView(log *pterm.Logger, useStatic bool) MultiTaskView {
	if useStatic {
		return newStatic(log)
	}
	return newDynamic(log)
}
