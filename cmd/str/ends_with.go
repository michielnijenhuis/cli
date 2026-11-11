package main

import (
	"os"
	"strings"

	"github.com/michielnijenhuis/cli"
)

var EndsWithCmd = &cli.Command{
	Name:        "ends-with",
	Description: "Checks if a string ends with a specified suffix",
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
			Name:        "suffix",
			Description: "The suffix to check for",
			Required:    true,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		suffix := ctx.IO.String("suffix")

		result := strings.HasSuffix(input, suffix)

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
