package command

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	Error "github.com/michielnijenhuis/cli/error"
	Input "github.com/michielnijenhuis/cli/input"
	Output "github.com/michielnijenhuis/cli/output"
)

type CommandHandle func(self *Command) (int, error)
type CommandInitializer func(input Input.InputInterface, output Output.OutputInterface)
type CommandInteracter func(input Input.InputInterface, output Output.OutputInterface)

type Command struct {
	handle      CommandHandle
	initializer CommandInitializer
	interacter  CommandInteracter

	name        string
	description string
	aliases     []string
	help        string

	definition             *Input.InputDefinition
	applicationDefinition  *Input.InputDefinition
	isSingleCommand        bool
	hidden                 bool
	fullDefinition         *Input.InputDefinition
	ignoreValidationErrors bool
	synopsis               map[string]string
	usages                 []string
	input                  Input.InputInterface
	output                 Output.OutputInterface
}

func NewCommand(name string, handle CommandHandle) *Command {
	c := &Command{
		handle:      handle,
		initializer: nil,
		interacter:  nil,

		name:                   name,
		description:            "",
		aliases:                make([]string, 0),
		help:                   "",
		definition:             Input.NewInputDefinition(nil, nil),
		applicationDefinition:  nil,
		fullDefinition:         nil,
		hidden:                 false,
		isSingleCommand:        false,
		ignoreValidationErrors: false,
		synopsis:               make(map[string]string),
		usages:                 make([]string, 0),
		input:                  nil,
		output:                 nil,
	}

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

func (c *Command) SetInitializer(initializer CommandInitializer) {
	c.initializer = initializer
}

func (c *Command) SetInteracter(interactor CommandInteracter) {
	c.interacter = interactor
}

func (c *Command) IgnoreValidationErrors() {
	c.ignoreValidationErrors = true
}

func (c *Command) SetApplicationDefinition(definition *Input.InputDefinition) {
	c.applicationDefinition = definition
	c.fullDefinition = nil
}

func (c *Command) ApplicationDefinition() *Input.InputDefinition {
	return c.applicationDefinition
}

func (c *Command) MergeApplication(mergeArgs bool) {
	if c.applicationDefinition == nil {
		return
	}

	fullDefinition := Input.NewInputDefinition(nil, nil)
	fullDefinition.SetOptions(c.definition.GetOptionsArray())
	fullDefinition.AddOptions(c.applicationDefinition.GetOptionsArray())

	if mergeArgs {
		fullDefinition.SetArguments(c.applicationDefinition.GetArgumentsArray())
		fullDefinition.AddArguments(c.definition.GetArgumentsArray())
	} else {
		fullDefinition.SetArguments(c.definition.GetArgumentsArray())
	}

	c.fullDefinition = fullDefinition
}

func (c *Command) IsEnabled() bool {
	return true
}

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

	c.MergeApplication(true)

	input.Bind(c.GetDefinition())
	err := input.Parse()
	if err != nil && !c.ignoreValidationErrors {
		return 1, err
	}

	if c.initializer != nil {
		c.initializer(input, output)
	}

	if c.interacter != nil && input.IsInteractive() {
		c.interacter(input, output)
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

func (c *Command) AddArgument(name string, mode Input.InputArgumentMode, description string, defaultValue Input.InputType, validator Input.InputValidator) *Command {
	if c.definition != nil {
		c.definition.AddArgument(Input.NewInputArgument(name, mode, description, defaultValue, validator))
	}

	if c.fullDefinition != nil {
		c.definition.AddArgument(Input.NewInputArgument(name, mode, description, defaultValue, validator))
	}

	return c
}

func (c *Command) AddOption(name string, shortcut string, mode Input.InputOptionMode, description string, defaultValue Input.InputType, validator Input.InputValidator) *Command {
	if c.definition != nil {
		c.definition.AddOption(Input.NewInputOption(name, shortcut, mode, description, defaultValue, validator))
	}

	if c.fullDefinition != nil {
		c.fullDefinition.AddOption(Input.NewInputOption(name, shortcut, mode, description, defaultValue, validator))
	}

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

func (c *Command) ProcessedHelp() string {
	name := c.GetName()
	isSingleCommand := c.isSingleCommand

	placeholders := []string{`%command_name%`, `%command_full_name%`}

	var title string
	if !isSingleCommand {
		executable, err := os.Executable()
		if err != nil {
			title = executable + " " + name
		} else {
			title = os.Args[0]
		}
	}

	replacements := []string{name, title}

	help := c.GetHelp()
	if help == "" {
		help = c.GetDescription()
	}

	for i, placeholder := range placeholders {
		help = strings.Replace(help, placeholder, replacements[i], -1)
	}

	return help
}

func (c *Command) validateName(name string) {
	re := regexp.MustCompile("^[^:]+(:[^:]+)*")
	if !re.MatchString(name) {
		panic(fmt.Sprintf("Command name \"%s\" is invalid.", name))
	}
}

func (c *Command) Input() Input.InputInterface {
	if c.input == nil {
		panic("Command.Input() can only be called inside the scope of the command handle.")
	}
	return c.input
}

func (c *Command) Output() Output.OutputInterface {
	if c.output == nil {
		panic("Command.Output() can only be called inside the scope of the command handle.")
	}
	return c.output
}

func (c *Command) Describe(output Output.OutputInterface, options uint) {

}
