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
		Name:           "app",
		Description:    "Beautiful CLI application",
		Version:        "v1.0.0",
		AutoExit:       true,
		CatchErrors:    true,
		PromptForInput: true,
		Arguments: []cli.Arg{
			&cli.StringArg{
				Name:        "name",
				Description: "Name of the person",
				Options:     []string{"Michiel", "Michael", "Miguel"},
				Required:    true,
			},
		},
		Run: func(io *cli.IO) {
			answer, err := io.Search("What is your name?", func(s string) cli.SearchResult {
				list := make([]string, 0, 26)
				for i := 'a'; i <= 'z'; i++ {
					list = append(list, string(i))
				}
				return list
				m := map[string]string{
					"Michiel": "Halouminator",
					"Michael": "Mperor",
					"Miguel":  "Muzzul",
				}

				if s == "" {
					return m
				}

				s = strings.ToLower(s)
				for k, v := range m {
					v = strings.ToLower(v)
					if !strings.Contains(v, s) {
						delete(m, k)
					}
				}

				return m
			}, "")

			if err != nil {
				io.Err(err)
				return
			}

			io.Success(fmt.Sprintf("Hello %s", answer))
		},
	}

	if err := app.Execute(); err != nil {
		log.Fatalln(err)
	}
}
