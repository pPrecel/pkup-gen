package view

import (
	"fmt"
	"io"
	"strings"

	"github.com/pPrecel/PKUP/pkg/github"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
)

func init() {
	pterm.Success.Prefix = pterm.Prefix{
		Text:  "✓",
		Style: pterm.NewStyle(pterm.FgGreen),
	}
	pterm.Error.Prefix = pterm.Prefix{
		Text:  "✗",
		Style: pterm.NewStyle(pterm.FgRed),
	}
	pterm.Warning.Prefix = pterm.Prefix{
		Text:  "✗",
		Style: pterm.NewStyle(pterm.FgYellow),
	}
}

type dynamicMultiView struct {
	log          *pterm.Logger
	multiPrinter pterm.MultiPrinter
	tasks        map[string]taskChannels
}

func newDynamic(log *pterm.Logger) MultiTaskView {
	return &dynamicMultiView{
		log:          log,
		multiPrinter: pterm.DefaultMultiPrinter,
		tasks:        map[string]taskChannels{},
	}
}

func (mtv *dynamicMultiView) NewWriter() io.Writer {
	return mtv.multiPrinter.NewWriter()
}

func (mtv *dynamicMultiView) Add(name string, valuesChan chan *github.CommitList, errorChan chan error) {
	mtv.tasks[name] = taskChannels{
		valuesChan: valuesChan,
		errorChan:  errorChan,
	}
}

func (mtv *dynamicMultiView) Run() error {
	workingSpinners, err := startSpinnersWithPrinter(mtv.tasks, &mtv.multiPrinter)
	if err != nil {
		return err
	}

	mtv.multiPrinter.Start()
	for len(workingSpinners) > 0 {
		for name, channels := range mtv.tasks {
			if selectChannelsForSpinners(workingSpinners, name, channels) {
				delete(workingSpinners, name)
			}
		}
	}

	mtv.multiPrinter.Stop()
	return nil
}

func selectChannelsForSpinners(workingSpinners map[string]*pterm.SpinnerPrinter, taskName string, channels taskChannels) bool {
	select {
	case err, ok := <-channels.errorChan:
		if ok {
			workingSpinners[taskName].Fail(err)
		}
	case commitList, ok := <-channels.valuesChan:
		if ok {
			if len(commitList.Commits) == 0 {
				workingSpinners[taskName].Warning(
					fmt.Sprintf("skipping '%s' no user activity detected", taskName),
				)
			} else {
				text := buildTreeString(
					fmt.Sprintf(
						"found %d commits for '%s'",
						len(commitList.Commits), taskName),
					commitsToStringList(commitList),
				)
				workingSpinners[taskName].Success(text)
			}
		}
	default:
		return false
	}
	return true
}

func buildTreeString(rootText string, values []string) string {
	children := []pterm.TreeNode{}
	for i := range values {
		children = append(children, pterm.TreeNode{
			Text: values[i],
		})
	}

	text, _ := pterm.DefaultTree.WithRoot(pterm.TreeNode{
		Text:     rootText,
		Children: children,
	}).Srender()
	return text
}

func startSpinnersWithPrinter(tasks map[string]taskChannels, multi *pterm.MultiPrinter) (map[string]*pterm.SpinnerPrinter, error) {
	spinners := map[string]*pterm.SpinnerPrinter{}
	for name := range tasks {
		text := fmt.Sprintf("Processing '%s'...", name)
		spinner, err := pterm.DefaultSpinner.
			WithWriter(multi.NewWriter()).
			WithStyle(pterm.NewStyle(pterm.FgGray)).
			WithText(text).
			Start()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to start spinner for '%s'", name)
		}

		spinners[name] = spinner
	}

	return spinners, nil
}

func commitsToStringList(list *github.CommitList) []string {
	stringList := []string{}
	for i := range list.Commits {
		commit := list.Commits[i]

		line := strings.Split(commit.GetCommit().GetMessage(), "\n")[0]
		stringList = append(stringList, line)
	}

	return stringList
}
