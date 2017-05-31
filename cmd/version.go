package cmd

import (
	"os"

	"github.com/Dataman-Cloud/swan/version"
	"github.com/urfave/cli"
)

func VersionCmd() cli.Command {
	return cli.Command{
		Name:        "version",
		Description: "show version",
		Usage:       "display version info",
		Action: func(c *cli.Context) error {
			return version.TextFormatTo(os.Stdout)
		},
	}
}
