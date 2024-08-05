package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/michielnijenhuis/cli"
)

func Execute() {
	cmd := &cli.Command{
		Name:        "test",
		Description: "This is a test command",
		Help:        "Show some help information",
		Handle: func(self *cli.Command) (code int, err error) {
			cmd := "echo Hello!"
			sh := "sh"

			arg, err := self.StringArgument("count")
			if err != nil {
				return 1, err
			}

			count, err := strconv.Atoi(arg)
			if err != nil {
				return 1, err
			}

			channel := make(chan string)
			counter := 0
			for i := 0; i < count; i++ {
				go func() {
					out, e := self.Exec(cmd, sh, false)

					if e != nil {
						err = e
						code = 1
					}

					channel <- strings.TrimSuffix(out, "\n")
					counter++

					if counter >= count {
						close(channel)
					}
				}()
			}

			o := self.Output()
			for x := range channel {
				o.Writeln(x, 0)
			}

			return code, err
		},
	}

	cmd.AddArgument(&cli.InputArgument{
		Name:         "count",
		Description:  "The amount of times to print",
		DefaultValue: "5",
		Validator: func(value cli.InputType) error {
			_, ok := value.(int)
			if ok {
				return nil
			}
			str, ok := value.(string)
			if ok {
				_, err := strconv.Atoi(str)
				if err == nil {
					return nil
				}
			}
			return fmt.Errorf("expected an integer, received \"%s\"", str)
		},
	})

	app := &cli.Application{
		Name:        "app",
		Version:     "v1.0.0",
		CatchErrors: true,
		AutoExit:    true,
	}

	app.Add(cmd)
	app.Run()
}
