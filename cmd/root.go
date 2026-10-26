package cmd

import (
	cli "github.com/michielnijenhuis/cli/internal"
)

var RootCmd = &cli.Command{
	Name:        "cli",
	Description: "A CLI toolkit",
	Aliases:     []string{"cli"},
	Version:     "1.0.0",
	AutoExit:    true,
}

func init() {
	subCommands := []*cli.Command{
		RunCmd,
		InstallCmd,
		UninstallCmd,
		PrintCmd,
		ArgCmd,
		FlagCmd,
	}

	for _, cmd := range subCommands {
		err := RootCmd.AddCommand(cmd)
		if err != nil {
			panic(err)
		}
	}
}
