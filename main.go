package main

import (
	"os"

	"github.com/pPrecel/PKUP/cmd"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var (
	version      = "local"
	commit       = "local"
	date         = "local"
	buildOs      = "local"
	buildArch    = "local"
	projectOwner = "pPrecel"
	projectRepo  = "pkup-gen"
	pkupClientID = ""
)

func main() {
	log := pterm.DefaultLogger.
		WithTime(false)
	opts := &cmd.Options{
		BuildVersion: version,
		BuildCommit:  commit,
		BuildDate:    date,
		BuildOs:      buildOs,
		BuildArch:    buildArch,
		ProjectOwner: projectOwner,
		ProjectRepo:  projectRepo,
		PkupClientID: pkupClientID,
		Log:          log,
	}

	app := &cli.App{
		Name:  "pkup",
		Usage: "Easly generate .diff files with all users merged content in the last PKUP period",
		UsageText: "pkup gen \\\n" +
			"\t\t--username <username> \\\n" +
			"\t\t--repo <org1>/<repo1> \\\n" +
			"\t\t--repo <org2>/<repo2>",

		Flags: []cli.Flag{
			cli.HelpFlag,
		},
		Commands: []*cli.Command{
			cmd.NewGenCommand(opts),
			cmd.NewComposeCommand(opts),
			cmd.NewVersionCommand(opts),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal("program error", log.Args("error", err))
	}
}
