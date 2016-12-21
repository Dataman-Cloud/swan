package command

import (
	"encoding/json"
	"fmt"
	//"github.com/Dataman-Cloud/swan/types"
	"github.com/urfave/cli"
	"os"
)

// NewScaleUpCommand returns the CLI command for "scale-up"
func NewScaleUpCommand() cli.Command {
	return cli.Command{
		Name:      "scale-up",
		Usage:     "scale up application",
		ArgsUsage: "[name]",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "instances",
				Usage: "instances to be scaled",
			},
		},
		Action: func(c *cli.Context) error {
			if err := scaleApp(c, "up"); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}
}

// NewScaleDownCommand returns the CLI command for "scale-down"
func NewScaleDownCommand() cli.Command {
	return cli.Command{
		Name:      "scale-down",
		Usage:     "scale down application",
		ArgsUsage: "[name]",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "instances",
				Usage: "instances to be scaled",
			},
		},
		Action: func(c *cli.Context) error {
			if err := scaleApp(c, "down"); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}
}

func scaleApp(c *cli.Context, op string) error {
	if len(c.Args()) == 0 {
		return fmt.Errorf("App ID required")
	}

	instances := c.Int("instances")
	if instances == 0 {
		return nil
	}

	httpClient := NewHTTPClient(fmt.Sprintf("%s/%s/scale-up", "/apps", c.Args()[0]))

	if op == "down" {
		httpClient = NewHTTPClient(fmt.Sprintf("%s/%s/scale-down", "/apps", c.Args()[0]))
	}

	var data struct {
		Instances int
	}

	data.Instances = instances
	payload, _ := json.Marshal(data)
	_, err := httpClient.Patch(payload)
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err.Error())
	}

	return nil
}
