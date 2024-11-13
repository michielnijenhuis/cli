package cmd

import (
	"fmt"
	"log"

	"github.com/michielnijenhuis/cli"
)

func Execute() {
	childA := &cli.Command{
		Name:        "child-a",
		Description: "Child command A",
		Run: func(io *cli.IO) {
			tmux := io.Tmux()
			ses := "bro"
			cmds := []string{
				tmux.NewDetachedSessionCommand(ses),
				tmux.RenameWindowCommand(ses, "1", "bruh"),
				tmux.NewWindowCommand(ses, "brah"),
				tmux.NewWindowCommand(ses, "broseph"),
				tmux.NewWindowCommand(ses, "broheim"),
				tmux.SplitWindowHorizontallyCommand(ses, "bruh"),
				tmux.SplitWindowHorizontallyCommand(ses, "brah"),
				tmux.SplitWindowHorizontallyCommand(ses, "broseph"),
				tmux.SelectWindowCommand(ses, "bruh"),
				tmux.SelectPaneCommand(ses, "bruh", 1),
				tmux.SendKeysCommand(ses, "bruh", "echo 'Hello, bruh!'"),
				tmux.AttachSessionCommand(ses),
			}

			tmux.ExecMultiple(cmds)
		},
	}

	childA2 := &cli.Command{
		Name:        "some-long-command",
		Aliases:     []string{"s"},
		Description: "A beautiful and brilliant command",
		Run: func(io *cli.IO) {
			io.Warn("This command has a long name")
		},
	}

	childB := &cli.Command{
		Name:        "child-b",
		Description: "Child command B",
		Run: func(io *cli.IO) {
			io.Info("Hello from child C")
		},
	}

	childC := &cli.Command{
		Name:        "child-c",
		Description: "Child command C",
		Arguments: []cli.Arg{
			&cli.StringArg{
				Name:        "name",
				Description: "Name of the person",
				Required:    true,
				Options:     []string{"michiel", "michiel2"},
			},
		},
		Run: func(io *cli.IO) {
			name := io.String("name")
			io.Info(fmt.Sprintf("Hello from child B, %s", name))
		},
	}

	childD := &cli.Command{
		Name:        "offspring-d",
		Description: "Child command D",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "name",
				Description: "Name of the person",
				Shortcuts:   []string{"n"},
				Options:     []string{"foo", "barr"},
			},
		},
		Arguments: []cli.Arg{
			&cli.StringArg{
				Name:        "test",
				Description: "bro",
				Required:    true,
				Options:     []string{"broseph", "broheim"},
			},
			&cli.StringArg{
				Name:        "test-2",
				Description: "bro-2",
				// Required:    true,
				// Options: []string{"brosephino", "brochacho"},
			},
			&cli.ArrayArg{
				Name:        "test-3",
				Description: "bro-3",
				Options:     []string{"brah", "bro", "bruh", "breh", "bruuuh"},
			},
		},
		NativeFlags: []string{"help"},
		Run: func(io *cli.IO) {
			name := io.String("name")
			test := io.String("test")
			io.Info(fmt.Sprintf("Hello from offspring d, %s, %s", name, test))
		},
	}

	childA.AddCommand(childB)
	childB.AddCommand(childC)
	childB.AddCommand(childD)

	rootCmd := &cli.Command{
		Name:               "omg",
		Description:        "Beautiful CLI application",
		Version:            "v1.0.0",
		AutoExit:           true,
		CatchErrors:        true,
		NativeFlags:        []string{"help", "version", "quiet"},
		CascadeNativeFlags: true,
		Run: func(io *cli.IO) {
			io.Writeln("test")
		},
	}

	rootCmd.AddCommand(childA)
	rootCmd.AddCommand(childA2)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
