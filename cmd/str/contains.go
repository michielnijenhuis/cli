package main

import (
	"os"
	"strings"

	"github.com/michielnijenhuis/cli"
)

var ContainsCmd = &cli.Command{
	Name:        "contains",
	Description: "Checks if a string contains a specified substring",
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
			Name:        "substring",
			Description: "The substring to check for",
			Required:    true,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		substring := ctx.IO.String("substring")

		result := strings.Contains(input, substring)

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
