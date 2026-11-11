package main

import (
	"github.com/michielnijenhuis/cli"
)

var KebabCmd = &cli.Command{
	Name:        "kebab",
	Description: "Convert string to kebab-case",
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The input string to convert to kebab-case",
			Required:    true,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		output := ToKebabCase(input)
		ctx.IO.Write(output)
	},
}
