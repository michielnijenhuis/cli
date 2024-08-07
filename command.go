package cli

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type CommandHandle func(self *Command) (int, error)
type CommandInitializer func(input *Input, output *Output)
type CommandInteracter func(input *Input, output *Output)

type Command struct {
	Handle      CommandHandle
	Initializer CommandInitializer
	Interacter  CommandInteracter

	Name        string
	Description string
	Aliases     []string
	Help        string

	definition             *InputDefinition
	applicationDefinition  *InputDefinition
	isSingleCommand        bool
	Hidden                 bool
	fullDefinition         *InputDefinition
	IgnoreValidationErrors bool
	synopsis               map[string]string
	usages                 []string
	input                  *Input
	output                 *Output
	meta                   any
}

func (c *Command) SetApplicationDefinition(definition *InputDefinition) {
	c.applicationDefinition = definition
	c.fullDefinition = nil
}

func (c *Command) ApplicationDefinition() *InputDefinition {
	return c.applicationDefinition
}

func (c *Command) MergeApplication(mergeArgs bool) {
	if c.applicationDefinition == nil {
		return
	}

	fullDefinition := &InputDefinition{}
	fullDefinition.SetOptions(c.NativeDefinition().GetOptions())
	fullDefinition.AddOptions(c.applicationDefinition.GetOptions())

	if mergeArgs {
		fullDefinition.SetArguments(c.applicationDefinition.GetArguments())
		fullDefinition.AddArguments(c.NativeDefinition().GetArguments())
	} else {
		fullDefinition.SetArguments(c.NativeDefinition().GetArguments())
	}

	c.fullDefinition = fullDefinition
}

func (c *Command) IsEnabled() bool {
	return true
}

func (c *Command) execute(input *Input, output *Output) (int, error) {
	checkPtr(input, "input")
	checkPtr(output, "output")

	c.input = input
	c.output = output

	return c.Handle(c)
}

func (c *Command) Fail(e string) (int, error) {
	return 1, errors.New(e)
}

func (c *Command) StringArgument(name string) (string, error) {
	return c.Input().StringArgument(name)
}

func (c *Command) ArrayArgument(name string) ([]string, error) {
	return c.Input().ArrayArgument(name)
}

func (c *Command) Arguments() map[string]InputType {
	return c.Input().Arguments()
}

func (c *Command) BoolOption(name string) (bool, error) {
	return c.Input().BoolOption(name)
}

func (c *Command) StringOption(name string) (string, error) {
	return c.Input().StringOption(name)
}

func (c *Command) ArrayOption(name string) ([]string, error) {
	return c.Input().ArrayOption(name)
}

func (c *Command) Options() map[string]InputType {
	return c.Input().Options()
}

