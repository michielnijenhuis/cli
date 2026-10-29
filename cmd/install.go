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
		&cli.BoolFlag{
			Name:        "check",
			Shortcuts:   []string{"c"},
			Description: "Only check if the command has been installed already.",
		},
	},
	Arguments: []cli.Arg{
		&cli.StringArg{
			Name:        "config",
			Description: "The .y(a)ml configuration file for the CLI app or command",
		},
		&cli.StringArg{
			Name:        "script",
			Description: "The .sh script that provides the command implementations",
		},
	},
	RunE: func(args []string, ctx *cli.CommandCtx) error {
		check := ctx.IO.Bool("check")
		config := ctx.IO.String("config")
		script := ctx.IO.String("script")

		if config == "" {
			currentDir, err := os.Getwd()
			if err != nil {
				return err
			}

			ymlFiles, err := filepath.Glob(currentDir + "/*.yml")
			if err != nil {
				return err
			}

			if len(ymlFiles) == 0 {
				ymlFiles, err = filepath.Glob(currentDir + "/*.yaml")
			}

			if err != nil {
				return err
			}

			if len(ymlFiles) == 0 {
				return fmt.Errorf("a .y(a)ml configuration file is required.")
			}

			if len(ymlFiles) > 1 {
				return fmt.Errorf("multiple .y(a)ml files found. Please provide a configuration file yourself.")
			}

			config = ymlFiles[0]
		}

		if config == "" {
			return fmt.Errorf("a .y(a)ml configuration file is required.")
		} else if err := cli.IsFile(config); err != nil {
			return err
		}

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
		if err != nil {
			if ctx.IO.IsVerbose() {
				return fmt.Errorf("invalid configuration file:\n%w", err)
			}

			return fmt.Errorf("invalid configuration file")
		}

		if script == "" {
			script = strings.TrimSuffix(configAbsPath, filepath.Ext(configAbsPath)) + ".sh"
		}

		if err = cli.IsFile(script); err != nil {
			return err
		}

		scriptAbsPath, err := filepath.Abs(script)
		if err != nil {
			return fmt.Errorf("failed to determine absolute path: %w", err)
		}

		script = scriptAbsPath

		commandPath := filepath.Join(binDir, app.Name)
		_, err = os.Stat(commandPath)
		commandExists := err == nil

		if commandExists {
			if check {
				os.Exit(0)
			}

			if !ctx.IO.Bool("force") {
				return fmt.Errorf("file %s already exists", commandPath)
			} else if err = os.Remove(commandPath); err != nil {
				return fmt.Errorf("failed to remove %q: %w", commandPath, err)
			}
		} else if check {
			os.Exit(1)
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

		if err = os.Chmod(script, 0o755); err != nil {
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
