package main

import (
	"strings"

	"github.com/michielnijenhuis/cli"
)

var UpperCmd = &cli.Command{
	Name:        "upper",
	Description: "Converts a string to uppercase",
	Flags:       []cli.Flag{},
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The input string to convert to uppercase",
			Required:    true,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		ctx.IO.Write(strings.ToUpper(input))
	},
}
