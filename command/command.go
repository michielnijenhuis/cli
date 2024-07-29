package command

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/michielnijenhuis/cli/input"
	"github.com/michielnijenhuis/cli/output"
	"github.com/michielnijenhuis/cli/style"
)

type CommandHandle func(self *Command) (int, error)
type CommandInitializer func(input input.InputInterface, output output.OutputInterface)
type CommandInteracter func(input input.InputInterface, output output.OutputInterface)

type Command struct {
	Handle      CommandHandle
	initializer CommandInitializer
	interacter  CommandInteracter

	Name        string
	Description string
	Aliases     []string
	Help        string

	definition             *input.InputDefinition
	applicationDefinition  *input.InputDefinition
	isSingleCommand        bool
	hidden                 bool
	fullDefinition         *input.InputDefinition
	ignoreValidationErrors bool
	synopsis               map[string]string
	usages                 []string
	input                  input.InputInterface
	output                 output.OutputInterface
	meta                   any
	validatedName          bool
}

func NewCommand(name string, handle CommandHandle) *Command {
	c := &Command{
		Handle:                 handle,
		initializer:            nil,
		interacter:             nil,
		Name:                   name,
		Description:            "",
		Aliases:                make([]string, 0),
		Help:                   "",
		definition:             input.NewInputDefinition(nil, nil),
		applicationDefinition:  nil,
		fullDefinition:         nil,
		hidden:                 false,
		isSingleCommand:        false,
		ignoreValidationErrors: false,
		synopsis:               make(map[string]string),
		usages:                 make([]string, 0),
		input:                  nil,
		output:                 nil,
		meta:                   nil,
		validatedName:          false,
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

func (c *Command) SetApplicationDefinition(definition *input.InputDefinition) {
	c.applicationDefinition = definition
	c.fullDefinition = nil
}

func (c *Command) ApplicationDefinition() *input.InputDefinition {
	return c.applicationDefinition
}

func (c *Command) MergeApplication(mergeArgs bool) {
	if c.applicationDefinition == nil {
		return
	}

	fullDefinition := input.NewInputDefinition(nil, nil)
	fullDefinition.SetOptions(c.GetNativeDefinition().GetOptionsArray())
	fullDefinition.AddOptions(c.applicationDefinition.GetOptionsArray())

	if mergeArgs {
		fullDefinition.SetArguments(c.applicationDefinition.GetArgumentsArray())
		fullDefinition.AddArguments(c.GetNativeDefinition().GetArgumentsArray())
	} else {
		fullDefinition.SetArguments(c.GetNativeDefinition().GetArgumentsArray())
	}

	c.fullDefinition = fullDefinition
}

func (c *Command) IsEnabled() bool {
	return true
}

func (c *Command) execute(input input.InputInterface, output output.OutputInterface) (int, error) {
	c.input = input
	c.output = style.NewStyle(input, output)

	return c.Handle(c)
}

func (c *Command) Fail(e string) (int, error) {
	return 1, errors.New(e)
}

func (c *Command) StringArgument(name string) (string, error) {
	return c.input.GetStringArgument(name)
}

func (c *Command) ArrayArgument(name string) ([]string, error) {
	return c.input.GetArrayArgument(name)
}

func (c *Command) Arguments() map[string]input.InputType {
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

func (c *Command) Options() map[string]input.InputType {
	return c.input.GetOptions()
}

func (c *Command) Run(i input.InputInterface, o output.OutputInterface) (int, error) {
	if i == nil {
		argvInput, err := input.NewArgvInput(nil, nil)
		if err != nil {
			return 1, err
		}
		i = argvInput
	}

	if o == nil {
		o = output.NewConsoleOutput(0, true, nil)
	}

	c.MergeApplication(true)

	i.Bind(c.GetDefinition())
	err := i.Parse()
	if err != nil && !c.ignoreValidationErrors {
		return 1, err
	}

	if c.initializer != nil {
		c.initializer(i, o)
	}

	if c.interacter != nil && i.IsInteractive() {
		c.interacter(i, o)
	}

	if i.HasArgument("command") {
		command, _ := i.GetStringArgument("command")
		if command == "" {
			i.SetArgument("command", c.GetName())
		}
	}

	validationErr := i.Validate()

	if validationErr != nil {
		return 1, validationErr
	}

	return c.execute(i, o)
}

func (c *Command) SetDefinition(definition *input.InputDefinition) {
	if definition != nil {
		c.definition = definition
	} else {
		c.definition = input.NewInputDefinition(nil, nil)
	}

	c.fullDefinition = nil
}

func (c *Command) GetDefinition() *input.InputDefinition {
	if c.fullDefinition == nil {
		return c.GetNativeDefinition()
	}

	return c.fullDefinition
}

func (c *Command) GetNativeDefinition() *input.InputDefinition {
	if c.definition == nil {
		c.definition = input.NewInputDefinition(nil, nil)
	}

	return c.definition
}

func (c *Command) AddArgument(name string, mode input.InputArgumentMode, description string, defaultValue input.InputType, validator input.InputValidator) *Command {
	if c.definition != nil {
		c.GetNativeDefinition().AddArgument(input.NewInputArgument(name, mode, description, defaultValue, validator))
	}

	if c.fullDefinition != nil {
		c.GetNativeDefinition().AddArgument(input.NewInputArgument(name, mode, description, defaultValue, validator))
	}

	return c
}

func (c *Command) AddOption(name string, shortcut string, mode input.InputOptionMode, description string, defaultValue input.InputType, validator input.InputValidator) *Command {
	if c.definition != nil {
		c.GetNativeDefinition().AddOption(input.NewInputOption(name, shortcut, mode, description, defaultValue, validator))
	}

	if c.fullDefinition != nil {
		c.fullDefinition.AddOption(input.NewInputOption(name, shortcut, mode, description, defaultValue, validator))
	}

	return c
}

func (c *Command) SetName(name string) *Command {
	c.validateName(name)
	c.Name = name
	return c
}

func (c *Command) GetName() string {
	if c.Name == "" {
		return "CLI Tool"
	}

	return c.Name
}

func (c *Command) SetHidden(hidden bool) *Command {
	c.hidden = hidden
	return c
}

func (c *Command) IsHidden() bool {
	return c.hidden
}

func (c *Command) SetDescription(description string) *Command {
	c.Description = description
	return c
}

func (c *Command) GetDescription() string {
	return c.Description
}

func (c *Command) SetHelp(help string) *Command {
	c.Help = help
	return c
}

func (c *Command) GetHelp() string {
	return c.Help
}

func (c *Command) SetAliases(aliases []string) *Command {
	if aliases != nil {
		for _, alias := range aliases {
			c.validateName(alias)
		}

		c.Aliases = aliases
	} else {
		c.Aliases = []string{}
	}

	return c
}

func (c *Command) GetAliases() []string {
	if c.Aliases == nil {
		c.Aliases = []string{}
	}

	return c.Aliases
}

func (c *Command) GetSynopsis(short bool) string {
	key := "long"
	if short {
		key = "short"
	}

	if c.synopsis == nil {
		c.synopsis = make(map[string]string)
	}

	if c.synopsis[key] == "" {
		c.synopsis[key] = strings.TrimSpace(fmt.Sprintf("%s %s", c.Name, c.GetNativeDefinition().GetSynopsis(short)))
	}

	return c.synopsis[key]
}

func (c *Command) AddUsage(usage string) *Command {
	if !strings.HasPrefix(usage, c.Name) {
		usage = c.Name + " " + usage
	}

	if c.usages == nil {
		c.usages = make([]string, 0)
	}

	c.usages = append(c.usages, usage)

	return c
}

func (c *Command) GetUsages() []string {
	if c.usages == nil {
		c.usages = make([]string, 0)
	}

	return c.usages
}

func (c *Command) ProcessedHelp() string {
	name := c.GetName()
	isSingleCommand := c.isSingleCommand

	placeholders := []string{`%command.name%`, `%command.full_name%`}

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

func (c *Command) Input() input.InputInterface {
	if c.input == nil {
		panic("Command.Input() can only be called inside the scope of the command handle.")
	}
	return c.input
}

func (c *Command) Output() output.OutputInterface {
	if c.output == nil {
		panic("Command.Output() can only be called inside the scope of the command handle.")
	}
	return c.output
}

func (c *Command) Describe(output output.OutputInterface, options uint) {

}

func (c *Command) Meta() any {
	return c.meta
}

func (c *Command) SetMeta(meta any) {
	c.meta = meta
}
