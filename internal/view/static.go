package view

import (
	"fmt"
	"io"
	"os"

	"github.com/pPrecel/PKUP/pkg/github"
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

func (sv *staticView) Add(name string, valuesChan chan *github.CommitList, errorChan chan error) {
	sv.tasks[name] = taskChannels{
		valuesChan: valuesChan,
		errorChan:  errorChan,
	}
}

func (sv *staticView) Run() error {
	for name := range sv.tasks {
		sv.log.Info(fmt.Sprintf("Processing %s...", name))
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
	case commitList, ok := <-channels.valuesChan:
		if ok {
			if len(commitList.Commits) == 0 {
				log.Warn(
					fmt.Sprintf("skipping '%s' no user activity detected", taskName),
				)
			} else {
				text := fmt.Sprintf("found %d commits for '%s'", len(commitList.Commits), taskName)
				log.Info(text, log.Args("commits", commitsToStringList(commitList)))
			}
		}
	default:
		return false
	}
	return true
}
