package main

import (
	"github.com/michielnijenhuis/cli"
)

var PascalCmd = &cli.Command{
	Name:        "pascal",
	Description: "Convert string to PascalCase",
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The input string to convert to PascalCase",
			Required:    true,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		output := ToPascalCase(input)
		ctx.IO.Write(output)
	},
}
