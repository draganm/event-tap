package create

import (
	"fmt"

	"github.com/draganm/event-tap/client"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{

		Name: "create",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "webhook-url",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "code",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "batch-limit",
				Value: 100,
			},
		},

		Action: func(c *cli.Context) error {
			cl := client.FromContext(c.Context)
			id, err := cl.CreateTap(c.Context, client.CreateTapOptions{
				Name:       c.String("name"),
				Code:       c.String("code"),
				WebhookURL: c.String("webhook-url"),
				BatchLimit: c.Int("batch-limit"),
			})
			if err != nil {
				return fmt.Errorf("could not list taps: %w", err)
			}

			fmt.Println("created", id)

			return nil
		},
	}
}
