package cmd

import (
	"time"

	"github.com/michielnijenhuis/cli"
)

func Execute() {
	cmd := &cli.Command{
		Name:        "testing",
		Description: "This is a test command",
		Help:        "Show some help information",
		Handle: func(self *cli.Command) (code int, err error) {
			cli.AddTheme("prompt", cli.Theme{
				Foreground: "red",
			})

			cli.AddTheme("info", cli.Theme{
				Background: "magenta",
			})

			self.Spinner(func() {
				time.Sleep(2000 * time.Millisecond)
			}, "Waiting...")

			self.NewLine(1)
			self.Ok("Done!")
			self.NewLine(1)
			self.Info("Done!")
			self.NewLine(1)
			self.Err("Done!")
			self.NewLine(1)
			self.Warn("Done!")
			self.NewLine(1)
			self.Alert("Done!")
			self.NewLine(1)
			self.Comment("Done!")

			return 0, nil
		},
	}

	app := &cli.Application{
		Name:        "app",
		Version:     "v1.0.0",
		CatchErrors: true,
		AutoExit:    true,
	}

	cli.SetBaseTheme("red", "cyan")

	app.Add(cmd)
	app.Run()
}
