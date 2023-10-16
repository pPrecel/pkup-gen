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
	return mtv.multiPrinter.NewWriter()
}

func (mtv *dynamicMultiView) Add(name string, valuesChan chan []string, errorChan chan error) {
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
	case PRs, ok := <-channels.valuesChan:
		if ok {
			if len(PRs) == 0 {
				workingSpinners[taskName].Warning(
					fmt.Sprintf("skipping '%s' no user activity detected", taskName),
				)
			} else {
				text := buildPRsTreeString(
					fmt.Sprintf("found %d PRs for repo '%s'", len(PRs), taskName), PRs,
				)
				workingSpinners[taskName].Success(text)
			}
		}
	default:
		return false
	}
	return true
}

func buildPRsTreeString(rootText string, PRs []string) string {
	children := []pterm.TreeNode{}
	for i := range PRs {
		children = append(children, pterm.TreeNode{
			Text: PRs[i],
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
