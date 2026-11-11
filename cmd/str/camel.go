package main

import (
	"github.com/michielnijenhuis/cli"
)

var CamelCmd = &cli.Command{
	Name:        "camel",
	Description: "Convert string to camelCase",
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The input string to convert to camelCase",
			Required:    true,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		output := ToCamelCase(input)
		ctx.IO.Write(output)
	},
}
