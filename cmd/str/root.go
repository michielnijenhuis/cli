package main

import (
	"log"

	"github.com/michielnijenhuis/cli"
)

var RootCmd = &cli.Command{
	Name:        "str",
	Description: "String helpers",
	Version:     "1.0.0",
	AutoExit:    true,
}

func init() {
	subCommands := []*cli.Command{
		CamelCmd,
		CapitalizeCmd,
		CharAtCmd,
		ContainsCmd,
		EndsWithCmd,
		JoinCmd,
		KebabCmd,
		LenCmd,
		LowerCmd,
		PadEndCmd,
		PadStartCmd,
		PascalCmd,
		RepeatCmd,
		ReplaceCmd,
		SliceCmd,
		SnakeCmd,
		StartsWithCmd,
		TrimCmd,
		UpperCmd,
	}

	for _, cmd := range subCommands {
		err := RootCmd.AddCommand(cmd)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
