package cmd

import (
	"log"

	"github.com/michielnijenhuis/cli"
)

func Execute() {
	app := &cli.Command{
		Name:           "app",
		Description:    "Beautiful CLI application",
		Version:        "v1.0.0",
		PromptForInput: true,
		AutoExit:       true,
		CatchErrors:    true,
		Run: func(io *cli.IO) {
			io.Writeln("Hello from app")
		},
	}

	cmd := &cli.Command{
		Name:        "cmd",
		Description: "Beautiful command",
		Aliases:     []string{"c"},
		Run: func(io *cli.IO) {
			io.Writeln("Hello from app->cmd")
		},
	}

	cmd2 := &cli.Command{
		Name:        "a-very-long-name",
		Description: "Beautiful command 2",
		Run: func(io *cli.IO) {
			io.Writeln("Hello from app->cmd2")
		},
	}

	app.AddCommand(cmd)
	app.AddCommand(cmd2)

	sub := &cli.Command{
		Name:        "sub",
		Description: "Beautiful sub-subcommand",
		Aliases:     []string{"s"},
		Arguments: []cli.Arg{
			&cli.StringArg{
				Name:        "arg",
				Description: "Argument",
			},
		},
		Run: func(io *cli.IO) {
			io.Writeln("Hello from app->cmd->sub")
		},
	}

	cmd.AddCommand(sub)

	if err := app.Execute(); err != nil {
		log.Fatalln(err)
	}
}
