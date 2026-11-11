package main

import (
	"github.com/michielnijenhuis/cli"
)

var SnakeCmd = &cli.Command{
	Name:        "snake",
	Description: "Convert string to snake_case",
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The input string to convert to snake_case",
			Required:    true,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		output := ToSnakeCase(input)
		ctx.IO.Write(output)
	},
}
