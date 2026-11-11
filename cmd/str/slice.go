package main

import (
	"fmt"
	"strconv"

	"github.com/michielnijenhuis/cli"
)

var SliceCmd = &cli.Command{
	Name:        "slice",
	Description: "Returns a slice of the input string",
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The input string to slice",
			Required:    true,
		},
		&cli.StringArg{
			Name:        "start",
			Description: "The start index of the slice",
			Value:       "0",
			Validator:   cli.IsPositiveOrZeroInt,
		},
		&cli.StringArg{
			Name:        "end",
			Description: "The end index of the slice",
			Validator:   cli.IsInt,
		},
	},
	RunE: func(args []string, ctx *cli.CommandCtx) error {
		input := ctx.IO.String("input")
		start := ctx.IO.Int("start")
		end := len(input)

		endInput := ctx.IO.String("end")
		if endInput != "" {
			num, err := strconv.Atoi(endInput)
			if err != nil {
				return fmt.Errorf("value [%s] is not a valid integer:\n\t- %w", endInput, err)
			}

			end = num
		}

		if start > len(input) {
			ctx.IO.Write("")
			return nil
		}

		if end > len(input) {
			end = len(input)
		}

		if end < 0 {
			end = len(input) + end
		}

		if end < start {
			end = start
		}

		slice := input[start:end]

		ctx.IO.Write(slice)

		return nil
	},
}
