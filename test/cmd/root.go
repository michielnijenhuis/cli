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
			io.Info("Hello from child A")
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
		Name:           "child-b",
		Description:    "Child command B",
		PromptForInput: true,
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
		Arguments: []cli.Arg{
			&cli.StringArg{
				Name:        "name",
				Description: "Name of the person",
				Required:    true,
				Options:     []string{"foo", "barr"},
			},
		},
		Run: func(io *cli.IO) {
			name := io.String("name")
			io.Info(fmt.Sprintf("Hello from child B, %s", name))
		},
	}

	childA.AddCommand(childB)
	childB.AddCommand(childC)
	childB.AddCommand(childD)

	rootCmd := &cli.Command{
		Name:           "bruh",
		Description:    "Beautiful CLI application",
		Version:        "v1.0.0",
		AutoExit:       true,
		CatchErrors:    true,
		PromptForInput: true,
		NativeFlags:    []string{"help"},
	}

	rootCmd.AddCommand(childA)
	rootCmd.AddCommand(childA2)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
