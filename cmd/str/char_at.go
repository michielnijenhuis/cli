package main

import (
	// "strings"
	"github.com/michielnijenhuis/cli"
)

var CharAtCmd = &cli.Command{
	Name:        "char-at",
	Description: "Returns the character at the specified index",
	Flags:       []cli.Flag{},
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The string to get the character from",
			Required:    true,
		},
		&cli.StringArg{
			Name:        "index",
			Description: "The index of the character to return",
			Required:    true,
			Validator:   cli.IsInt,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		index := ctx.IO.Int("index")

		if index < 0 || index >= len(input) {
			ctx.IO.Write("")
			return
		}

		ctx.IO.Write(string(input[index]))
	},
}
