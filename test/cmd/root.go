package cmd

import (
	"log"
	"strings"

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
		Arguments: []cli.Arg{
			&cli.StringArg{
				Name:        "id",
				Description: "User ID",
				Required:    true,
			},
		},
		Run: func(io *cli.IO) {
			io.Writeln(strings.Join(io.Args, ","))
		},
	}

	cmd := &cli.Command{
		Name:        "cmd",
		Description: "Beautiful command",
		Aliases:     []string{"c"},
		Run: func(io *cli.IO) {
			io.Writeln(strings.Join(io.Args, ","))
		},
	}

	cmd2 := &cli.Command{
		Name:        "cmd2",
		Description: "Beautiful command 2",
		Run: func(io *cli.IO) {
			io.Writeln(strings.Join(io.Args, ","))
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
				Required:    true,
			},
		},
		Run: func(io *cli.IO) {
			io.Writeln(strings.Join(io.Args, ","))
		},
	}

	cmd.AddCommand(sub)

	if err := app.Execute(); err != nil {
		log.Fatalln(err)
	}
}
