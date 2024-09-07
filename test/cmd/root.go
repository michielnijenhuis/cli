package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/michielnijenhuis/cli"
)

func Execute() {
	autoExit := true

	app := &cli.Application{
		Name:        "app",
		Version:     "v1.0.0",
		CatchErrors: true,
		AutoExit:    autoExit,
		Commands: []*cli.Command{
			{
				Name:        "test",
				Aliases:     []string{"t"},
				Description: "Test command",
				Run:         test,
			},
		},
	}

	if err := app.Run(); err != nil {
		log.Fatalln(err)
	}
}

func test(c *cli.Ctx) {
	interval := time.Millisecond * 75
	frame := cli.SpinnerFrame{Frames: cli.DotSpinner}

	c.Output.Box("", "some\nmultiline\nmessage", "", "gray", "")

	v := c.View()

	v.HideCursor()
	defer v.ShowCursor()

	start := time.Now()
	duration := time.Duration(0)
	cp := c.ChildProcess("sleep 3")

	success := c.WithGracefulExit(func(done <-chan bool) {
		finished := make(chan bool)

		go func(cp *cli.ChildProcess) {
			cp.Run()
			duration = time.Since(start)
			finished <- true
		}(cp)

		for {
			select {
			case <-done:
			case <-finished:
				return
			default:
				v.RenderLinef("<fg=cyan>%s</> Sleeping...", frame.Next())
				time.Sleep(interval)
			}
		}
	})

	if success {
		v.RenderLinef("<fg=green>%s</> Done %s", cli.IconTick, cli.Dim(fmt.Sprintf("(took %.1f seconds)", float64(duration.Milliseconds())/1000)))
	} else {
		cp.Kill()
		v.RenderLinef("<fg=bright-red>%s</> Cancelled", cli.IconCross)
	}
}
