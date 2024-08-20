package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestApplicationCanRenderError(t *testing.T) {
	app := &Application{
		Name:        "app",
		CatchErrors: true,
	}

	cmd := &Command{
		Name: "test",
		Handle: func(self *Command) (int, error) {
			return 1, errors.New("Test error")
		},
	}

	app.Add(cmd)

	code, err := app.Run("test")

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
	app := &Application{
		Name:        "app",
		CatchErrors: true,
	}

	cmd := &Command{
		Name: "test",
		Handle: func(self *Command) (int, error) {
			return 0, nil
		},
	}

	app.Add(cmd)

	code, err := app.Run("test")

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
	app := &Application{
		Name:        "app",
		CatchErrors: true,
	}

	expectedError := "Oh no!"

	cmd := &Command{
		Name: "test",
		Handle: func(self *Command) (int, error) {
			panic(expectedError)
		},
	}

	app.Add(cmd)

	code, err := app.Run("test")

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
	app := &Application{
		Name:        "app",
		CatchErrors: true,
	}

	if _, err := app.Run("help"); err != nil {
		t.Error(err.Error())
	}
}

func TestApplicationCanListCommands(t *testing.T) {
	app := &Application{
		Name:        "app",
		CatchErrors: true,
	}

	if _, err := app.Run("list"); err != nil {
		t.Error(err.Error())
	}
}

func TestSumCommand(t *testing.T) {
	cmd := &Command{
		Name: "sum",
		Handle: func(self *Command) (int, error) {
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

	cmd.AddArgument(&InputArgument{
		Name:        "values",
		Mode:        InputArgumentIsArray | InputArgumentRequired,
		Description: "The values to sum",
		Validator: func(value InputType) error {
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
		},
	})

	app := &Application{
		Name:        "app",
		CatchErrors: true,
	}

	app.Add(cmd)

	if _, err := app.Run("sum 1 2 3 4"); err != nil {
		t.Fatal(err.Error())
	}
}

func TestHelpCommandCanShowHelp(t *testing.T) {
	app := &Application{
		Name: "app",
	}

	cmd := &Command{
		Name:        "test",
		Description: "This is a test command that does nothing.",
		Help:        "Very useful help message.",
		Handle: func(self *Command) (int, error) {
			return 0, nil
		},
	}

	app.Add(cmd)

	if _, err := app.Run("help test"); err != nil {
		t.Error(err.Error())
	}
}

func TestCommandHasHelpFlag(t *testing.T) {
	app := &Application{
		Name: "app",
	}

	cmd := &Command{
		Name:        "test",
		Description: "This is a test command that does nothing.",
		Help:        "Very useful help message.",
		Handle: func(self *Command) (int, error) {
			return 0, nil
		},
	}

	app.Add(cmd)

	if _, err := app.Run("test -h"); err != nil {
		t.Error(err.Error())
	}
}

func TestCanSuggestAlternatives(t *testing.T) {
	app := &Application{
		Name: "app",
	}

	cmd := &Command{
		Name:        "test",
		Description: "This is a test command that does nothing.",
		Help:        "Very useful help message.",
		Handle: func(self *Command) (int, error) {
			return 0, nil
		},
	}

	app.Add(cmd)

	if _, err := app.Run("testt"); err != nil {
		if !strings.HasPrefix(err.Error(), "command \"testt\" is not defined") {
			t.Error(err)
		}
	} else {
		t.Error("Expected an error")
	}
}

func TestCommandCanExecChildProcesses(t *testing.T) {
	var out string
	expected := "Hello, world!"

	cmd := &Command{
		Name:        "test",
		Description: "This is a test command that does nothing.",
		Help:        "Very useful help message.",
		Handle: func(self *Command) (int, error) {
			var err error
			out, err = self.Exec("echo 'Hello, world!'", "", false)
			if err != nil {
				return 1, err
			}
			return 0, nil
		},
	}

	input := NewInput([]string{}...)

	if _, err := cmd.RunWith(input, nil); err != nil {
		t.Error(err.Error())
		return
	}

	out = strings.TrimSuffix(out, "\n")

	if out != expected {
		t.Errorf("Expected \"%s\", got \"%s\"", expected, out)
	}
}
