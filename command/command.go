package command

import (
	"fmt"
	"regexp"
	"strings"

	Error "github.com/michielnijenhuis/cli/error"
	Helper "github.com/michielnijenhuis/cli/helper"
	Input "github.com/michielnijenhuis/cli/input"
	Output "github.com/michielnijenhuis/cli/output"
)

type CommandHandle func(self *Command) (int, error)

type Command struct {
	handle CommandHandle

	name        string
	description string
	aliases     []string
	help        string

	// application *Application.Application
	definition             *Input.InputDefinition
	hidden                 bool
	fullDefinition         *Input.InputDefinition
	ignoreValidationErrors bool
	synopsis               map[string]string
	usages                 []string
	helperSet              *Helper.HelperSet
	input                  Input.InputInterface
	output                 Output.OutputInterface
}

func NewCommand(name string, handle CommandHandle) *Command {
	c := &Command{
		handle: handle,

		name:                   name,
		description:            "",
		aliases:                make([]string, 0),
		help:                   "",
		definition:             Input.NewInputDefinition(nil, nil),
		hidden:                 false,
		fullDefinition:         nil,
		ignoreValidationErrors: false,
		synopsis:               make(map[string]string),
		usages:                 make([]string, 0),
		helperSet:              nil,
		input:                  nil,
		output:                 nil,
	}

	c.configure()

	if name != "" {
		aliases := strings.Split(name, "|")
		name = aliases[0]
		c.SetAliases(aliases[1:])
	}

	if name != "" {
		c.SetName(name)
	}

	return c
}

func (c *Command) IgnoreValidationErrors() {
	c.ignoreValidationErrors = true
}

func (c *Command) MergeApplication(helperSet *Helper.HelperSet, definition *Input.InputDefinition, mergeArgs bool) {
	c.helperSet = helperSet

	fullDefinition := Input.NewInputDefinition(nil, nil)
	fullDefinition.SetOptions(c.definition.GetOptionsArray())

	if definition != nil {
		fullDefinition.AddOptions(definition.GetOptionsArray())
	}

	if mergeArgs {
		if definition != nil {
			fullDefinition.SetArguments(definition.GetArgumentsArray())
		}
		fullDefinition.AddArguments(c.definition.GetArgumentsArray())
	} else {
		fullDefinition.SetArguments(c.definition.GetArgumentsArray())
	}

	c.fullDefinition = fullDefinition
}

func (c *Command) SetHelperSet(helperSet *Helper.HelperSet) {
	c.helperSet = helperSet
}

func (c *Command) GetHelperSet() *Helper.HelperSet {
	return c.helperSet
}

func (c *Command) IsEnabled() bool {
	return true
}

func (c *Command) configure() {}

func (c *Command) execute(input Input.InputInterface, output Output.OutputInterface) (int, error) {
	c.input = input
	c.output = output // TODO: Style

	exitCode, err := c.handle(c)
	if err != nil {
		_, manuallyFailed := err.(Error.ManuallyFailedError)
		if manuallyFailed {
			output.Writeln(err.Error(), 0)
			return 1, err
		}
	}

	return exitCode, err
}

func (c *Command) Fail(err string) (int, error) {
	return 1, Error.NewManuallyFailedError(err)
}

func (c *Command) StringArgument(name string) (string, error) {
	return c.input.GetStringArgument(name)
}

func (c *Command) ArrayArgument(name string) ([]string, error) {
	return c.input.GetArrayArgument(name)
}

func (c *Command) Arguments() map[string]Input.InputType {
	return c.input.GetArguments()
}

func (c *Command) BoolOption(name string) (bool, error) {
	return c.input.GetBoolOption(name)
}

func (c *Command) StringOption(name string) (string, error) {
	return c.input.GetStringOption(name)
}

func (c *Command) ArrayOption(name string) ([]string, error) {
	return c.input.GetArrayOption(name)
}

func (c *Command) Options() map[string]Input.InputType {
	return c.input.GetOptions()
}

func (c *Command) interact(input Input.InputInterface, output Output.OutputInterface) {}

