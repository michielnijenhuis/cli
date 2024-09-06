package cmd

import (
	"log"

	"github.com/michielnijenhuis/cli"
)

func Execute() {
	autoExit := true

	app := &cli.Application{
		Name:        "app",
		Version:     "v1.0.0",
		CatchErrors: true,
		AutoExit:    autoExit,
		Commands: []*cli.Command{
			{
				Name:        "test",
				Aliases:     []string{"t"},
				Description: "Test command",
				Run:         test,
			},
		},
	}

	if err := app.Run(); err != nil && !autoExit {
		log.Fatalln(err)
	}
}

func test(c *cli.Ctx) {
	panic("oh no")
	c.Note("Wadup\nAll good bro?")
	c.Ok("Wadup, how you doin")
	c.Info("All good")
	c.Warn("Sure?")
	c.Error("Nah brah")
}
