package main

import (
	// "strings"
	"strings"

	"github.com/michielnijenhuis/cli"
)

var PadEndCmd = &cli.Command{
	Name:        "pad-end",
	Description: "Pads a string with the specified fill string from the end",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "fill",
			Shortcuts:   []string{"f"},
			Description: "The fill string to pad the string with",
			Value:       " ",
		},
	},
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "",
			Required:    true,
		},
		&cli.StringArg{
			Name:        "length",
			Description: "The length to pad the string to",
			Required:    true,
			Validator:   cli.IsInt,
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		input := ctx.IO.String("input")
		fill := ctx.IO.String("fill")
		length := ctx.IO.Int("length")

		currentLength := len(input)

		if currentLength >= length {
			ctx.IO.Write(input)
			return
		}

		padLength := length - currentLength
		pad := strings.Repeat(fill, padLength)
		ctx.IO.Write(input + pad)
	},
}
