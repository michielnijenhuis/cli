package cli

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/michielnijenhuis/cli/helper/array"
	"github.com/michielnijenhuis/cli/terminal"
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
	PromptForCommand       bool
	PrintHelpFunc          func(o *Output, command *Command)
	NativeFlags            []string
	CascadeNativeFlags     bool
	definition             *InputDefinition
	synopsis               map[string]string
	usages                 []string
	parent                 *Command
	commands               map[string]*Command
	runningCommand         *Command
	initialized            bool
	validated              bool
	configuredIO           bool
	input                  *Input
	output                 *Output
}

func (c *Command) Execute(args ...string) (err error) {
	width, height := terminal.Size()
	os.Setenv("LINES", fmt.Sprint(height))
	os.Setenv("COLUMNS", fmt.Sprint(width))

	i := NewInput(args...)
	i.Strict = c.Strict
	o := NewOutput(i)

	c.input = i
	c.output = o

	var caughtError = false

	if c.CatchErrors {
		defer func() {
			if r := recover(); r != nil {
				recovered, ok := r.(error)
				if ok {
					err = recovered
				} else {
					str, ok := r.(string)
					if ok {
						err = errors.New(str)
					} else {
						err = fmt.Errorf("%v", r)
					}
				}

				caughtError = true
				c.handleError(o, err)
			}
		}()
	}

	c.configureIO(i, o)
	c.InitDefaultCompletionCmd(o.Stream)
	c.initCompleteCmd(i.Args)
	err = c.execute(i, o)

	if !caughtError {
		c.handleError(o, err)
	}

	return
}

func (c *Command) handleError(o *Output, err error) {
	if err != nil {
		c.RenderError(o, err)
	}

	if c.AutoExit {
		c.Exit(err)
	}
}

