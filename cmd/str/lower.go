package main

import (
	"strings"

	"github.com/michielnijenhuis/cli"
)

var LowerCmd = &cli.Command{
	Name:        "lower",
	Description: "Converts a string to lowercase",
	Flags:       []cli.Flag{},
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The input string to convert to lowercase",
			Required:    true,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		ctx.IO.Write(strings.ToLower(input))
	},
}
