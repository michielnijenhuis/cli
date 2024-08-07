package cmd

import (
	"fmt"
	"strings"

	"github.com/michielnijenhuis/cli"
)

func Execute() {
	cmd := &cli.Command{
		Name:        "testing",
		Description: "This is a test command",
		Help:        "Show some help information",
		Handle: func(c *cli.Command) (int, error) {
			var branch string
			var err error

			c.Spinner(func() {
				// time.Sleep(1000 * time.Millisecond)
				cp := c.Spawn("wd c && cd cli-go && echo $(current_branch)", "zsh", false)
				branch, err = cp.Run()

			}, "")

			if err != nil {
				c.Err(err)
			} else {
				c.Comment(fmt.Sprintf("Branch: %s", strings.TrimSpace(branch)))
			}

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
