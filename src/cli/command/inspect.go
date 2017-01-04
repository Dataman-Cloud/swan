package command

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Dataman-Cloud/swan/src/types"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

// NewInspectCommand returns the CLI command for "show"
func NewInspectCommand() cli.Command {
	return cli.Command{
		Name:      "inspect",
		Usage:     "inspect application info",
		ArgsUsage: "[name]",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "json",
				Usage: "List tasks with json format",
			},
			cli.BoolFlag{
				Name:  "history",
				Usage: "List task histories",
			},
		},

		Action: func(c *cli.Context) error {
			if err := inspectApplication(c); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}
}

// inspectApplication executes the "inspect" command.
func inspectApplication(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return fmt.Errorf("App ID required")
	}

	httpClient := NewHTTPClient(fmt.Sprintf("/apps/%s", c.Args()[0]))
	resp, err := httpClient.Get()
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err.Error())
	}
	defer resp.Body.Close()

	var app *types.App
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return err
	}

	data, err := json.Marshal(&app.Tasks)
	if err != nil {
		return err
	}

	if c.IsSet("json") {
		fmt.Fprintln(os.Stdout, string(data))
	} else {
		printTaskTable(app.Tasks)
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
		"ADDRESS",
		"STATUS",
		"VERSIONID",
		"HISTORIES",
	})
	for _, task := range tasks {
		tb.Append([]string{
			task.ID,
			task.AppID,
			fmt.Sprintf("%.2f", task.CPU),
			fmt.Sprintf("%.f", task.Mem),
			fmt.Sprintf("%.f", task.Disk),
			task.AgentHostname,
			task.Status,
			task.VersionID,
			fmt.Sprintf("%d", len(task.History)),
		})
	}
	tb.Render()
}

//func printHistoies(histories []*api.TaskHistory) {
//	tb := tablewriter.NewWriter(os.Stdout)
//	tb.SetHeader([]string{
//		"ID",
//		"CPU",
//		"MEM",
//		"DISK",
//		"ADDRESS",
//		"REASON",
//		"VERSIONID",
//	})
//	for _, history := range histories {
//		tb.Append([]string{
//			history.ID,
//			fmt.Sprintf("%.2f", history.Cpu),
//			fmt.Sprintf("%.f", history.Mem),
//			fmt.Sprintf("%.f", history.Disk),
//			history.AgentHostname,
//			history.Reason,
//			history.VersionId,
//		})
//	}
//	tb.Render()
//}
