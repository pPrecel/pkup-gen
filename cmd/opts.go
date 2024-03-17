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
	repos         map[string][]string
	reportFields  map[string]string
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

func parseReposMap(log *pterm.Logger, args []string) (map[string][]string, error) {
	repos := map[string][]string{}
	for i := range args {
		arg := args[i]

		log.Debug("parsing flag", log.Args("argument", arg))
		argSlice := strings.Split(arg, "/")
		if len(argSlice) != 2 {
			return nil, fmt.Errorf("repo '%s' must be in format <org>/<repo>", arg)
		}

		repos[argSlice[0]] = append(repos[argSlice[0]], argSlice[1])
	}

	return repos, nil
}
