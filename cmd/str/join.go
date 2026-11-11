package main

import (
	// "strings"
	"strings"

	"github.com/michielnijenhuis/cli"
)

var JoinCmd = &cli.Command{
	Name:        "join",
	Description: "Joins a list of strings with the specified glue",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "glue",
			Description: "The glue to use when joining the strings",
			Shortcuts:   []string{"g"},
		},
	},
	Arguments: []cli.Arg{
		&cli.ArrayArg{
			Name:        "input",
			Description: "The input strings to join",
			Min:         1,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.Array("input")
		glue := ctx.IO.String("glue")

		ctx.IO.Write(strings.Join(input, glue))
	},
}
