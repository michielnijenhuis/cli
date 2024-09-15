package cmd

import (
	"github.com/michielnijenhuis/cli"
)

func Execute() {
	app := &cli.Application{
		Name:        "app",
		Version:     "v1.0.0",
		CatchErrors: true,
		Commands: []*cli.Command{
			{
				Name:        "test",
				Aliases:     []string{"t"},
				Description: "Test command",
				Run:         test,
			},
		},
	}

	app.RunExit()
}

func test(c *cli.Ctx) {
	answer, err := c.Ask("Wadup?", "")
	if err != nil {
		c.Err(err)
		return
	}

	c.Spinner(func() {
		c.Sh("sleep 2")
	}, "Processing...")

	c.NewLine(1)
	c.Info(answer)
}
