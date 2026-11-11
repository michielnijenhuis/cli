package main

import (
	"regexp"
	"strings"

	"github.com/michielnijenhuis/cli"
)

var ReplaceCmd = &cli.Command{
	Name:        "replace",
	Description: "Replace substrings in a string",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "regex",
			Description: "Treat the search value as a regular expression.",
			Shortcuts:   []string{"r"},
		},
		&cli.StringFlag{
			Name:        "count",
			Description: "The number of replacements to perform.",
			Shortcuts:   []string{"c", "n"},
			Value:       "-1",
			Validator:   cli.IsInt,
		},
	},
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "input",
			Description: "The input string in which to perform the replacement.",
			Required:    true,
		},
		&cli.StringArg{
			Name:        "searchValue",
			Description: "A string or regular expression to search for.",
			Required:    true,
		},
		&cli.StringArg{
			Name:        "replaceValue",
			Description: "A string containing the text to replace.",
			Required:    true,
		},
	},
	RunE: func(args []string, ctx *cli.CommandCtx) error {
		input := ctx.IO.String("input")
		searchValue := ctx.IO.String("searchValue")
		regex := ctx.IO.Bool("regex")
		replaceValue := ctx.IO.String("replaceValue")

		if regex {
			searchRegex, err := regexp.Compile(searchValue)
			if err != nil {
				return err
			}

			result := searchRegex.ReplaceAllString(input, replaceValue)
			ctx.IO.Write(result)
		} else {
			count := ctx.IO.Int("count")
			result := input
			result = strings.Replace(input, searchValue, replaceValue, count)

			ctx.IO.Write(result)
		}

		return nil
	},
}
