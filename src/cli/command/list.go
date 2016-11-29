package command

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"

	"github.com/Dataman-Cloud/swan/src/types"
)

// NewListCommand returns the CLI command for "list"
func NewListCommand() cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: "list all applications",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "json",
				Usage: "List apps with json format",
			},
			cli.BoolFlag{
				Name:  "all",
				Usage: "List all apps",
			},
		},
		Action: func(c *cli.Context) error {
			if err := listApplications(c); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
			}
			return nil
		},
	}
}

// listApplications executes the "list" command.
func listApplications(c *cli.Context) error {
	httpClient := NewHTTPClient("/v1/apps")
	resp, err := httpClient.Get()
	if err != nil {
		return fmt.Errorf("Unable to do request: %s", err.Error())
	}
	defer resp.Body.Close()

	var apps []*types.Application
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		return err
	}

	listApps := make([]*types.Application, 0)
	if !c.Bool("all") {
		for _, app := range apps {
			if app.Status == "RUNNING" {
				listApps = append(listApps, app)
			}
		}
	} else {
		listApps = apps
	}

	if c.IsSet("json") {
		printJson(listApps)
	} else {
		printTable(listApps)
	}

	return nil
}

// printTable output apps list as table format.
func printTable(apps []*types.Application) {
	tb := tablewriter.NewWriter(os.Stdout)
	tb.SetHeader([]string{
		"ID",
		"Name",
		"Instances",
		"RunAS",
		"ClusterID",
		"Status",
		"Created",
	})
	for _, app := range apps {
		tb.Append([]string{
			app.ID,
			app.Name,
			fmt.Sprintf("%d", app.Instances),
			app.RunAs,
			app.ClusterId,
			app.Status,
			time.Unix(app.Created, 0).Format("2006-01-02 15:04:05"),
		})
	}
	tb.Render()
}

// printJson output apps list as json format.
func printJson(apps []*types.Application) error {
	data, err := json.Marshal(&apps)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, string(data))

	return nil
}
