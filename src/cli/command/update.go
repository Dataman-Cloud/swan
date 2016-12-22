package command

import (
	"fmt"
	//"github.com/Dataman-Cloud/swan/src/types"
	"encoding/json"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
)

// NewUpdateCommand implement the CLI command for "update"
func NewUpdateCommand() cli.Command {
	return cli.Command{
		Name:      "update",
		Usage:     "Update app",
		ArgsUsage: "[name]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "version",
				Usage: "version to updated to",
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

func NewProceedUpdateCommand() cli.Command {
	return cli.Command{
		Name:      "update-proceed",
		Usage:     "Proceed update app",
		ArgsUsage: "[name]",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "instances",
				Usage: "instances to be updated",
			},
		},
		Action: func(c *cli.Context) error {
			if err := proceedUpdateApp(c); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}

}

func NewCancelUpdateCommand() cli.Command {
	return cli.Command{
		Name:      "update-cancel",
		Usage:     "Cancel app update",
		ArgsUsage: "[name]",
		Action: func(c *cli.Context) error {
			if err := cancelUpdateApp(c); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}

}

func updateApp(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return fmt.Errorf("App ID required")
	}

	if c.String("version") == "" {
		return fmt.Errorf("Version must be specified. --version")
	}

	payload, err := ioutil.ReadFile(c.String("version"))
	if err != nil {
		return fmt.Errorf("Read json file failed: %s", err.Error())
	}

	//var version types.Version
	//if err := json.Unmarshal(file, &version); err != nil {
	//	return fmt.Errorf("Unmarshal error: %s", err.Error())
	//}

	httpClient := NewHTTPClient(fmt.Sprintf("/apps/%s", c.Args()[0]))

	_, err = httpClient.Put(payload)
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err.Error())
	}

	return nil
}

func cancelUpdateApp(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return fmt.Errorf("App ID required")
	}

	httpClient := NewHTTPClient(fmt.Sprintf("/apps/%s/cancel-update", c.Args()[0]))
	_, err := httpClient.Patch(nil)
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err.Error())
	}

	return nil
}

func proceedUpdateApp(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return fmt.Errorf("App ID required")
	}

	instances := c.Int("instances")
	if instances == 0 {
		return nil
	}
	var data struct {
		Instances int
	}

	data.Instances = instances
	payload, _ := json.Marshal(data)

	httpClient := NewHTTPClient(fmt.Sprintf("/apps/%s/proceed-update", c.Args()[0]))
	_, err := httpClient.Patch(payload)
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err.Error())
	}

	return nil
}
