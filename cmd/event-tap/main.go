package main

import (
	"fmt"

	"github.com/draganm/event-tap/client"
	"github.com/draganm/event-tap/cmd/event-tap/create"
	"github.com/draganm/event-tap/cmd/event-tap/ls"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "event-tap",
		Description: "command line utility to control event-tap service",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "event-tap-server-url",
				EnvVars:  []string{"EVENT_TAP_SERVER_URL"},
				Required: true,
			},
		},
		Commands: []*cli.Command{
			ls.Command(),
			create.Command(),
		},
		EnableBashCompletion: true,
		Before: func(c *cli.Context) error {
			cl, err := client.New(c.String("event-tap-server-url"))
			if err != nil {
				return fmt.Errorf("could not create client: %w", err)
			}
			c.Context = client.ContextWithClient(c.Context, cl)
			return nil
		},
	}
	app.RunAndExitOnError()
}
