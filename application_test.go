package cli

import (
	"errors"
	"fmt"
	"testing"

	Application "github.com/michielnijenhuis/cli/application"
	Command "github.com/michielnijenhuis/cli/command"
	Input "github.com/michielnijenhuis/cli/input"
)

func TestApplicationCanRenderError(t *testing.T) {
	app := Application.NewApplication("app", "v1.0.0")
	app.SetCatchErrors(true)
	app.SetAutoExit(false)

	cmd := Command.NewCommand("test", func(self *Command.Command) (int, error) {

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

func TestApplicationCanSuccesfullyExecuteCommand(t *testing.T) {
	app := Application.NewApplication("app", "v1.0.0")
	app.SetCatchErrors(true)
	app.SetAutoExit(false)

	cmd := Command.NewCommand("test", func(self *Command.Command) (int, error) {
		return 0, nil
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

	if code != 0 {
		t.Errorf("Expected code 0, got: %d", code)
	}

	if errMsg != "" {
		t.Errorf("Expected no error, got: %s", errMsg)
	}
}

func TestApplicationCanRecover(t *testing.T) {
	app := Application.NewApplication("app", "v1.0.0")
	app.SetCatchErrors(true)
	app.SetAutoExit(false)

	expectedError := "Oh no!"

	cmd := Command.NewCommand("test", func(self *Command.Command) (int, error) {
		panic(expectedError)
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

	if errMsg != expectedError {
		t.Errorf("Expected error \"%s\", got: %s", expectedError, errMsg)
	}
}

func TestApplicationCanShowHelp(t *testing.T) {
	app := Application.NewApplication("app", "v1.0.0")
	app.SetCatchErrors(true)
	app.SetAutoExit(false)

	argv := map[string]Input.InputType{
		"command": "help",
	}

	input, _ := Input.NewObjectInput(argv, nil)
	parseError := input.Validate()

	if parseError != nil {
		fmt.Print(parseError.Error())
	}

	app.Run(input, nil)
}

func TestApplicationCanListCommands(t *testing.T) {
	app := Application.NewApplication("app", "v1.0.0")
	app.SetCatchErrors(true)
	app.SetAutoExit(false)

	argv := map[string]Input.InputType{
		"command": "list",
	}

	input, _ := Input.NewObjectInput(argv, nil)
	parseError := input.Validate()

	if parseError != nil {
		fmt.Print(parseError.Error())
	}

	app.Run(input, nil)
}
