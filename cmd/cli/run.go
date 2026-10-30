package main

import (
	"os"

	"github.com/michielnijenhuis/cli"
)

var RunCmd = &cli.Command{
	Name:        "run",
	Description: "Run a command as defined in a yaml file",
	Hidden:      true,
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "config",
			Description: "The .yaml configuration file for the CLI app or command",
			Required:    true,
			Validator:   cli.IsFile,
		},
		&cli.StringArg{
			Name:        "script",
			Description: "The .sh script that provides the command implementations",
			Validator:   cli.IsFile,
		},
	},
	RunE: func(args []string, ctx *cli.CommandCtx) error {
		config := ctx.IO.String("config")
		script := ctx.IO.String("script")
		hasScript := script != ""

		cmd, err := cli.Parse(config)
		if err != nil {
			ctx.IO.Critical(err)
			os.Exit(1)
		}

		os.Setenv("CLI_CMD_CONFIG", config)
		os.Setenv("CLI_CMD_SCRIPT", script)

		cli.ConfigureGumTheme()

		ignoreArgs := make(map[string]bool)
		ignoreArgs["run"] = false
		ignoreArgs[config] = false
		if hasScript {
			ignoreArgs[script] = false
		}

		forwardedArgs := make([]string, 0, len(args))
		for _, arg := range args {
			ignored, exists := ignoreArgs[arg]

			if !exists || ignored {
				forwardedArgs = append(forwardedArgs, arg)
			} else if exists && !ignored {
				ignoreArgs[arg] = true
			}
		}

		if len(forwardedArgs) == 1 && forwardedArgs[0] == "" {
			forwardedArgs = []string{}
		}

		return cmd.Execute(forwardedArgs...)

		// return nil
	},
}
