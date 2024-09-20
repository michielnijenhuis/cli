package cmd

import (
	"github.com/michielnijenhuis/cli"
)

func Execute() {
	app := &cli.Application{
		Name:    "app",
		Version: "v1.0.0",
		// CatchErrors: true,
		Commands: []*cli.Command{
			{
				Name:        "test:ing",
				Description: "Test command",
				Run:         test,
				Arguments: []cli.Arg{
					&cli.StringArg{
						Name:        "arg",
						Description: "Argument",
						Required:    true,
					},
				},
			},
			{
				Name:        "test:bro",
				Aliases:     []string{"t"},
				Description: "Test command 2",
				Run:         test2,
			},
			{
				Name:        "test:sup:dude",
				Description: "Test command 4",
				Run:         test4,
				Arguments: []cli.Arg{
					&cli.StringArg{
						Name:        "foo",
						Description: "Foo?",
						Required:    true,
					},
				},
			},
			{
				Name:        "hell:naaaah",
				Description: "Test command 3",
				Run:         test3,
			},
		},
	}

	app.RunExit()
}

func test(io *cli.IO) {
	io.Warn(io.String("arg"))
}

func test2(io *cli.IO) {
	io.Warn("BRO! Wait, what?")
}

func test3(io *cli.IO) {
	io.Info("Praise Jesus")
}

func test4(io *cli.IO) {
	io.Info("Hail Satan: " + io.String("foo"))
}
