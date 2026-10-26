package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cli "github.com/michielnijenhuis/cli/internal"
)

var InstallCmd = &cli.Command{
	Name:        "install",
	Description: "Install a command as defined in a yaml file as an executable",
	Aliases:     []string{"i"},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "force",
			Shortcuts:   []string{"f"},
			Description: "Overwrite the executable if it already exists",
		},
	},
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "config",
			Description: "The .yml/.yaml configuration file for the CLI app or command",
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

		configAbsPath, err := filepath.Abs(config)
		if err != nil {
			return fmt.Errorf("failed to determine absolute path: %w", err)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to determine home directory: %w", err)
		}

		binDir := filepath.Join(homeDir, ".local", "bin")
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			return fmt.Errorf("failed to create %q: %w", binDir, err)
		}

		app, err := cli.ParseApp(config)

		script := ctx.IO.String("script")

		if script == "" {
			script = strings.TrimSuffix(configAbsPath, filepath.Ext(configAbsPath)) + ".sh"
		} else {
			scriptAbsPath, err := filepath.Abs(script)
			if err != nil {
				return fmt.Errorf("failed to determine absolute path: %w", err)
			}

			script = scriptAbsPath
		}

		commandPath := filepath.Join(binDir, app.Name)
		_, err = os.Stat(commandPath)
		commandExists := err == nil

		if commandExists {
			if !ctx.IO.Bool("force") {
				return fmt.Errorf("file %s already exists", commandPath)
			} else if err = os.Remove(commandPath); err != nil {
				return fmt.Errorf("failed to remove %q: %w", commandPath, err)
			}
		}

		var commandFmt = `
#!/usr/bin/env bash

set -e

cli run %s %s "$@"
`

		commandScript := fmt.Sprintf(strings.TrimSpace(commandFmt)+"\n", configAbsPath, script)

		err = os.WriteFile(commandPath, []byte(commandScript), 0o755)
		if err != nil {
			return err
		}

		err = os.Chmod(commandPath, 0o755)
		if err != nil {
			return err
		}

		ctx.IO.Writelnf("<fg=green>✓</> CLI application \"%s\" installed succesfully.", app.Name)
		ctx.IO.Writelnf("<fg=gray>  * Config path:  %s</>", configAbsPath)
		ctx.IO.Writelnf("<fg=gray>  * Script path:  %s</>", script)
		ctx.IO.Writelnf("<fg=gray>  * Command path: %s</>", commandPath)
		ctx.IO.NewLine(1)
		ctx.IO.Writelnf("<fg=cyan>Run <fg=bright-white>$ %s</> in your terminal to try it out!", app.Name)

		return nil
	},
}
