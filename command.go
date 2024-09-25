package cli

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/michielnijenhuis/cli/helper/array"
)

type CommandHandle func(io *IO)
type CommandHandleE func(io *IO) error

type Command struct {
	Name                   string
	Description            string
	Version                string
	LongVersion            string
	Aliases                []string
	Help                   string
	Commands               []*Command
	Flags                  []Flag
	Arguments              []Arg
	Run                    CommandHandle
	RunE                   CommandHandleE
	AutoExit               bool
	CatchErrors            bool
	Strict                 bool
	IgnoreValidationErrors bool
	Hidden                 bool
	PromptForInput         bool
	PrintHelpFunc          func(o *Output, command *Command)
	definition             *InputDefinition
	synopsis               map[string]string
	usages                 []string
	parent                 *Command
	commands               map[string]*Command
	runningCommand         *Command
	initialized            bool
	validated              bool
}

func (c *Command) Execute(args ...string) (err error) {
	width, height := TerminalSize()
	os.Setenv("LINES", fmt.Sprint(height))
	os.Setenv("COLUMNS", fmt.Sprint(width))

	i := NewInput(args...)
	i.Strict = c.Strict
	o := NewOutput(i)

	if c.CatchErrors {
		defer func() {
			if r := recover(); r != nil {
				err = r.(error)
			}
		}()
	}

	c.configureIO(i, o)

	if err = c.execute(i, o); err != nil {
		c.RenderError(o, err)
	}

	if c.AutoExit {
		if err != nil {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	return
}

func (c *Command) Subcommand(name string) *Command {
	c.init()
	return c.commands[name]
}

func (c *Command) HasSubcommands() bool {
	return len(c.Commands) > 0 || len(c.commands) > 0
}

func (c *Command) execute(i *Input, o *Output) error {
	c.validate()

	if i.HasParameterFlag("--version", true) || i.HasParameterFlag("-V", true) {
		o.Writeln(c.version(), 0)
		return nil
	}

	definition := c.Definition()
	command, args, err := c.findCommand(i, definition)

	if err != nil {
		notFound, ok := err.(*CommandNotFoundError)
		var alternatives []string
		if ok {
			alternatives = notFound.Alternatives()
		}
		interactive := i.IsInteractive()

		if ok && len(alternatives) == 1 && interactive {
			theme, _ := GetTheme("error")

			promptText := make([]string, 0, 3)
			if theme.Padding {
				promptText = append(promptText, "")
			}
			promptText = append(promptText, o.CreateBlock([]string{fmt.Sprintf("command \"%s\" is not defined", strings.Join(args, " "))}, "error", theme, true)...)
			promptText = append(promptText, Eol)

			alternative := alternatives[0]
			prompt := NewConfirmPrompt(i, o, fmt.Sprintf("Do you want to run \"%s\" instead?", alternative), false)
			prompt.Prefix = strings.Join(promptText, Eol)
			runAlternative, err := prompt.Render()
			if err != nil {
				return err
			}

			if !runAlternative {
				return nil
			}

			prompt.Clear()
			incorrectName := args[len(args)-1]
			args[len(args)-1] = alternative
			altenativeCmd := c
			for _, arg := range args {
				altenativeCmd = altenativeCmd.commands[arg]
			}

			command = altenativeCmd
			i.tokens = array.Remove(i.tokens, incorrectName)
		} else {
			if len(alternatives) > 1 {
				err = fmt.Errorf("command \"%s\" is ambiguous.\nDid you mean one of these?\n - %s", strings.Join(args, " "), strings.Join(alternatives, "\n - "))
			}

			return err
		}
	}

	wantsHelp := i.HasParameterFlag("--help", true) || i.HasParameterFlag("-h", true)
	if wantsHelp && command != nil {
		c.printHelp(o, command)
		return nil
	}

	err = i.Bind(command.Definition())
	if err != nil && !command.IgnoreValidationErrors {
		return err
	}

	err = i.Validate()
	if err != nil && !command.IgnoreValidationErrors {
		if command.PromptForInput {
			missingArgsErr, ok := err.(ErrMissingArguments)
			if !ok {
				return err
			}

			err = command.doPromptForInput(i, o, missingArgsErr.MissingArguments())
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	c.runningCommand = command
	io := &IO{
		Input:      i,
		Output:     o,
		definition: command.Definition(),
		Args:       i.Args,
	}

	if command.RunE != nil {
		err = command.RunE(io)
	} else if command.Run != nil {
		command.Run(io)
	} else if command == c || command.HasSubcommands() {
		c.printHelp(o, command)
	} else {
		panic("command must have a handle or subcommands")
	}

	c.runningCommand = nil

	return err
}

func (c *Command) GetHelp() string {
	version := c.version()

	if version != "" {
		return version
	}

	return "Console Tool"
}

func (c *Command) All() map[string]*Command {
	c.init()

	cmds := make(map[string]*Command)
	for _, cmd := range c.commands {
		cmds[cmd.Name] = cmd
	}

	return cmds
}

func (c *Command) version() string {
	c.validate()

	if c.LongVersion != "" {
		return c.LongVersion
	}

	var desc string
	if c.Description != "" {
		desc = c.Description
	} else {
		desc = c.Name
	}

	if c.Version != "" {
		return fmt.Sprintf("%s <accent>%s</accent>", desc, c.Version)
	}

	return desc
}

func (c *Command) AddCommand(command *Command) {
	if command == c {
		panic("cannot add a command to itself")
	}

	c.init()

	if command.Name == "" {
		panic("Commands must have a name.")
	}

	if _, exists := c.commands[command.Name]; exists {
		panic(fmt.Sprintf("command \"%s\" already exists", command.Name))
	}

	c.commands[command.Name] = command

	for _, alias := range command.Aliases {
		if _, exists := c.commands[alias]; exists {
			panic(fmt.Sprintf("command \"%s\" already exists", alias))
		}

		c.commands[alias] = command
	}

	command.SetParent(c)
	command.validate()
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
		c.synopsis[key] = strings.TrimSpace(fmt.Sprintf("%s %s", c.FullName(), c.Definition().Synopsis(short)))
	}

	return c.synopsis[key]
}

func (c *Command) AddUsage(usage string) {
	if !strings.HasPrefix(usage, c.Name) {
		usage = c.Name + " " + usage
	}

	if c.usages == nil {
		c.usages = make([]string, 0)
	}

	c.usages = append(c.usages, usage)
}

func (c *Command) Usages() []string {
	if c.usages == nil {
		c.usages = make([]string, 0)
	}

	return c.usages
}

func (c *Command) ProcessedHelp() string {
	name := c.Name
	isSingleCommand := c.parent == nil && len(c.Commands) == 0 && len(c.commands) == 0

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
	if help == "" && c.HasSubcommands() {
		help = fmt.Sprintf("Use \"%s [command] --help\" for more information about a command", c.FullName())
	}

	for i, placeholder := range placeholders {
		help = strings.ReplaceAll(help, placeholder, replacements[i])
	}

	return help
}

func (c *Command) validate() {
	if c.validated {
		return
	}
	c.validated = true

	re := regexp.MustCompile("^[^:]+(:[^:]+)*")
	if !re.MatchString(c.Name) {
		panic(fmt.Sprintf("command name \"%s\" is invalid", c.Name))
	}

	for _, alias := range c.Aliases {
		if !re.MatchString(alias) {
			panic(fmt.Sprintf("command alias \"%s\" is invalid", alias))
		}
	}
}

func (c *Command) RenderError(o *Output, err error) {
	o.Err(err)

	if c.runningCommand != nil {
		o.Writeln(
			fmt.Sprintf("<accent>%s %s</accent>", c.Name, c.runningCommand.Synopsis(false)),
			VerbosityQuiet,
		)
	}
}

func (c *Command) configureIO(i *Input, o *Output) {
	if i.HasParameterFlag("--ansi", true) {
		o.SetDecorated(true)
	} else if i.HasParameterFlag("--no-ansi", true) {
		o.SetDecorated(false)
	}

	if i.HasParameterFlag("--no-interaction", true) || i.HasParameterFlag("-n", true) {
		i.SetInteractive(false)
	}

	shellVerbosity, err := strconv.Atoi(os.Getenv("SHELL_VERBOSITY"))
	if err != nil {
		shellVerbosity = 0
	}

	switch shellVerbosity {
	case -1:
		o.SetVerbosity(VerbosityQuiet)
	case 1:
		o.SetVerbosity(VerbosityVerbose)
	case 2:
		o.SetVerbosity(VerbosityVeryVerbose)
	case 3:
		o.SetVerbosity(VerbosityDebug)
	default:
		shellVerbosity = 0
	}

	if i.HasParameterFlag("--quiet", true) || i.HasParameterFlag("-q", true) {
		o.SetVerbosity(VerbosityQuiet)
		shellVerbosity = -1
	} else {
		if i.HasParameterFlag("-vvv", true) ||
			i.HasParameterFlag("--verbose=3", true) ||
			i.ParameterFlag("--verbose", false, true) == "3" {
			o.SetVerbosity(VerbosityDebug)
			shellVerbosity = 3
		} else if i.HasParameterFlag("-vv", true) ||
			i.HasParameterFlag("--verbose=2", true) ||
			i.ParameterFlag("--verbose", false, true) == "2" {
			o.SetVerbosity(VerbosityVeryVerbose)
			shellVerbosity = 2
		} else if i.HasParameterFlag("-v", true) ||
			i.HasParameterFlag("--verbose=1", true) ||
			i.HasParameterFlag("--verbose", true) {
			o.SetVerbosity(VerbosityVerbose)
			shellVerbosity = 1
		}
	}

	if shellVerbosity == -1 {
		i.SetInteractive(false)
	}

	os.Setenv("SHELL_VERBOSITY", fmt.Sprint(shellVerbosity))
}

func (c *Command) SetParent(parent *Command) {
	c.parent = parent
}

func (c *Command) Parent() *Command {
	return c.parent
}

func (c *Command) findCommand(input *Input, definition *InputDefinition) (*Command, []string, error) {
	isOption := false
	argc := len(input.Args)
	arguments := make([]string, 0)

	for idx, token := range input.Args {
		// Is option
		if strings.HasPrefix(token, "-") {
			// If we have arguments, we can't have options anymore,
			// as the command arguments must be chained
			if len(arguments) > 0 {
				break
			}

			// Has value, or is last token
			if strings.Contains(token, "=") || idx+1 >= argc {
				continue
			}

			// If it's a long option, consider that everything after "--" is the option name.
			// Otherwise, use the last char (if it's a short option set, only the last one can take a value with space separator)
			var name string
			if strings.HasPrefix(token, "--") {
				name = token[2:]
			} else {
				name = token[len(token)-1:]
			}

			flag, _ := definition.Flag(name)
			if flag == nil {
				// Try again with the shortcut
				flag, _ = definition.FlagForShortcut(name)

				if flag == nil {
					continue
				}
			}

			// If flag accepts a value, check if the next token is not an option value
			if FlagAcceptsValue(flag) && !strings.HasPrefix(input.Args[idx+1], "-") {
				isOption = true
			}

			continue
		}

		// Is value for option
		if isOption {
			isOption = false
			continue
		}

		arguments = append(arguments, token)
	}

	var command *Command = c
	for i, arg := range arguments {
		cmd := command.commands[arg]
		if cmd != nil {
			input.tokens = array.Remove(input.tokens, arg)
			command = cmd
		} else {
			if len(command.Definition().arguments) == 0 && command.HasSubcommands() {
				alternatives := c.findAlternatives(arg, array.SortedKeys(command.commands))
				return nil, arguments[:i+1], CommandNotFound(fmt.Sprintf("command \"%s\" does not exist", arg), alternatives)
			}

			break
		}
	}

	return command, nil, nil
}

func (c *Command) defaultInputDefinition() *InputDefinition {
	helpFlag := &BoolFlag{
		Name:        "help",
		Shortcuts:   []string{"h"},
		Description: "Display help for the program or a given command",
	}
	versionFlag := &BoolFlag{
		Name:        "version",
		Shortcuts:   []string{"V"},
		Description: "Display the program version",
	}
	quietFlag := &BoolFlag{
		Name:        "quiet",
		Shortcuts:   []string{"q"},
		Description: "Do not output any message",
	}
	verboseflag := &BoolFlag{
		Name:        "verbose",
		Shortcuts:   []string{"v", "vv", "vvv"},
		Description: "Increase the verbosity of messages: normal (1), verbose (2) or debug (3)",
	}
	ansiFlag := &BoolFlag{
		Name:        "ansi",
		Negatable:   true,
		Description: "Force (or disable --no-ansi) ANSI output",
	}
	noInteractionFlag := &BoolFlag{
		Name:        "no-interaction",
		Shortcuts:   []string{"n"},
		Description: "Do not ask any interactive question",
	}

	flags := []Flag{
		helpFlag,
		quietFlag,
		verboseflag,
		versionFlag,
		ansiFlag,
		noInteractionFlag,
	}

	definition := &InputDefinition{}
	definition.SetFlags(flags)

	return definition
}

func (c *Command) init() {
	if c.initialized {
		return
	}

	c.initialized = true

	if c.commands == nil {
		c.commands = make(map[string]*Command)
	}

	for _, command := range c.Commands {
		command.SetParent(c)
		c.AddCommand(command)
	}

	c.Commands = nil
}

func (c *Command) Definition() *InputDefinition {
	if c.definition == nil {
		nativeDefinition := c.defaultInputDefinition()

		c.definition = &InputDefinition{}
		c.definition.SetArguments(c.Arguments)
		c.definition.SetFlags(c.Flags)
		c.definition.AddArguments(nativeDefinition.GetArguments())
		c.definition.AddFlags(nativeDefinition.GetFlags())
	}

	return c.definition
}

func (c *Command) printHelp(output *Output, command *Command) {
	if c.PrintHelpFunc != nil {
		c.PrintHelpFunc(output, command)
		return
	}

	d := TextDescriptor{output}
	d.DescribeCommand(command, &DescriptorOptions{})
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

		i.Args = append(i.Args, answer)
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

		i.Args = append(i.Args, answers...)

		return nil
	default:
		panic("unsupported argument type")
	}
}

func (c *Command) Namespace() string {
	names := make([]string, 0)

	p := c.parent
	for p != nil {
		names = append(names, p.Name)
		p = p.parent
	}

	slices.Reverse(names)

	return strings.Join(names, " ")
}

func (c *Command) FullName() string {
	ns := c.Namespace()
	if ns == "" {
		return c.Name
	}

	return ns + " " + c.Name
}

func (c *Command) findAlternatives(name string, collection []string) []string {
	treshold := int(1e3)
	alternatives := make(map[string]int)

	for _, item := range collection {
		_, exists := alternatives[item]
		lev := levenshtein(name, item)
		if lev <= len(name)/3 || (name != "" && strings.Contains(item, name)) {
			if exists {
				alternatives[item] += lev
			} else {
				alternatives[item] += treshold
			}
		} else if exists {
			alternatives[item] += treshold
		}
	}

	for _, item := range collection {
		lev := levenshtein(name, item)
		if lev <= len(name)/3 || strings.Contains(item, name) {
			_, ok := alternatives[item]
			if ok {
				alternatives[item] -= lev
			} else {
				alternatives[item] = lev
			}
		}
	}

	filteredAlternatives := make([]string, 0, len(alternatives))
	for k, v := range alternatives {
		if v < 2*treshold {
			filteredAlternatives = append(filteredAlternatives, k)
		}
	}

	sort.Strings(filteredAlternatives)
	return filteredAlternatives
}

func levenshtein(a string, b string) int {
	aLen := len(a)
	bLen := len(b)

	if aLen == 0 {
		return bLen
	}

	if bLen == 0 {
		return aLen
	}

	matrix := make([][]int, 0, max(aLen, bLen))

	for i := 0; i <= aLen; i++ {
		s := make([]int, 0, bLen)
		s = append(s, i)
		matrix = append(matrix, s)
	}

	for j := 0; j <= bLen; j++ {
		if j == 0 {
			matrix[0][j] = j
		} else {
			matrix[0] = append(matrix[0], j)
		}
	}

	for i := 1; i <= aLen; i++ {
		for j := 1; j <= bLen; j++ {
			var cost int
			if a[i-1] != b[j-1] {
				cost = 1
			}

			mLen := len(matrix)
			if i >= mLen {
				matrix = append(matrix, make([]int, 0))
			}

			v := min(matrix[i-1][j]+1, matrix[i][j-1]+1, matrix[i-1][j-1]+cost)

			iLen := len(matrix[i])
			if j >= iLen {
				matrix[i] = append(matrix[i], v)
			} else {
				matrix[i][j] = v
			}
		}
	}

	return matrix[aLen][bLen]
}
