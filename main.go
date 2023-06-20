package main

import (
	"fmt"
	"os"

	"github.com/pPrecel/PKUP/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	logo = `
.______    __  ___  __    __  .______
|   _  \  |  |/  / |  |  |  | |   _  \
|  |_)  | |  '  /  |  |  |  | |  |_)  |__ _  ___ _ __
|   ___/  |    <   |  |  |  | |   ___// _' |/ _ \ '_ \
|  |      |  .  \  |  '--'  | |  |   | (_| |  __/ | | |
| _|      |__|\__\  \______/  | _|    \__, |\___|_| |_|
                                      |___/`
)

var (
	version = "local"
)

func main() {
	log := logrus.New()

	// print logo before any action
	fmt.Printf("%s\n\n", logo)

	opts := &cmd.Options{
		Version: version,
		Log:     log,
	}

	app := &cli.App{
		Name:  "pkup",
		Usage: "Easly generate .patch files with all users merged content in the last PKUP period",
		UsageText: "pkup gen --token <personal-access-token> \\\n" +
			"\t\t--username <username> \\\n" +
			"\t\t--repo <org1>/<repo1> \\\n" +
			"\t\t--repo <org2>/<repo2>",

		Flags: []cli.Flag{
			cli.HelpFlag,
		},
		Commands: []*cli.Command{
			cmd.NewGenCommand(opts),
			cmd.NewVersionCommand(opts),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
