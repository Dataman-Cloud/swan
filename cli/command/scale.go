package command

import (
	//"encoding/json"
	"fmt"
	//"github.com/Dataman-Cloud/swan/types"
	"github.com/urfave/cli"
	"os"
)

// NewScaleCommand returns the CLI command for "scale"
func NewScaleCommand() cli.Command {
	return cli.Command{
		Name:      "scale",
		Usage:     "scale down or scale up application",
		ArgsUsage: "[name]",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "instances",
				Usage: "instances to be scaled",
			},
		},
		Action: func(c *cli.Context) error {
			if err := scaleApplication(c); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}
}

// scaleApplication executes the "scale" command.
func scaleApplication(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return fmt.Errorf("App ID required")
	}

	instances := c.String("instances")
	if instances == "" {
		return fmt.Errorf("instances must be specified")
	}

	httpClient := NewHTTPClient(fmt.Sprintf("%s/%s/scale?instances=%s", "/v1/apps", c.Args()[0], instances))
	_, err := httpClient.Post(nil)
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err.Error())
	}

	return nil
}
