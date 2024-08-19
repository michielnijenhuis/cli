package cmd

import (
	"github.com/michielnijenhuis/cli"
)

func Execute() {
	cmd := &cli.Command{
		Name:        "testing",
		Description: "This is a test command",
		Help:        "Show some help information",
		Handle: func(c *cli.Command) (int, error) {
			rows := []map[string]any{
				{
					"":    "row 1",
					"foo": "bar",
					"bar": "baz",
					"baz": "yoink",
				},
				{
					"":    "row 2",
					"foo": "bruh",
					"bar": "broseph",
					"baz": "brah",
				},
			}

			o := c.Output()

			style := cli.NewTableStyle("box")
			style.CellRowContentFormat = "<primary> %s </primary>"

			headers := []string{"", "foo", "bar", "baz"}
			table := o.CreateTable(headers, o.CreateTableRowsFromMaps(headers, rows), &cli.TableOptions{
				Style:       "box-double",
				HeaderTitle: "header",
				FooterTitle: "footer",
				// Align:       "center",
				// DisplayOrientation: cli.DisplayOrientationHorizontal,
			})
			table.SetColumnStyle(0, style)

			table.Render()
			o.NewLine(1)

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
