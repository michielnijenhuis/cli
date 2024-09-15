package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type CommandHandle func(ctx *Ctx)
type CommandHandleE func(ctx *Ctx) error
type CommandInitializer func(i *Input, o *Output)
type CommandInteracter func(i *Input, o *Output)
type PromptFunc func(i *Input, o *Output, arg Arg) error

type Command struct {
	Run         CommandHandle
	RunE        CommandHandleE
	Initializer CommandInitializer
	Interacter  CommandInteracter

	Name           string
	Description    string
	Aliases        []string
	Help           string
	Signature      string
	Flags          []Flag
	Arguments      []Arg
	Debug          bool
	PromptForInput bool

	definition             *InputDefinition
	applicationDefinition  *InputDefinition
	isSingleCommand        bool
	Hidden                 bool
	fullDefinition         *InputDefinition
	IgnoreValidationErrors bool
	synopsis               map[string]string
	usages                 []string
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

	if mergeArgs {
		fullDefinition.SetArguments(c.applicationDefinition.GetArguments())
		fullDefinition.AddArguments(c.NativeDefinition().GetArguments())
	} else {
		fullDefinition.SetArguments(c.NativeDefinition().GetArguments())
	}

	fullDefinition.SetFlags(c.NativeDefinition().GetFlags())
	fullDefinition.AddFlags(c.applicationDefinition.GetFlags())

	c.fullDefinition = fullDefinition
}

func (c *Command) IsEnabled() bool {
	return true
}

func (c *Command) execute(input *Input, output *Output) (int, error) {
	ctx := &Ctx{
		Input:      input,
		Output:     output,
		definition: c.fullDefinition,
		Args:       input.Args,
		Debug:      c.Debug,
		Logger:     output.Logger,
	}

	if c.RunE != nil {
		err := c.RunE(ctx)
		if err != nil && ctx.Code == 0 {
			ctx.Code = 1
		}

		return ctx.Code, err
	}

	if c.Run == nil {
		panic("command must have a handle")
	}

	c.Run(ctx)
	return ctx.Code, nil
}

func (c *Command) Execute(args ...string) (int, error) {
	return c.ExecuteWith(NewInput(args...), nil)
}

func (c *Command) ExecuteWith(i *Input, o *Output) (int, error) {
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

	validationErr := i.Validate()

	if validationErr != nil {
		if c.PromptForInput {
			missingArgsErr, ok := validationErr.(ErrMissingArguments)
			if !ok {
				return 1, validationErr
			}

			err = c.doPromptForInput(i, o, missingArgsErr.MissingArguments())
			if err != nil {
				return 1, validationErr
			}
		} else {
			return 1, validationErr
		}
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
		c.definition.SetArguments(c.Arguments)
		c.definition.SetFlags(c.Flags)
	}

	return c.definition
}

func (c *Command) AddArgument(arg Arg) *Command {
	c.NativeDefinition().AddArgument(arg)

	if c.fullDefinition != nil {
		c.fullDefinition.AddArgument(arg)
	}

	return c
}

func (c *Command) AddFlag(flag Flag) *Command {
	c.NativeDefinition().AddFlag(flag)

	if c.fullDefinition != nil {
		c.fullDefinition.AddFlag(flag)
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
		c.synopsis[key] = strings.TrimSpace(fmt.Sprintf("%s %s", c.Name, c.Definition().Synopsis(short)))
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

func (c *Command) doPromptForInput(i *Input, o *Output, missingArgs []string) error {
	for _, arg := range missingArgs {
		a, err := i.definition.Argument(arg)
		if err != nil {
			return err
		}

		err = c.promptArgument(i, o, a)
		if err != nil {
			return err
		}
	}

	o.NewLine(1)

	return i.Validate()
}

func (c *Command) promptArgument(i *Input, o *Output, arg Arg) error {
	name := arg.GetName()

	desc := arg.GetDescription()
	if desc == "" {
		panic(fmt.Sprintf("argument \"%s\" is missing a description", name))
	}

	q := strings.ToLower(string(desc[0])) + desc[1:]

	switch a := arg.(type) {
	case *StringArg:
		var answer string
		var err error
		if a.Options != nil {
			prompt := NewSelectPrompt(i, o, fmt.Sprintf("What is %s?", q), a.Options, nil, "")
			prompt.Required = true
			answer, err = prompt.Render()
		} else {
			prompt := NewTextPrompt(i, o, fmt.Sprintf("What is %s?", q), "")
			prompt.Required = true
			answer, err = prompt.Render()
		}

		if err != nil {
			return err
		}

		i.SetArgument(name, answer)

		return nil
	case *ArrayArg:
		// TODO: implement options (requires multiselect prompt)
		prompt := NewArrayPrompt(i, o, fmt.Sprintf("What is %s?", q), nil)
		prompt.Required = true
		answers, err := prompt.Render()
		if err != nil {
			return err
		}

		for _, answer := range answers {
			i.SetArgument(name, answer)
		}

		return nil
	default:
		panic("unsupported argument type")
	}
}
