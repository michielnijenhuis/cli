package cmd

import (
	"github.com/michielnijenhuis/cli"
)

func Execute() {
	cmd := &cli.Command{
		Name:        "test",
		Description: "This is a test command",
		Help:        "Show some help information",
		Handle: func(c *cli.Command) (int, error) {
			o := c.Output()

			headers := []string{"", "foo", "bar", "baz"}
			rows := [][]*cli.TableCell{
				{
					cli.NewTableCell("<fg=red>!</> row 1"),
					cli.NewTableCell("value 1"),
					cli.NewTableCell("value 2"),
					cli.NewTableCell("value <fg=cyan>3</>"),
					// cli.NewTableCell("value 3"),
				},
				{
					cli.NewTableCell("<fg=red>!</> row 2"),
					cli.NewTableCell("value 4"),
					cli.NewTableCell("<accent>value</accent> 5"),
					// cli.NewTableCell("value 5"),
					cli.NewTableCell("value 6"),
				},
				{
					cli.NewTableCell("<fg=red>!</> row 3"),
					cli.NewTableCell("<options=underscore>value</> 7"),
					// cli.NewTableCell("value 7"),
					cli.NewTableCell("value 8"),
					cli.NewTableCell("value 9"),
				},
			}

			table := cli.NewTable(o)
			table.SetHeaders(headers)
			table.SetRows(rows)
			table.SetStyleByName("box")

			o.NewLine(1)
			table.Render()

			return 0, nil
		},
	}

	app := &cli.Application{
		Name:    "app",
		Version: "v1.0.0",
		// CatchErrors: true,
	}

	app.Add(cmd)
	app.Run()
}
