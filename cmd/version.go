package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

func NewVersionCommand(opts *Options) *cli.Command {
	actionOpts := &versionActionOpts{
		Options: opts,
	}
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Shows tool version",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "json",
				Usage:   "print output using standard log and JSON format",
				Action: func(_ *cli.Context, b bool) error {
					actionOpts.ci = b
					opts.Log = opts.Log.WithFormatter(pterm.LogFormatterJSON)
					return nil
				},
			},
			&cli.BoolFlag{
				Name:               "v",
				Usage:              "show more info",
				Destination:        &actionOpts.v,
				DisableDefaultText: true,
			},
			&cli.BoolFlag{
				Name:               "vv",
				Usage:              "show much more info",
				Destination:        &actionOpts.vv,
				DisableDefaultText: true,
			},
		},
		Action: func(_ *cli.Context) error {
			text, err := buildVersionOutput(actionOpts)
			fmt.Print(text)
			return err
		},
	}
}

func buildVersionOutput(opts *versionActionOpts) (string, error) {
	data := getVersionData(opts)

	if opts.ci {
		byteData, err := json.Marshal(data)
		return string(byteData), err
	}

	text := ""
	for key, value := range data {
		text += fmt.Sprintf("%s: %s\n", key, value)
	}

	return text, nil
}

func getVersionData(opts *versionActionOpts) map[string]string {
	data := map[string]string{}
	data["version"] = opts.BuildVersion
	if opts.v {
		data["project-url"] = fmt.Sprintf("https://github.com/%s/%s", opts.ProjectOwner, opts.ProjectRepo)
		data["build-commit"] = opts.BuildCommit
		data["build-data"] = opts.BuildDate
	} else if opts.vv {
		data["project-url"] = fmt.Sprintf("https://github.com/%s/%s", opts.ProjectOwner, opts.ProjectRepo)
		data["build-commit"] = opts.BuildCommit
		data["build-data"] = opts.BuildDate
		data["build-arch"] = opts.BuildArch
		data["build-os"] = opts.BuildOs
	}

	return data
}
