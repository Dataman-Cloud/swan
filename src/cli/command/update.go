package command

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

// NewUpdateCommand implement the CLI command for "update"
func NewUpdateCommand() cli.Command {
	return cli.Command{
		Name:      "update",
		Usage:     "Update app",
		ArgsUsage: "[name]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "full",
				Usage: "update entire application",
			},
			cli.IntFlag{
				Name:  "instances",
				Usage: "instances to be updated",
			},
		},
		Action: func(c *cli.Context) error {
			if err := updateApp(c); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}
}

func updateApp(c *cli.Context) error {
	return fmt.Errorf("update has not implemented!")
}
