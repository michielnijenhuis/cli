package cli

import (
	"errors"
	"fmt"
	"testing"

	Command "github.com/michielnijenhuis/cli/command"
	Input "github.com/michielnijenhuis/cli/input"
)

func TestApplicationCanRenderError(t *testing.T) {
	app := NewApplication("app", "v1.0.0")
	app.SetCatchErrors(true)
	app.SetAutoExit(false)

	cmd := Command.NewCommand("test", func (self *Command.Command) (int, error) {
		return 1, errors.New("Test error")
	})

	app.Add(cmd)

	argv := map[string]Input.InputType{
		"command": "test",
	}

	input, _ := Input.NewObjectInput(argv, nil)
	parseError := input.Validate()

	if parseError != nil {
		fmt.Print(parseError.Error())
	}

	code, err := app.Run(input, nil)

	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}

	if code != 1 {
		t.Errorf("Expected code 1, got: %d", code)
	}

	if errMsg != "Test error" {
		t.Errorf("Expected error \"Test error\", got: %s", errMsg)
	}
}