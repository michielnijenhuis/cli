package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/michielnijenhuis/cli"
)

var UninstallCmd = &cli.Command{
	Name:        "uninstall",
	Description: "Uninstall a command as defined in a yaml file as an executable",
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "name",
			Description: "The name of the command you wish to uninstall",
			Required:    true,
		},
	},
	RunE: func(args []string, ctx *cli.CommandCtx) error {
		name := ctx.IO.String("name")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to determine home directory: %w", err)
		}

		commandPath := filepath.Join(homeDir, ".local", "bin", name)
		err = cli.IsFile(commandPath)
		if err != nil {
			return fmt.Errorf("command %s was not found at %s", name, commandPath)
		}

		err = os.Remove(commandPath)
		if err != nil {
			return fmt.Errorf("failed to remove command \"%s\" at %s: %w", name, commandPath, err)
		}

		ctx.IO.Writelnf("<fg=green>✓</> CLI application \"%s\" was uninstalled succesfully.", name)

		return nil
	},
}
