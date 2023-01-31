package delete

import (
	"fmt"

	"github.com/draganm/event-tap/client"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{

		Name: "delete",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Required: true,
			},
		},

		Action: func(c *cli.Context) error {
			cl := client.FromContext(c.Context)
			err := cl.Delete(c.Context, c.String("id"))
			if err != nil {
				return fmt.Errorf("could not delete tap: %w", err)
			}
			fmt.Println("deleted")
			return nil
		},
	}
}
