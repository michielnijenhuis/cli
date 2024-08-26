package cmd

import (
	"github.com/michielnijenhuis/cli"
)

func Execute() {
	// cmd := &cli.Command{
	// 	Name:        "test",
	// 	Description: "This is a test command",
	// 	Help:        "Show some help information",
	// 	Handle: func(c *cli.Command) (int, error) {
	// 		o := c.Output()

	// 		headers := []string{"", "foo", "bar", "baz"}
	// 		rows := [][]*cli.TableCell{
	// 			{
	// 				cli.NewTableCell("<fg=red>!</> row 1"),
	// 				cli.NewTableCell(" !  value 1"),
	// 				cli.NewTableCell("value 2"),
	// 				cli.NewTableCell("value <fg=cyan>3</>?!?"),
	// 				// cli.NewTableCell("value 3"),
	// 			},
	// 			{
	// 				cli.NewTableCell("<fg=red>!</> row 2"),
	// 				cli.NewTableCell("valuee 4"),
	// 				cli.NewTableCell("<accent>valueeee</accent> 5"),
	// 				// cli.NewTableCell("value 5"),
	// 				cli.NewTableCell("value   6"),
	// 			},
	// 			{
	// 				cli.NewTableCell("<fg=red>!</> row 3"),
	// 				cli.NewTableCell("<options=underscore>vaaaaaalue</> 7"),
	// 				// cli.NewTableCell("value 7"),
	// 				cli.NewTableCell("valllllllllue 8"),
	// 				cli.NewTableCell("value 999"),
	// 			},
	// 		}

	// 		table := cli.NewTable(o)
	// 		table.SetHeaders(headers)
	// 		table.SetRows(rows)
	// 		table.SetStyleByName("box")

	// 		o.NewLine(1)
	// 		table.Render()

	// 		return 0, nil
	// 	},
	// }

	app := &cli.Application{
		Name:    "app",
		Version: "v1.0.0",
		// CatchErrors: true,
	}

	cmd, err := cli.NewCommand(`signature|sig : Test sig command
									{--Q|queue= : Some option description}
									{user?* : The name of the user}`,
		func(c *cli.Command) (int, error) {
			c.Exec("echo \"Hello World!\"", "zsh", true)
			return 0, nil
		})

	if err != nil {
		panic(err)
	}

	app.Add(cmd)
	app.Run()
}