func (c *Command) initialize(input Input.InputInterface, output Output.OutputInterface) {}

func (c *Command) Run(input Input.InputInterface, output Output.OutputInterface) (int, error) {
	if input == nil {
		argvInput, err := Input.NewArgvInput(nil, nil)
		if err != nil {
			return 1, err
		}
		input = argvInput
	}

	if output == nil {
		output = Output.NewConsoleOutput(0, true, nil)
	}

	err := input.Bind(c.GetDefinition())
	if err != nil && !c.ignoreValidationErrors {
		return 1, err
	}

	c.initialize(input, output)

	if input.IsInteractive() {
		c.interact(input, output)
	}

	if input.HasArgument("command") {
		command, _ := input.GetStringArgument("command")
		if command == "" {
			input.SetArgument("command", c.GetName())
		}
	}

	validationErr := input.Validate()

	if validationErr != nil {
		return 1, validationErr
	}

	return c.execute(input, output)
}

func (c *Command) SetDefinition(definition *Input.InputDefinition) {
	c.definition = definition
	c.fullDefinition = nil
}

func (c *Command) GetDefinition() *Input.InputDefinition {
	if c.fullDefinition == nil {
		return c.GetNativeDefinition()
	}

	return c.fullDefinition
}

func (c *Command) GetNativeDefinition() *Input.InputDefinition {
	if c.definition == nil {
		panic("Command is not correctly initialized. Create a new command using the \"NewCommand()\" function.")
	}

	return c.definition
}

func (c *Command) AddArgument(name string, mode uint, description string, defaultValue Input.InputType, validator Input.InputValidator) *Command {
	return c
}

func (c *Command) AddOption(name string, shortcut []string, mode uint, description string, defaultValue Input.InputType, validator Input.InputValidator) *Command {
	return c
}

func (c *Command) SetName(name string) *Command {
	c.validateName(name)
	c.name = name
	return c
}

func (c *Command) GetName() string {
	if c.name == "" {
		return "CLI Tool"
	}
	return c.name
}

func (c *Command) SetHidden(hidden bool) *Command {
	c.hidden = hidden
	return c
}

func (c *Command) IsHidden() bool {
	return c.hidden
}

func (c *Command) SetDescription(description string) *Command {
	c.description = description
	return c
}

func (c *Command) GetDescription() string {
	return c.description
}

func (c *Command) SetHelp(help string) *Command {
	c.help = help
	return c
}

func (c *Command) GetHelp() string {
	return c.help
}

func (c *Command) SetAliases(aliases []string) *Command {
	for _, alias := range aliases {
		c.validateName(alias)
	}

	c.aliases = aliases
	return c
}

func (c *Command) GetAliases() []string {
	return c.aliases
}

func (c *Command) GetSynopsis(short bool) string {
	key := "long"
	if short {
		key = "short"
	}

	if c.synopsis[key] == "" {
		c.synopsis[key] = strings.TrimSpace(fmt.Sprintf("%s %s", c.name, c.definition.GetSynopsis(short)))
	}

	return c.synopsis[key]
}

func (c *Command) AddUsage(usage string) *Command {
	if !strings.HasPrefix(usage, c.name) {
		usage = c.name + " " + usage
	}

	c.usages = append(c.usages, usage)

	return c
}

func (c *Command) GetUsages() []string {
	return c.usages
}

func (c *Command) GetHelper(name string) (Helper.HelperInterface, error) {
	if c.helperSet == nil {
		return nil, Error.NewInvalidArgumentError("Cannot retrieve helper because there is no HelperSet defined. Did you forget to add your command to the application or to set the application on the command using the MergeApplication() method? You can also set the HelperSet directly using the SetHelperSet() method.")
	}

	return c.helperSet.Get(name)
}

func (c *Command) GetProcessedHelp() string {
	return "TODO: Command.GetProcessedHelp()"
}

func (c *Command) validateName(name string) {
	re := regexp.MustCompile("^[^:]+(:[^:]+)*")
	if !re.MatchString(name) {
		panic(fmt.Sprintf("Command name \"%s\" is invalid.", name))
	}
}
