package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"

	"github.com/michielnijenhuis/cli"
)

var FlagCmd = &cli.Command{
	Name:        "flag",
	Description: "Print the value of the given flag",
	Aliases:     []string{"f"},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "exit",
			Shortcuts:   []string{"x"},
			Description: "Exit with 1 if false, or 0 if true. Works for boolean flags only.",
		},
	},
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "flag",
			Description: "The name of the flag to print",
			Required:    true,
		},
	},
	RunE: func(args []string, ctx *cli.CommandCtx) error {
		payload := os.Getenv("CLI_PAYLOAD")
		if payload == "" {
			return errors.New("Payload is not available")
		}

		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return err
		}

		var p cli.Payload
		err = json.Unmarshal(decoded, &p)
		if err != nil {
			return err
		}

		exit := ctx.IO.Bool("exit")
		flagValue := p.Flags[ctx.IO.String("flag")]
		switch value := flagValue.(type) {
		case string:
			ctx.IO.Write(value)
		case bool:
			if value {
				if exit {
					os.Exit(0)
				}

				ctx.IO.Write("true")
			} else {
				if exit {
					os.Exit(1)
				}

				ctx.IO.Write("false")
			}
		}

		return nil
	},
}
