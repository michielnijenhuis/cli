package cmd

import (
	"errors"

	"github.com/michielnijenhuis/cli"
)

func Execute() {
	cmd := &cli.Command{
		Name:        "test",
		Description: "This is a test command",
		Help:        "Show some help information",
		Handle: func(c *cli.Command) (int, error) {
			return 0, nil
		},
	}

	cmd.AddOption(&cli.InputOption{
		Name:         "bro",
		Description:  "Alternatives for the word `bro`",
		Mode:         cli.InputOptionRequired,
		DefaultValue: "brah",
		Validator: func(value cli.InputType) error {
			return errors.New("should not occur")
		},
	})

	app := &cli.Application{
		Name:    "app",
		Version: "v1.0.0",
		// CatchErrors: true,
	}

	app.Add(cmd)
	app.Run()
}
