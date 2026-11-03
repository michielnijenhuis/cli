package main

import (
	"github.com/michielnijenhuis/cli"
)

var RootCmd = &cli.Command{
	Name:            "cli",
	Description:     "A CLI toolkit",
	Aliases:         []string{"cli"},
	Version:         "1.4.0",
	AutoExit:        true,
	WithNativeFlags: true,
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
