package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
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

type composeActionOpts struct {
	*Options

	config string
	since  cli.Timestamp
	until  cli.Timestamp
	ci     bool
}

type sendActionOpts struct {
	*Options

	config          string
	reportTimestamp cli.Timestamp
	ci              bool
}

type genActionOpts struct {
	*Options

	since         cli.Timestamp
	until         cli.Timestamp
	outputDir     string
	token         string
	username      string
	enterpriseURL string
	templatePath  string
	orgs          []string
	repos         []string
	reportFields  map[string]string
	uniqueOnly    bool
	allBranches   bool
	ci            bool
}

type versionActionOpts struct {
	*Options
	v  bool
	vv bool
	ci bool
}

func (opts *genActionOpts) setDefaults() error {
	if opts.outputDir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get pwd error: %s", err.Error())
		}
		opts.outputDir = pwd
	}

	return nil
}

func parseReportFields(args []string) (map[string]string, error) {
	reportFields := map[string]string{}
	for _, field := range args {
		vals := strings.Split(field, "=")
		if len(vals) != 2 {
			return nil, fmt.Errorf("failed to parse '%s' report field", field)
		}

		reportFields[vals[0]] = vals[1]
	}

	return reportFields, nil
}