func (c *Command) Exit(err error) {
	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func (c *Command) Subcommand(name string) *Command {
	if err := c.init(); err != nil {
		return nil
	}

	return c.commands[name]
}

func (c *Command) HasSubcommands() bool {
	return len(c.Commands) > 0 || len(c.commands) > 0
}

func (c *Command) execute(i *Input, o *Output) error {
	if err := c.validate(); err != nil {
		return err
	}

	if c.hasFlag(i, "version") {
		if i.HasParameterFlag("--version", true) || i.HasParameterFlag("-V", true) {
			o.Writeln(c.version(), 0)
			return nil
		}
	}

	var command *Command
	var args []string
	var err error

	if len(i.Args) == 0 && c.PromptForCommand && c.HasSubcommands() && c.Run == nil && c.RunE == nil {
		command, err = c.promptForCommand(i, o)
	}

	if err != nil {
		return err
	} else if command == nil {
		command, args, err = c.findCommand(i.Args, &i.tokens)
	} else if command == c {
		os.Exit(1)
	}

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

	if c.CascadeNativeFlags && command.NativeFlags != nil {
		command.NativeFlags = c.NativeFlags
	}

	command.configuredIO = true

	wantsHelp := c.hasFlag(i, "help") && (i.HasParameterFlag("--help", true) || i.HasParameterFlag("-h", true))
	if wantsHelp && command != nil {
		command.printHelp(o)
		return nil
	}

	def, err := command.Definition()
	if err != nil {
		return err
	}

	err = i.Bind(def)
	if err != nil && !command.IgnoreValidationErrors {
		return err
	}

	// inspection, _ := i.Inspect(i.Args)
	// n, _ := def.Flag("name")
	// if inspection.FlagIsGiven(n) {
	// 	fmt.Print("GIVEN")
	// 	return nil
	// }
	// fmt.Print(inspection.String())
	// return nil

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
		Command:    command,
		Input:      i,
		Output:     o,
		definition: def,
		Args:       i.Args,
	}

	if command.RunE != nil {
		err = command.RunE(io)
	} else if command.Run != nil {
		command.Run(io)
	} else if command == c || command.HasSubcommands() {
		command.printHelp(o)
	} else {
		return errors.New("command must have a handle or subcommands")
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
	if err := c.init(); err != nil {
		return nil
	}

	cmds := make(map[string]*Command)
	for _, cmd := range c.commands {
		cmds[cmd.Name] = cmd
	}

	return cmds
}

func (c *Command) version() string {
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

func (c *Command) AddCommand(command *Command) error {
	if command == c {
		return errors.New("cannot add a command to itself")
	}

	if err := c.init(); err != nil {
		return err
	}

	if command.Name == "" {
		return errors.New("commands must have a name")
	}

	if _, exists := c.commands[command.Name]; exists {
		return fmt.Errorf("command \"%s\" already exists", command.Name)
	}

	c.commands[command.Name] = command

	for _, alias := range command.Aliases {
		if _, exists := c.commands[alias]; exists {
			return fmt.Errorf("command \"%s\" already exists", alias)
		}

		c.commands[alias] = command
	}

	command.SetParent(c)
	return command.validate()
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
		d, _ := c.Definition()
		c.synopsis[key] = strings.TrimSpace(fmt.Sprintf("%s %s", c.FullName(), d.Synopsis(short)))
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
	if help == "" && c.HasSubcommands() && slices.Contains(c.NativeFlags, "help") {
		help = fmt.Sprintf("Use \"%s [command] --help\" for more information about a command", c.FullName())
	}

	for i, placeholder := range placeholders {
		help = strings.ReplaceAll(help, placeholder, replacements[i])
	}

	return help
}

func (c *Command) validate() error {
	if c.validated {
		return nil
	}
	c.validated = true

	re := regexp.MustCompile("^[^:]+(:[^:]+)*")
	if !re.MatchString(c.Name) {
		return fmt.Errorf("command name \"%s\" is invalid", c.Name)
	}

	for _, alias := range c.Aliases {
		if !re.MatchString(alias) {
			return fmt.Errorf("command alias \"%s\" is invalid", alias)
		}
	}

	return nil
}

func (c *Command) RenderError(o *Output, err error) {
	o.Err(err)

	if c.runningCommand != nil {
		o.Writeln(
			fmt.Sprintf("<accent>%s</accent>", c.runningCommand.Synopsis(false)),
			VerbosityQuiet,
		)
	}
}

func (c *Command) hasFlag(i *Input, name string) bool {
	return i.HasFlag(name) || c.NativeFlags == nil || slices.Contains(c.NativeFlags, name)
}

func (c *Command) configureIO(i *Input, o *Output) {
	if c.configuredIO {
		return
	} else {
		c.configuredIO = true
	}

	if c.hasFlag(i, "ansi") {
		if i.HasParameterFlag("--ansi", true) {
			o.SetDecorated(true)
		} else if i.HasParameterFlag("--no-ansi", true) {
			o.SetDecorated(false)
		}
	}

	if c.hasFlag(i, "no-interaction") {
		if i.HasParameterFlag("--no-interaction", true) || i.HasParameterFlag("-n", true) {
			i.SetInteractive(false)
		}
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

	if c.hasFlag(i, "quiet") && (i.HasParameterFlag("--quiet", true) || i.HasParameterFlag("-q", true)) {
		o.SetVerbosity(VerbosityQuiet)
		shellVerbosity = -1
	} else if c.hasFlag(i, "verbose") {
		if i.HasParameterFlag("-vvv", true) {
			o.SetVerbosity(VerbosityDebug)
			shellVerbosity = 3
		} else if i.HasParameterFlag("-vv", true) {
			o.SetVerbosity(VerbosityVeryVerbose)
			shellVerbosity = 2
		} else if i.HasParameterFlag("-v", true) {
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

// TODO: use inspector?
func (c *Command) findCommand(args []string, tokens *[]string) (*Command, []string, error) {
	isOption := false
	argc := len(args)
	arguments := make([]string, 0)

	current := c
	definition, _ := current.Definition()
	toRemove := make([]string, 0)

	for idx, token := range args {
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
			if FlagAcceptsValue(flag) && !strings.HasPrefix(args[idx+1], "-") {
				isOption = true
			}

			continue
		}

		// Is value for option
		if isOption {
			isOption = false
			continue
		}

		current.init()
		cmd := current.commands[token]
		if cmd != nil {
			current = cmd
			definition, _ = current.Definition()
			toRemove = append(toRemove, token)
		} else {
			if len(definition.arguments) == 0 && current.HasSubcommands() {
				alternatives := c.findAlternatives(token, array.SortedKeys(current.commands))
				return nil, arguments[:idx+1], CommandNotFound(fmt.Sprintf("command \"%s\" does not exist", token), alternatives)
			}
		}

		// arguments = append(arguments, token)
	}

	if tokens != nil {
		for _, token := range toRemove {
			*tokens = array.Remove(*tokens, token)
		}
	}

	return current, nil, nil
}

func (c *Command) defaultInputDefinition() (*InputDefinition, error) {
	definition := &InputDefinition{}
	flags := make([]Flag, 0, 6)
	requested := c.NativeFlags

	if requested != nil && len(requested) == 0 {
		return definition, nil
	}

	all := requested == nil

	if all || slices.Contains(requested, "help") {
		var description string
		if c.HasSubcommands() {
			description = "Display help for a command or a given command"
		} else {
			description = "Display help for the command"
		}

		helpFlag := &BoolFlag{
			Name:        "help",
			Shortcuts:   []string{"h"},
			Description: description,
		}
		flags = append(flags, helpFlag)
	}

	if all || slices.Contains(requested, "version") {
		versionFlag := &BoolFlag{
			Name:        "version",
			Shortcuts:   []string{"V"},
			Description: "Display the command version",
		}
		flags = append(flags, versionFlag)
	}

	if all || slices.Contains(requested, "quiet") {
		quietFlag := &BoolFlag{
			Name:        "quiet",
			Shortcuts:   []string{"q"},
			Description: "Do not output any message",
		}
		flags = append(flags, quietFlag)
	}

	if all || slices.Contains(requested, "verbose") {
		verboseflag := &BoolFlag{
			Name:        "verbose",
			Shortcuts:   []string{"v", "vv", "vvv"},
			Description: "Increase the verbosity of messages: normal, verbose or debug",
		}
		flags = append(flags, verboseflag)
	}

	if all || slices.Contains(requested, "ansi") {
		ansiFlag := &BoolFlag{
			Name:        "ansi",
			Negatable:   true,
			Description: "Force (or disable --no-ansi) ANSI output",
		}
		flags = append(flags, ansiFlag)
	}

	if all || slices.Contains(requested, "no-interaction") {
		noInteractionFlag := &BoolFlag{
			Name:        "no-interaction",
			Shortcuts:   []string{"n"},
			Description: "Do not ask any interactive question",
		}
		flags = append(flags, noInteractionFlag)
	}

	err := definition.SetFlags(flags)

	return definition, err
}

func (c *Command) init() error {
	if c.initialized {
		return nil
	}

	c.initialized = true

	if c.commands == nil {
		c.commands = make(map[string]*Command)
	}

	for _, command := range c.Commands {
		command.SetParent(c)
		if err := c.AddCommand(command); err != nil {
			return err
		}
	}

	c.Commands = nil
	return nil
}

func (c *Command) Definition() (*InputDefinition, error) {
	var err error
	if c.definition == nil {
		nativeDefinition, e := c.defaultInputDefinition()
		if e != nil {
			err = e
		}

		c.definition = &InputDefinition{}
		if e := c.definition.SetArguments(c.Arguments); e != nil && err != nil {
			err = e
		}

		if e := c.definition.SetFlags(c.Flags); e != nil && err != nil {
			err = e
		}

		if e := c.definition.AddArguments(nativeDefinition.GetArguments()); e != nil && err != nil {
			err = e
		}

		if e := c.definition.AddFlags(nativeDefinition.GetFlags()); e != nil && err != nil {
			err = e
		}
	}

	return c.definition, err
}

func (c *Command) printHelp(output *Output) {
	if c.PrintHelpFunc != nil {
		c.PrintHelpFunc(output, c)
		return
	}

	d := TextDescriptor{output}
	d.DescribeCommand(c, &DescriptorOptions{})
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
		return fmt.Errorf("argument \"%s\" is missing a description", name)
	}

	q := strings.ToLower(string(desc[0])) + desc[1:]

	switch a := arg.(type) {
	case *StringArg:
		var answer string
		var err error
		if a.Options != nil {
			prompt := NewSearchPrompt(i, o, fmt.Sprintf("What is %s?", q), func(s string) SearchResult {
				s = strings.ToLower(s)
				opts := make([]string, 0, len(a.Options))
				for _, v := range a.Options {
					if strings.Contains(strings.ToLower(v), s) {
						opts = append(opts, v)
					}
				}
				return opts
			}, "")
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
		err = i.SetArgument(name, answer)
		if err != nil {
			return err
		}

		return nil
	case *ArrayArg:
		var answers []string
		var err error
		if len(a.Options) > 0 {
			// TODO: use multiselect prompt with search
			prompt := NewMultiSelectPrompt(i, o, fmt.Sprintf("What is %s?", q), a.Options, a.Value)
			prompt.Required = true
			answers, err = prompt.Render()
			if err != nil {
				return err
			}
		} else {
			prompt := NewArrayPrompt(i, o, fmt.Sprintf("What is %s?", q), nil)
			prompt.Required = true
			answers, err = prompt.Render()
			if err != nil {
				return err
			}
		}

		for _, answer := range answers {
			err = i.SetArgument(name, answer)
			if err != nil {
				return err
			}
		}

		i.Args = append(i.Args, answers...)

		return nil
	default:
		return errors.New("unsupported argument type")
	}
}

func (c *Command) Root() *Command {
	root := c
	for root.parent != nil {
		root = root.parent
	}
	return root
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

func (c *Command) promptForCommand(i *Input, o *Output) (*Command, error) {
	var target *Command = c
	var lastPrompt *SearchPrompt

	defer func() {
		if lastPrompt != nil {
			lastPrompt.Clear()
		}

		if target != c {
			o.Writeln(Dim(fmt.Sprintf("Running command \"<primary>%s</primary>\"", target.FullName())), 0)
			o.NewLine(1)
		}
	}()

	for {
		if err := target.init(); err != nil {
			return nil, err
		}

		if !target.HasSubcommands() {
			break
		}

		label := "Select the command to run"
		if target != nil {
			label = "Select the child command to run for <primary>" + target.FullName() + "</primary>"
		}

		options := make(map[string]string, len(target.commands))
		maxCommandNameLength := 0
		for _, cmd := range target.commands {
			if cmd.Hidden {
				continue
			}

			maxCommandNameLength = max(maxCommandNameLength, len(cmd.Name))
		}

		for _, cmd := range target.commands {
			if cmd.Hidden {
				continue
			}
			length := len(cmd.Name)
			indent := ""
			if length < maxCommandNameLength {
				indent = strings.Repeat(" ", maxCommandNameLength-length)
			}

			options[cmd.Name] = fmt.Sprintf("<primary>%s</primary>%s  %s", cmd.Name, indent, cmd.Description)
		}

		if lastPrompt != nil {
			lastPrompt.Clear()
			lastPrompt = nil
		}

		prompt := NewSearchPrompt(i, o, label, func(s string) SearchResult {
			if s == "" {
				return options
			}

			filtered := make([]string, 0, len(options))
			for _, option := range options {
				if strings.Contains(option, s) {
					filtered = append(filtered, option)
				}
			}

			return filtered
		}, "")
		prompt.Required = false
		lastPrompt = prompt

		answer, err := prompt.Render()
		if err != nil {
			return nil, err
		}

		if answer == "" {
			break
		}

		name := StripEscapeSequences(strings.Split(answer, " ")[0])

		cmd := target.Subcommand(name)
		if cmd == nil {
			return nil, fmt.Errorf("command \"%s\" not found", name)
		}

		target = cmd
	}

	return target, nil
}

func (c *Command) Subcommands() map[string]*Command {
	if err := c.init(); err != nil {
		return nil
	}

	cmds := make(map[string]*Command)
	for _, cmd := range c.commands {
		cmds[cmd.Name] = cmd
	}

	return cmds
}

func (c *Command) Flag(name string) (Flag, error) {
	return c.definition.Flag(name)
}

func (c *Command) Arg(name string) (Arg, error) {
	return c.definition.Argument(name)
}
