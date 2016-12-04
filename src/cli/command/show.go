package command

import (
	"encoding/json"
	"fmt"
	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	"os"
)

// NewShowCommand returns the CLI command for "show"
func NewShowCommand() cli.Command {
	return cli.Command{
		Name:      "show",
		Usage:     "show application or task info",
		ArgsUsage: "[name]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "json",
				Usage: "List tasks with json format",
			},
		},

		Action: func(c *cli.Context) error {
			if err := showApplication(c); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}
}

// showApplication executes the "show" command.
func showApplication(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return fmt.Errorf("App ID required")
	}

	httpClient := NewHTTPClient(fmt.Sprintf("/v1/apps/%s/tasks", c.Args()[0]))
	resp, err := httpClient.Get()
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err.Error())
	}
	defer resp.Body.Close()

	var tasks []*types.Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return err
	}

	data, err := json.Marshal(&tasks)
	if err != nil {
		return err
	}

	if c.IsSet("json") {
		fmt.Fprintln(os.Stdout, string(data))
	} else {
		printTaskTable(tasks)
	}

	return nil
}

// printTable output tasks list as table format.
func printTaskTable(tasks []*types.Task) {
	tb := tablewriter.NewWriter(os.Stdout)
	tb.SetHeader([]string{
		"Name",
		"APPID",
		"CPUS",
		"MEM",
		"DISK",
		"NETWORK",
		"ADDRESS",
		"STATUS",
	})
	for _, task := range tasks {
		tb.Append([]string{
			task.Name,
			task.AppId,
			fmt.Sprintf("%.2f", task.Cpus),
			fmt.Sprintf("%.f", task.Mem),
			fmt.Sprintf("%.f", task.Disk),
			task.Network,
			task.AgentHostname,
			task.Status,
		})
	}
	tb.Render()
}
