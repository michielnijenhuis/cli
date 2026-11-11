package main

import (
	"log"

	"github.com/michielnijenhuis/cli"
)

var RootCmd = &cli.Command{
	Name:            "cli",
	Description:     "A CLI toolkit",
	Version:         cli.Version,
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
		CursorCmd,
	}

	for _, cmd := range subCommands {
		err := RootCmd.AddCommand(cmd)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
