package cmd

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
)

type Options struct {
	Version string
	Log     *pterm.Logger
}

type genActionOpts struct {
	*Options
	perdiod       int
	dir           string
	repos         map[string][]string
	username      string
	token         string
	enterpriseURL string
	withClosed    bool
}

func (opts *genActionOpts) setDefaults() error {
	if opts.dir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get pwd error: %s", err.Error())
		}
		opts.dir = pwd
	}

	return nil
}
