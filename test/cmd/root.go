package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/michielnijenhuis/cli"
)

type Foo struct {
	Name string
}

type Bar struct {
	*Foo
}

func Execute() {
	app := &cli.Command{
		Name:        "app",
		Description: "Beautiful CLI application",
		Version:     "v1.0.0",
		AutoExit:    true,
		CatchErrors: true,
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

			io.Success(fmt.Sprintf("Hello %s", strings.Join(selected, ", ")))
		},
	}

	if err := app.Execute(); err != nil {
		log.Fatalln(err)
	}
}
