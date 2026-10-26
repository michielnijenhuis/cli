package cmd

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"

	cli "github.com/michielnijenhuis/cli/internal"
)

var ArgCmd = &cli.Command{
	Name:        "arg",
	Description: "Print the value of the argument at the given index",
	Aliases:     []string{"a"},
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "name",
			Description: "The name of the argument to print",
			Required:    true,
		},
	},
	RunE: func(args []string, ctx *cli.CommandCtx) error {
		payload := os.Getenv("CLI_PAYLOAD")
		if payload == "" {
			return errors.New("Payload is not available")
		}

		name := ctx.IO.String("name")

		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return err
		}

		var p cli.Payload
		err = json.Unmarshal(decoded, &p)
		if err != nil {
			return err
		}

		argValue := p.Args[name]
		switch value := argValue.(type) {
		case string:
			ctx.IO.Write(value)
		}

		return nil
	},
}
