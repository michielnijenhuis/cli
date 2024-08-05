package cmd

import (
	"github.com/michielnijenhuis/cli"
)

func Execute() {
	cmd := &cli.Command{
		Name:        "test",
		Description: "This is a test command",
		Help:        "Show some help information",
		Handle: func(self *cli.Command) (code int, err error) {
			o := self.Output()

			a, e := o.AskHidden("does this work?", nil)

			if e != nil {
				panic(e)
			}

			o.Writelnf("a: %s%s", 0, a, "!")

			return 0, nil
		},
	}

	app := &cli.Application{
		Name:        "app",
		Version:     "v1.0.0",
		CatchErrors: true,
		AutoExit:    true,
	}

	app.Add(cmd)
	app.Run()
}
