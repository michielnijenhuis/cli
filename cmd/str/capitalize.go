package main

import (
	"strings"

	"github.com/michielnijenhuis/cli"
)

var CapitalizeCmd = &cli.Command{
	Name:        "capitalize",
	Description: "Capitalizes the first letter of the string",
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The string to capitalize",
			Required:    true,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")

		if len(input) == 0 {
			ctx.IO.Write(input)
			return
		}

		if len(input) == 1 {
			ctx.IO.Write(strings.ToUpper(input))
			return
		}

		capitalized := strings.ToUpper(string(input[0])) + input[1:]
		ctx.IO.Write(capitalized)
	},
}
