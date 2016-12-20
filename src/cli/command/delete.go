package command

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

// NewDeleteCommand returns the CLI command for "delete"
func NewDeleteCommand() cli.Command {
	return cli.Command{
		Name:      "delete",
		Usage:     "delete application",
		ArgsUsage: "[name]",
		Action: func(c *cli.Context) error {
			if err := deleteApplication(c); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}
}

// deleteApplication executes the "delete" command.
func deleteApplication(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return fmt.Errorf("name required")
	}

	httpClient := NewHTTPClient(fmt.Sprintf("%s/%s", "/apps", c.Args()[0]))
	_, err := httpClient.Delete()
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err.Error())
	}

	return nil
}
