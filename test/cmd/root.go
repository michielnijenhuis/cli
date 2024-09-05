package cmd

import (
	"github.com/michielnijenhuis/cli"
)

func Execute() {
	app := &cli.Application{
		Name:        "app",
		Version:     "v1.0.0",
		CatchErrors: true,
		AutoExit:    true,
		Commands: []*cli.Command{
			{
				Name:        "test",
				Aliases:     []string{"t"},
				Description: "Test command",
				Run:         test,
			},
		},
	}

	app.Run()
}

func test(c *cli.Ctx) {
	c.Spinner(func() {
		c.Exec("sleep 2", "", false)
	}, "Sleeping")
}
