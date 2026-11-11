package main

import (
	"github.com/michielnijenhuis/cli"
)

var RepeatCmd = &cli.Command{
	Name:        "repeat",
	Description: "Repeats a string",
	Flags:       []cli.Flag{},
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The string to repeat",
			Required:    true,
		},
		&cli.StringArg{
			Name:        "count",
			Description: "The number of times to repeat the string",
			Required:    true,
			Validator:   cli.IsPositiveOrZeroInt,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		count := ctx.IO.Int("count")

		for i := 0; i < count; i++ {
			ctx.IO.Write(input)
		}
	},
}
