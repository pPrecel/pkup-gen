package cmd

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
)

type Options struct {
	BuildVersion string
	BuildCommit  string
	BuildDate    string
	BuildOs      string
	BuildArch    string
	ProjectOwner string
	ProjectRepo  string
	PkupClientID string
	Log          *pterm.Logger
}

type genActionOpts struct {
	*Options

	perdiod       int
	dir           string
	token         string
	username      string
	enterpriseURL string
	repos         map[string][]string
	withClosed    bool
	ci            bool
}

type versionActionOpts struct {
	*Options
	v  bool
	vv bool
	ci bool
}

func (opts *genActionOpts) setDefaults() error {
	if opts.dir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get pwd error: %s", err.Error())
		}
		opts.dir = pwd
	}

	return nil
}
