package main

import (
	"github.com/michielnijenhuis/cli"
)

var LenCmd = &cli.Command{
	Name:        "len",
	Description: "Returns the length of the input string",
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The input string",
			Required:    true,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		ctx.IO.Writef("%d", len(input))
	},
}
