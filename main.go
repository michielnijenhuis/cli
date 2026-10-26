package main

import (
	"log"

	"github.com/michielnijenhuis/cli/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
