package logo

import (
	"fmt"

	"github.com/pterm/pterm"
)

var (
	pkupLayers = []string{
		`.______    __  ___  __    __  .______`,
		`|   _  \  |  |/  / |  |  |  | |   _  \`,
		`|  |_)  | |  '  /  |  |  |  | |  |_)  |`,
		`|   ___/  |    <   |  |  |  | |   ___/`,
		`|  |      |  .  \  |  '--'  | |  |   `,
		`| _|      |__|\__\  \______/  | _|    `,
		`                                      `,
	}
	genLayers = []string{
		``,
		``,
		`__ _  ___ _ __`,
		`/ _' |/ _ \ '_ \`,
		`| (_| |  __/ | | |`,
		`\__, |\___|_| |_|`,
		`|___/      `,
	}
)

func Build(version string) string {
	return fmt.Sprint(
		pkupLayers[0], genLayers[0], "\n",
		pkupLayers[1], genLayers[1], "\n",
		pkupLayers[2], pterm.Bold.Sprint(pterm.Red(genLayers[2])), "\n",
		pkupLayers[3], pterm.Bold.Sprint(pterm.Red(genLayers[3])), "\n",
		pkupLayers[4], pterm.Bold.Sprint(pterm.Red(genLayers[4])), "\n",
		pkupLayers[5], pterm.Bold.Sprint(pterm.Red(genLayers[5])), "\n",
		pkupLayers[6], pterm.Bold.Sprint(pterm.Red(genLayers[6])), pterm.Bold.Sprint(pterm.Gray(version)),
		// TODO: pterm.Red(pkupLayers[6])
		// TODO: support 'New Version Available!' info
	)
}
