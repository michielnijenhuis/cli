package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/michielnijenhuis/cli/application"
	"github.com/michielnijenhuis/cli/command"
	"github.com/michielnijenhuis/cli/input"
)

func TestApplicationCanRenderError(t *testing.T) {
	app := &application.Application{
		Name:        "app",
		CatchErrors: true,
	}

	cmd := &command.Command{
		Name: "test",
		Handle: func(self *command.Command) (int, error) {
			return 1, errors.New("Test error")
		},
	}

	app.Add(cmd)

	input := input.Make("test")

	code, err := app.RunWith(input, nil)

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
	app := &application.Application{
		Name:        "app",
		CatchErrors: true,
	}

	cmd := &command.Command{
		Name: "test",
		Handle: func(self *command.Command) (int, error) {
			return 0, nil
		},
	}

	app.Add(cmd)

	input := input.Make("test")
	code, err := app.RunWith(input, nil)

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
	app := &application.Application{
		Name:        "app",
		CatchErrors: true,
	}

	expectedError := "Oh no!"

	cmd := &command.Command{
		Name: "test",
		Handle: func(self *command.Command) (int, error) {
			panic(expectedError)
		},
	}

	app.Add(cmd)

	input := input.Make("test")

	code, err := app.RunWith(input, nil)

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
	app := &application.Application{
		Name:        "app",
		CatchErrors: true,
	}

	input := input.Make("help")

	if _, err := app.RunWith(input, nil); err != nil {
		t.Error(err.Error())
	}
}

func TestApplicationCanListCommands(t *testing.T) {
	app := &application.Application{
		Name:        "app",
		CatchErrors: true,
	}

	input := input.Make("list")
	parseError := input.Validate()

	if parseError != nil {
		fmt.Print(parseError.Error())
	}

	if _, err := app.RunWith(input, nil); err != nil {
		t.Error(err.Error())
	}
}

func TestSumCommand(t *testing.T) {
	cmd := &command.Command{
		Name: "sum",
		Handle: func(self *command.Command) (int, error) {
			values, err := self.ArrayArgument("values")
			if err != nil {
				return 1, err
			}

			var sum int
			for _, v := range values {
				number, err := strconv.Atoi(v)
				if err != nil {
					return 1, err
				}

				sum += number
			}

			msg := fmt.Sprintf("Sum: %d", sum)

			self.Output().Writeln(msg, 0)
			return 0, nil
		}}

	cmd.SetDescription("Prints the sum of all given values.")
	cmd.DefineArgument("values", input.InputArgumentIsArray, "The values to sum", nil, func(value input.InputType) error {
		arr, ok := value.([]string)
		if ok {
			for _, v := range arr {
				_, err := strconv.Atoi(v)
				if err != nil {
					return err
				}
			}

			return nil
		}

		return errors.New("Value is not an array.")
	})

	input := input.Make("sum", "-vvv", "1", "2", "3", "4")

	app := &application.Application{
		Name:        "app",
		CatchErrors: true,
	}
	app.Add(cmd)

	_, err := app.RunWith(input, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestHelpCommandCanShowHelp(t *testing.T) {
	app := &application.Application{
		Name: "app",
	}

	cmd := &command.Command{
		Name:        "test",
		Description: "This is a test command that does nothing.",
		Help:        "Very useful help message.",
		Handle: func(self *command.Command) (int, error) {
			return 0, nil
		},
	}

	input := input.Make("help", "test")

	app.Add(cmd)

	if _, err := app.RunWith(input, nil); err != nil {
		t.Error(err.Error())
	}
}

func TestCommandHasHelpFlag(t *testing.T) {
	app := &application.Application{
		Name: "app",
	}

	cmd := &command.Command{
		Name:        "test",
		Description: "This is a test command that does nothing.",
		Help:        "Very useful help message.",
		Handle: func(self *command.Command) (int, error) {
			return 0, nil
		},
	}

	input := input.Make("test", "-h")

	app.Add(cmd)

	if _, err := app.RunWith(input, nil); err != nil {
		t.Error(err.Error())
	}
}

func TestCanSuggestAlternatives(t *testing.T) {
	app := &application.Application{
		Name: "app",
	}

	cmd := &command.Command{
		Name:        "test",
		Description: "This is a test command that does nothing.",
		Help:        "Very useful help message.",
		Handle: func(self *command.Command) (int, error) {
			return 0, nil
		},
	}

	input := input.Make("testt")

	app.Add(cmd)

	if _, err := app.RunWith(input, nil); err != nil {
		t.Error(err.Error())
	}
}

func TestCommandCanExecChildProcesses(t *testing.T) {
	var out string
	expected := "Hello, world!"

	cmd := &command.Command{
		Name:        "test",
		Description: "This is a test command that does nothing.",
		Help:        "Very useful help message.",
		Handle: func(self *command.Command) (int, error) {
			var err error
			out, err = self.Exec("echo 'Hello, world!'", "", false)
			if err != nil {
				return 1, err
			}
			return 0, nil
		},
	}

	input, _ := input.NewArgvInput([]string{}, nil)

	if _, err := cmd.Run(input, nil); err != nil {
		t.Error(err.Error())
		return
	}

	out = strings.TrimSuffix(out, "\n")

	if out != expected {
		t.Errorf("Expected %v, got %v", expected, out)
	}
}
