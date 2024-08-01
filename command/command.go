package command

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/michielnijenhuis/cli/exec"
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
	Hidden                 bool
	fullDefinition         *input.InputDefinition
	IgnoreValidationErrors bool
	synopsis               map[string]string
	usages                 []string
	input                  input.InputInterface
	output                 output.OutputInterface
	meta                   any
}

func (c *Command) SetInitializer(initializer CommandInitializer) {
	c.initializer = initializer
}

func (c *Command) SetInteracter(interactor CommandInteracter) {
	c.interacter = interactor
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
	fullDefinition.SetOptions(c.NativeDefinition().OptionsArray())
	fullDefinition.AddOptions(c.applicationDefinition.OptionsArray())

	if mergeArgs {
		fullDefinition.SetArguments(c.applicationDefinition.ArgumentsArray())
		fullDefinition.AddArguments(c.NativeDefinition().ArgumentsArray())
	} else {
		fullDefinition.SetArguments(c.NativeDefinition().ArgumentsArray())
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
	return c.input.StringArgument(name)
}

func (c *Command) ArrayArgument(name string) ([]string, error) {
	return c.input.ArrayArgument(name)
}

func (c *Command) Arguments() map[string]input.InputType {
	return c.input.Arguments()
}

func (c *Command) BoolOption(name string) (bool, error) {
	return c.input.BoolOption(name)
}

func (c *Command) StringOption(name string) (string, error) {
	return c.input.StringOption(name)
}

func (c *Command) ArrayOption(name string) ([]string, error) {
	return c.input.ArrayOption(name)
}

func (c *Command) Options() map[string]input.InputType {
	return c.input.Options()
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

	i.Bind(c.Definition())
	err := i.Parse()
	if err != nil && !c.IgnoreValidationErrors {
		return 1, err
	}

	if c.initializer != nil {
		c.initializer(i, o)
	}

	if c.interacter != nil && i.IsInteractive() {
		c.interacter(i, o)
	}

	if i.HasArgument("command") {
		command, _ := i.StringArgument("command")
		if command == "" {
			err := i.SetArgument("command", c.Name)
			if err != nil {
				return 1, err
			}
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

func (c *Command) Definition() *input.InputDefinition {
	if c.fullDefinition == nil {
		return c.NativeDefinition()
	}

	return c.fullDefinition
}

func (c *Command) NativeDefinition() *input.InputDefinition {
	if c.definition == nil {
		c.definition = input.NewInputDefinition(nil, nil)
	}

	return c.definition
}

func (c *Command) AddArgument(arg *input.InputArgument) *Command {
	c.NativeDefinition().AddArgument(arg)

	if c.fullDefinition != nil {
		c.fullDefinition.AddArgument(arg.Clone())
	}

	return c
}

func (c *Command) DefineArgument(name string, mode input.InputArgumentMode, description string, defaultValue input.InputType, validator input.InputValidator) *Command {
	a := input.NewInputArgument(name, mode, description)
	if defaultValue != nil {
		a.SetDefaultValue(defaultValue)
	}
	if validator != nil {
		a.SetValidator(validator)
	}
	c.NativeDefinition().AddArgument(a)

	if c.fullDefinition != nil {
		a := input.NewInputArgument(name, mode, description)
		if defaultValue != nil {
			a.SetDefaultValue(defaultValue)
		}
		if validator != nil {
			a.SetValidator(validator)
		}
		c.fullDefinition.AddArgument(a)
	}

	return c
}

func (c *Command) AddOption(option *input.InputOption) *Command {
	c.NativeDefinition().AddOption(option)

	if c.fullDefinition != nil {
		c.fullDefinition.AddOption(option.Clone())
	}

	return c
}

func (c *Command) DefineOption(name string, shortcut string, mode input.InputOptionMode, description string, defaultValue input.InputType, validator input.InputValidator) *Command {
	o := input.NewInputOption(name, shortcut, mode, description)
	if defaultValue != nil {
		o.SetDefaultValue(defaultValue)
	}
	if validator != nil {
		o.SetValidator(validator)
	}
	c.NativeDefinition().AddOption(o)

	if c.fullDefinition != nil {
		o := input.NewInputOption(name, shortcut, mode, description)
		if defaultValue != nil {
			o.SetDefaultValue(defaultValue)
		}
		if validator != nil {
			o.SetValidator(validator)
		}
		c.fullDefinition.AddOption(o)
	}

	return c
}

func (c *Command) SetName(name string) *Command {
	c.validateName(name)
	c.Name = name
	return c
}

func (c *Command) SetDescription(description string) *Command {
	c.Description = description
	return c
}

func (c *Command) SetHelp(help string) *Command {
	c.Help = help
	return c
}

func (c *Command) SetHidden(hidden bool) *Command {
	c.Hidden = hidden
	return c
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

func (c *Command) Synopsis(short bool) string {
	key := "long"
	if short {
		key = "short"
	}

	if c.synopsis == nil {
		c.synopsis = make(map[string]string)
	}

	if c.synopsis[key] == "" {
		c.synopsis[key] = strings.TrimSpace(fmt.Sprintf("%s %s", c.Name, c.NativeDefinition().Synopsis(short)))
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

func (c *Command) Usages() []string {
	if c.usages == nil {
		c.usages = make([]string, 0)
	}

	return c.usages
}

func (c *Command) ProcessedHelp() string {
	name := c.Name
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

	help := c.Help
	if help == "" {
		help = c.Description
	}

	for i, placeholder := range placeholders {
		help = strings.ReplaceAll(help, placeholder, replacements[i])
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

func (c *Command) Exec(cmd string, shell string, inherit bool) (string, error) {
	options := &exec.ChildProcessOptions{
		Shell:   shell,
		Inherit: inherit,
		Pipe:    !inherit,
	}

	if inherit {
		i := c.input
		o := c.output

		if stream, ok := i.(input.StreamableInputInterface); ok {
			options.Stdin = stream.Stream()
		}

		if stream, ok := o.(output.StreamOutputInterface); ok {
			options.Stdout = stream.Stream()
		}

		if console, ok := o.(output.ConsoleOutputInterface); ok {
			errorOutput := console.ErrorOutput()
			if stream, ok := errorOutput.(output.StreamOutputInterface); ok {
				options.Stderr = stream.Stream()
			}
		}
	}

	cp := &exec.ChildProcess{
		Cmd:                 cmd,
		ChildProcessOptions: options,
	}

	return cp.Run()
}
