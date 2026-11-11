package main

import (
	"os"
	"strings"

	"github.com/michielnijenhuis/cli"
)

var StartsWithCmd = &cli.Command{
	Name:        "starts-with",
	Description: "Checks if a string starts with a specified prefix",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "output",
			Description: "Output the result as `true` or `false`",
			Shortcuts:   []string{"o"},
		},
	},
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The input string to check",
			Required:    true,
		},
		&cli.StringArg{
			Name:        "prefix",
			Description: "The prefix to check for",
			Required:    true,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		prefix := ctx.IO.String("prefix")

		result := strings.HasPrefix(input, prefix)

		if ctx.IO.Bool("output") {
			if result {
				ctx.IO.Write("true")
			} else {
				ctx.IO.Write("false")
			}
		}

		if !result {
			os.Exit(1)
		}
	},
}
