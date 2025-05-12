package view

import (
	"fmt"
	"io"

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
	w := mtv.multiPrinter.NewWriter()

	// write anything to avoid problems with empty buffer on the pterm side
	_, _ = w.Write([]byte{0})

	return w
}

func (mtv *dynamicMultiView) Add(name string, valuesChan chan []*RepoCommit, errorChan chan error) {
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

	_, err = mtv.multiPrinter.Start()
	if err != nil {
		return err
	}

	for len(workingSpinners) > 0 {
		for name, channels := range mtv.tasks {
			if selectChannelsForSpinners(workingSpinners, name, channels) {
				delete(workingSpinners, name)
			}
		}
	}

	_, err = mtv.multiPrinter.Stop()
	return err
}

func selectChannelsForSpinners(workingSpinners map[string]*pterm.SpinnerPrinter, taskName string, channels taskChannels) bool {
	select {
	case err, ok := <-channels.errorChan:
		if ok {
			workingSpinners[taskName].Fail(err)
		}
	case commits, ok := <-channels.valuesChan:
		if ok {
			if len(commits) == 0 {
				workingSpinners[taskName].Warning(
					fmt.Sprintf("skipping %s no user activity detected", taskName),
				)
			} else {
				text := buildTreeString(
					fmt.Sprintf(
						"found %d commits for %s",
						len(commits), taskName),
					commitsToStringList(commits),
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
		text := fmt.Sprintf("Processing %s...", name)
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

func commitsToStringList(commits []*RepoCommit) []string {
	stringList := []string{}
	for _, commit := range commits {
		stringList = append(stringList, fmt.Sprintf("%s/%s - %s", commit.Org, commit.Repo, commit.Message))
	}

	return stringList
}
