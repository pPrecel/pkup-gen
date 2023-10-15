package cmd

import (
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
}
