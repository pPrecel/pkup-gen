package view

import (
	"fmt"
	"io"
	"os"

	"github.com/pterm/pterm"
)

type staticView struct {
	log   *pterm.Logger
	tasks map[string]taskChannels
}

func (sv *staticView) NewWriter() io.Writer {
	return os.Stdout
}

func newStatic(log *pterm.Logger) MultiTaskView {
	return &staticView{
		log:   log,
		tasks: map[string]taskChannels{},
	}
}

func (sv *staticView) Add(name string, valuesChan chan []string, errorChan chan error) {
	sv.tasks[name] = taskChannels{
		valuesChan: valuesChan,
		errorChan:  errorChan,
	}
}

func (sv *staticView) Run() error {
	for name := range sv.tasks {
		sv.log.Info(fmt.Sprintf("Processing '%s'...", name))
	}

	for len(sv.tasks) > 0 {
		for name, channels := range sv.tasks {
			if selectChannelsForLogger(sv.log, name, channels) {
				delete(sv.tasks, name)
			}
		}
	}

	return nil
}

func selectChannelsForLogger(log *pterm.Logger, taskName string, channels taskChannels) bool {
	select {
	case err, ok := <-channels.errorChan:
		if ok {
			log.Error(err.Error())
		}
	case PRs, ok := <-channels.valuesChan:
		if ok {
			if len(PRs) == 0 {
				log.Warn(
					fmt.Sprintf("skipping '%s' no user activity detected", taskName),
				)
			} else {
				text := fmt.Sprintf("found %d PRs for repo '%s'", len(PRs), taskName)
				log.Info(text, log.Args("prs", PRs))
			}
		}
	default:
		return false
	}
	return true
}
