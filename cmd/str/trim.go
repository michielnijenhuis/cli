package main

import (
	"strings"

	"github.com/michielnijenhuis/cli"
)

var TrimCmd = &cli.Command{
	Name:        "trim",
	Description: "Removes the leading and/or trailing characters as specified in the <accent>chars</accent> flag from a string",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "start",
			Description: "Remove only the leading characters specified in the <accent>chars</accent> flag",
			Shortcuts:   []string{"s"},
		},
		&cli.BoolFlag{
			Name:        "end",
			Description: "Remove only the trailing characters specified in the <accent>chars</accent> flag",
			Shortcuts:   []string{"e"},
		},
		&cli.StringFlag{
			Name:        "chars",
			Description: "The characters to trim from the string",
			Shortcuts:   []string{"c"},
			Value:       ` \t\n\r`,
		},
	},
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The input string in which to perform the replacement.",
			Required:    true,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		start := ctx.IO.Bool("start")
		end := ctx.IO.Bool("end")
		chars := ctx.IO.String("chars")

		result := input
		if start {
			result = strings.TrimLeft(result, chars)
		}

		if end {
			result = strings.TrimRight(result, chars)
		}

		if !start && !end {
			result = strings.Trim(result, chars)
		}

		ctx.IO.Write(result)
	},
}
