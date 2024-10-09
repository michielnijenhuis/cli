package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/michielnijenhuis/cli"
)

func Execute() {
	app := &cli.Command{
		Name:           "app",
		Description:    "Beautiful CLI application",
		Version:        "v1.0.0",
		AutoExit:       true,
		CatchErrors:    true,
		PromptForInput: true,
		Arguments: []cli.Arg{
			&cli.StringArg{
				Name:        "name",
				Description: "Your name",
				Options:     []string{"Michiel", "Suus", "Anita"},
				Required:    true,
			},
			&cli.ArrayArg{
				Name:        "names",
				Description: "Your names",
				// Options:     []string{"Michiel", "Suus", "Anita"},
				Min: 1,
			},
		},
		Run: func(io *cli.IO) {
			selected, err := io.MultiSelect("What is your name?", map[string]string{
				"Nijenhuis":   "Michiel",
				"Rouwenhorst": "Suus",
				"Lesman":      "Anita",
			}, nil)

			if err != nil {
				io.Err(err)
				return
			}

			io.Info(fmt.Sprintf("Hello %s", strings.Join(selected, ", ")))
		},
	}

	if err := app.Execute(); err != nil {
		log.Fatalln(err)
	}
}
