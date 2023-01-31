package ls

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/draganm/event-tap/client"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name: "ls",
		Action: func(c *cli.Context) error {
			cl := client.FromContext(c.Context)
			entries, err := cl.List(c.Context)
			if err != nil {
				return fmt.Errorf("could not list taps: %w", err)
			}

			tw := tablewriter.NewWriter(os.Stdout)
			tw.SetHeader([]string{"name", "ID", "web hook URL"})
			for _, e := range entries {
				tw.Append([]string{e.Name, e.ID, e.WebhookURL})
			}
			tw.Render()
			return nil
		},
	}
}
