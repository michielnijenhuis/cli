package cli

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/michielnijenhuis/cli/application"
	"github.com/michielnijenhuis/cli/command"
	"github.com/michielnijenhuis/cli/input"
)

func TestApplicationCanRenderError(t *testing.T) {
	fmt.Println("--- begin APPLICATION CAN RENDER test ---")

	app := application.NewApplication("app", "v1.0.0")
	app.SetCatchErrors(true)
	app.SetAutoExit(false)

	cmd := command.NewCommand("test", func(self *command.Command) (int, error) {

		return 1, errors.New("Test error")
	})

	app.Add(cmd)

	argv := map[string]input.InputType{
		"command": "test",
	}

	input, _ := input.NewObjectInput(argv, nil)
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

	fmt.Println("--- end APPLICATION CAN RENDER test ---")
	fmt.Println("")
}

func TestApplicationCanSuccesfullyExecuteCommand(t *testing.T) {
	fmt.Println("--- begin APPLICATION CAN SUCCESSFULLY EXECUTE COMMAND test ---")

	app := application.NewApplication("app", "v1.0.0")
	app.SetCatchErrors(true)
	app.SetAutoExit(false)

	cmd := command.NewCommand("test", func(self *command.Command) (int, error) {
		return 0, nil
	})

	app.Add(cmd)

	argv := map[string]input.InputType{
		"command": "test",
	}

	input, _ := input.NewObjectInput(argv, nil)
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

	fmt.Println("--- end APPLICATION CAN SUCCESSFULLY EXECUTE COMMAND test ---")
	fmt.Println("")
}

func TestApplicationCanRecover(t *testing.T) {
	fmt.Println("--- begin APPLICATION CAN RECOVER test ---")

	app := application.NewApplication("app", "v1.0.0")
	app.SetCatchErrors(true)
	app.SetAutoExit(false)

	expectedError := "Oh no!"

	cmd := command.NewCommand("test", func(self *command.Command) (int, error) {
		panic(expectedError)
	})

	app.Add(cmd)

	argv := map[string]input.InputType{
		"command": "test",
	}

	input, _ := input.NewObjectInput(argv, nil)
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

	fmt.Println("--- end APPLICATION CAN RECOVER test ---")
	fmt.Println("")
}

func TestApplicationCanShowHelp(t *testing.T) {
	fmt.Println("--- begin CAN SHOW HELP test ---")

	app := application.NewApplication("app", "v1.0.0")
	app.SetCatchErrors(true)
	app.SetAutoExit(false)

	argv := map[string]input.InputType{
		"command": "help",
	}

	input, _ := input.NewObjectInput(argv, nil)
	parseError := input.Validate()

	if parseError != nil {
		fmt.Print(parseError.Error())
	}

	app.Run(input, nil)
	fmt.Println("--- end CAN SHOW HELP test ---")
	fmt.Println("")
}

func TestApplicationCanListCommands(t *testing.T) {
	fmt.Println("--- begin CAN LIST COMMANDS test ---")
	app := application.NewApplication("app", "v1.0.0")
	app.SetCatchErrors(true)
	app.SetAutoExit(false)

	argv := map[string]input.InputType{
		"command": "list",
	}

	input, _ := input.NewObjectInput(argv, nil)
	parseError := input.Validate()

	if parseError != nil {
		fmt.Print(parseError.Error())
	}

	app.Run(input, nil)

	fmt.Println("--- end CAN LIST COMMANDS test ---")
	fmt.Println("")
}

func TestSumCommand(t *testing.T) {
	cmd := command.NewCommand("sum", func(self *command.Command) (int, error) {
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
	})
	cmd.SetDescription("Prints the sum of all given values.")
	cmd.AddArgument("values", input.INPUT_ARGUMENT_IS_ARRAY, "The values to sum", nil, func(value input.InputType) error {
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

	input, _ := input.NewObjectInput(map[string]input.InputType{
		"command":   "sum",
		"values":    []string{"1", "2", "3", "4"},
		"--verbose": "3",
	}, nil)

	app := application.NewApplication("app", "v1.0.0")
	app.SetCatchErrors(true)
	app.SetAutoExit(false)
	app.Add(cmd)

	_, err := app.Run(input, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
}
