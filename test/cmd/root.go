package cmd

import (
	"github.com/michielnijenhuis/cli"
)

func Execute() {
	app := &cli.Application{
		Name:        "app",
		Version:     "v1.0.0",
		CatchErrors: true,
		Commands: []*cli.Command{
			{
				Name:        "test",
				Aliases:     []string{"t"},
				Description: "Test command",
				Run:         test,
			},
		},
	}

	app.RunExit()
}

func test(c *cli.Ctx) {
	// cp := exec.Command("zsh", "-i", "-c", "wd c && pwd")
	// cp := exec.Command("zsh", "-i", "-c", "wd c && echo \"Dir: $(pwd)\"")
	// cp.Stdin = os.Stdin
	// cp.Stdout = os.Stdout
	// cp.Stderr = os.Stderr
	// cp.Run()
	c.Spinner(func() {
		c.Zsh("wd c && pwd")
	}, "Processing...")
}
