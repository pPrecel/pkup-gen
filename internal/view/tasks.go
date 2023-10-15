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

type MultiTaskView interface {
	Run() error
	Add(string, chan []string, chan error)
	NewWriter() io.Writer
}

type taskChannels struct {
	valuesChan chan []string
	errorChan  chan error
}

type multiTaskView struct {
	multiPrinter pterm.MultiPrinter
	tasks        map[string]taskChannels
}

func NewMultiTaskView() MultiTaskView {
	return &multiTaskView{
		multiPrinter: pterm.DefaultMultiPrinter,
		tasks:        map[string]taskChannels{},
	}
}

func (mtv *multiTaskView) NewWriter() io.Writer {
	return mtv.multiPrinter.NewWriter()
}

func (mtv *multiTaskView) Add(name string, valuesChan chan []string, errorChan chan error) {
	mtv.tasks[name] = taskChannels{
		valuesChan: valuesChan,
		errorChan:  errorChan,
	}
}

func (mtv *multiTaskView) Run() error {
	workingSpinners, err := startSpinnersWithPrinter(mtv.tasks, &mtv.multiPrinter)
	if err != nil {
		return err
	}

	mtv.multiPrinter.Start()
	for len(workingSpinners) > 0 {
		for name, channels := range mtv.tasks {
			n := name
			chs := channels
			select {
			case err, ok := <-chs.errorChan:
				if ok {
					workingSpinners[n].Fail(err)
				}

				delete(workingSpinners, n)
			case PRs, ok := <-chs.valuesChan:
				if ok {
					if len(PRs) == 0 {
						workingSpinners[n].Warning(
							fmt.Sprintf("skipping '%s' no user activity detected", n),
						)
					} else {
						text := buildPRsTreeString(
							fmt.Sprintf("found %d PRs for repo '%s'", len(PRs), n), PRs,
						)
						workingSpinners[n].Success(text)
					}
				}
				delete(workingSpinners, n)
			default:
			}
		}
	}

	mtv.multiPrinter.Stop()
	return nil
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