func (c *Command) Run(i *Input, o *Output) (int, error) {
	if i == nil {
		i = NewInput()
	}

	if o == nil {
		o = NewOutput(i)
	}

	c.MergeApplication(true)

	err := i.Bind(c.Definition())
	if err != nil && !c.IgnoreValidationErrors {
		return 1, err
	}

	if c.Initializer != nil {
		c.Initializer(i, o)
	}

	if c.Interacter != nil && i.IsInteractive() {
		c.Interacter(i, o)
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

func (c *Command) SetDefinition(definition *InputDefinition) {
	if definition != nil {
		c.definition = definition
	} else {
		c.definition = &InputDefinition{}
	}

	c.fullDefinition = nil
}

func (c *Command) Definition() *InputDefinition {
	if c.fullDefinition == nil {
		return c.NativeDefinition()
	}

	return c.fullDefinition
}

func (c *Command) NativeDefinition() *InputDefinition {
	if c.definition == nil {
		c.definition = &InputDefinition{}
	}

	return c.definition
}

func (c *Command) AddArgument(arg *InputArgument) *Command {
	c.NativeDefinition().AddArgument(arg)

	if c.fullDefinition != nil {
		c.fullDefinition.AddArgument(arg.Clone())
	}

	return c
}

func (c *Command) DefineArgument(name string, mode InputArgumentMode, description string, defaultValue InputType, validator InputValidator) *Command {
	a := &InputArgument{
		Name:        name,
		Mode:        mode,
		Description: description,
	}
	if defaultValue != nil {
		a.SetDefaultValue(defaultValue)
	}
	if validator != nil {
		a.SetValidator(validator)
	}
	c.NativeDefinition().AddArgument(a)

	if c.fullDefinition != nil {
		a = a.Clone()
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

func (c *Command) AddOption(option *InputOption) *Command {
	c.NativeDefinition().AddOption(option)

	if c.fullDefinition != nil {
		c.fullDefinition.AddOption(option.Clone())
	}

	return c
}

func (c *Command) DefineOption(name string, shortcut string, mode InputOptionMode, description string, defaultValue InputType, validator InputValidator) *Command {
	o := &InputOption{
		Name:        name,
		Shortcut:    shortcut,
		Mode:        mode,
		Description: description,
	}
	if defaultValue != nil {
		o.SetDefaultValue(defaultValue)
	}
	if validator != nil {
		o.SetValidator(validator)
	}
	c.NativeDefinition().AddOption(o)

	if c.fullDefinition != nil {
		o := &InputOption{
			Name:        name,
			Shortcut:    shortcut,
			Mode:        mode,
			Description: description,
		}
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

func (c *Command) Input() *Input {
	if c.input == nil {
		panic("Command.Input() can only be called inside the scope of the command handle.")
	}
	return c.input
}

func (c *Command) Output() *Output {
	if c.output == nil {
		panic("Command.Output() can only be called inside the scope of the command handle.")
	}
	return c.output
}

func (c *Command) Describe(output *Output, options uint) {

}

func (c *Command) Meta() any {
	return c.meta
}

func (c *Command) SetMeta(meta any) {
	c.meta = meta
}

func (c *Command) Spawn(cmd string, shell string, inherit bool) *ChildProcess {
	cp := &ChildProcess{
		Cmd:     cmd,
		Shell:   shell,
		Inherit: inherit,
		Pipe:    !inherit,
	}

	if inherit {
		i := c.input
		o := c.output

		cp.Stdin = i.Stream
		cp.Stdout = o.Stream
		cp.Stderr = o.Stderr.Stream
	}

	return cp
}

func (c *Command) Exec(cmd string, shell string, inherit bool) (string, error) {
	return c.Spawn(cmd, shell, inherit).Run()
}

func (c *Command) NewLine(count uint) {
	for count > 0 {
		c.output.Writeln("", 0)
		count--
	}
}

func (c *Command) Err(messages ...string) {
	c.writeLine(messages, "error")
}

func (c *Command) Info(messages ...string) {
	c.writeLine(messages, "info")
}

func (c *Command) Warn(messages ...string) {
	c.writeLine(messages, "warning")
}

func (c *Command) Ok(messages ...string) {
	c.writeLine(messages, "ok")
}

func (c *Command) Comment(messages ...string) {
	c.output.Comment(messages...)
}

func (c *Command) Alert(messages ...string) {
	length := 0
	for _, message := range messages {
		length = max(length, len(message))
	}
	length += 12

	c.writeLine([]string{fmt.Sprintf("<fg=yellow>%s </>", strings.Repeat("*", length))}, "alert")
	for i := range messages {
		messages[i] = fmt.Sprintf("%s<fg=yellow>*</>     %s     <fg=yellow>*</>", strings.Repeat(" ", 8), messages[i])
	}
	c.Writelns(messages)
	c.Writeln(fmt.Sprintf("<fg=yellow>%s%s</>", strings.Repeat(" ", 8), strings.Repeat("*", length)))
	c.NewLine(1)
}

func (c *Command) Write(message string) {
	c.output.Write(message, false, 0)
}

func (c *Command) Writeln(message string) {
	c.output.Writeln(message, 0)
}

func (c *Command) Writelns(messages []string) {
	c.output.Writelns(messages, 0)
}

func (c *Command) writeLine(messages []string, tag string) {
	if len(messages) == 0 {
		return
	}

	if tag != "" {
		messages[0] = fmt.Sprintf("<%s> %s </%s> %s", tag, strings.ToUpper(tag), tag, messages[0])

		for i := 1; i < len(messages); i++ {
			messages[i] = strings.Repeat(" ", len(tag)+3) + messages[i]
		}
	}

	c.output.Writelns(messages, 0)
}

func (c *Command) Spinner(fn func(), message string) {
	style, _ := c.output.Formatter().Style("prompt")
	s := NewSpinner(c.input, c.output, message, nil, style.foreground)
	s.Spin(fn)
}

func (c *Command) IsQuiet() bool {
	return c.output.IsQuiet()
}

func (c *Command) IsVerbose() bool {
	return c.output.IsVerbose()
}

func (c *Command) IsVeryVerbose() bool {
	return c.output.IsVeryVerbose()
}

func (c *Command) IsDebug() bool {
	return c.output.IsDebug()
}

func (c *Command) IsDecorated() bool {
	return c.output.IsDecorated()
}
