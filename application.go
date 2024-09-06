package cli

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
	"github.com/michielnijenhuis/cli/helper/maps"
)

type Application struct {
	Name           string
	Version        string
	LongVersion    string
	DefaultCommand string
	Debug          bool
	Commands       []*Command
	CatchErrors    bool
	AutoExit       bool
	SingleCommand  bool
	initialized    bool
	runningCommand *Command
	code           int
	definition     *InputDefinition
	commands       map[string]*Command
}

func (app *Application) Run(args ...string) error {
	return app.RunWith(NewInput(args...), nil)
}

func (app *Application) RunWith(i *Input, o *Output) (err error) {
	width, height, err := TerminalSize()
	if err == nil {
		os.Setenv("LINES", fmt.Sprint(height))
		os.Setenv("COLUMNS", fmt.Sprint(width))
	}

	if !app.Debug {
		if _, ok := os.LookupEnv("CLI_DEBUG"); ok {
			app.Debug = true
		}
	}

	if i == nil {
		i = NewInput()
	}

	if o == nil {
		o = NewOutput(i)
	}

	var shouldReturn bool

	if app.AutoExit {
		defer os.Exit(app.code)
	}

	if app.CatchErrors {
		defer func() {
			if r := recover(); r != nil {
				recoveredErr, isErr := r.(error)
				app.code = 1

				if !isErr {
					msg, isStr := r.(string)
					if isStr && msg != "" {
						recoveredErr = errors.New(msg)
					} else {
						recoveredErr = errors.New("an unknown error occurred")
					}
				}

				err = recoveredErr
				app.RenderError(o, err)
				shouldReturn = true
			}
		}()
	}

	app.configureIO(i, o)

	code, err := app.doRun(i, o)
	app.code = code

	if shouldReturn {
		return
	}

	if err != nil {
		if app.code <= 0 {
			app.code = 1
		}

		if !app.CatchErrors {
			return err
		}

		app.RenderError(o, err)
	}

	if app.AutoExit {
		if app.code > 255 {
			app.code = 255
		}
	}

	return
}

func (app *Application) doRun(i *Input, o *Output) (int, error) {
	if i.HasParameterFlag("--version", true) || i.HasParameterFlag("-V", true) {
		o.Writeln(app.version(), 0)

		return 0, nil
	}

	if app.Commands != nil {
		app.Add(app.Commands...)
		app.Commands = nil
	}

	name := app.commandName(i)
	if i.HasParameterFlag("--help", true) || i.HasParameterFlag("-h", true) || len(i.Args) == 0 {
		if name != "" {
			c, err := app.Find(name)
			if err != nil {
				return 1, err
			}
			app.printHelp(o, c)
			return 0, nil
		}

		app.printHelp(o, nil)
		return 0, nil
	}

	// Makes ArgvInput.FirstArgument() able to distinguish a flag from an argument.
	// Errors must be ignored, full binding/validation happens later when the command is known.
	i.Bind(app.Definition())

	if name == "" {
		name = app.DefaultCommand
		definition := app.Definition()
		definition.SetArguments(nil)
	}

	app.runningCommand = nil
	c, findCommandErr := app.Find(name)

	if findCommandErr != nil {
		notFound, ok := findCommandErr.(*CommandNotFoundError)
		var alternatives []string
		if ok {
			alternatives = notFound.Alternatives()
		}
		interactive := i.IsInteractive()

		if ok && len(alternatives) == 1 && interactive {
			o.Writeln("", 0)
			formattedBlock := FormatBlock([]string{fmt.Sprintf("command \"%s\" is not defined", name)}, "error", true)
			o.Writeln(formattedBlock, 0)

			alternative := alternatives[0]

			runAlternative, err := o.Confirm(fmt.Sprintf("Do you want to run \"%s\" instead?", alternative), false)
			if err != nil {
				return 1, err
			}

			if !runAlternative {
				return 1, nil
			}

			c, findCommandErr = app.Find(alternative)
			if findCommandErr != nil {
				return 1, findCommandErr
			}
		} else {
			namespace, err := app.FindNamespace(name)

			if _, ok := findCommandErr.(ErrorWithAlternatives); ok && err == nil {
				d := TextDescriptor{o}
				d.DescribeApplication(app, &DescriptorOptions{
					namespace:  namespace,
					rawText:    false,
					short:      false,
					totalWidth: 0,
				})
			}

			return 1, findCommandErr
		}
	}

	app.runningCommand = c
	cDebug := c.Debug
	c.Debug = app.Debug
	exitCode, runCommandErr := c.ExecuteWith(i, o)
	app.runningCommand = nil
	c.Debug = cDebug

	return exitCode, runCommandErr
}

func (app *Application) SetDefinition(definition *InputDefinition) {
	app.definition = definition
}

func (app *Application) Definition() *InputDefinition {
	if app.definition == nil {
		app.definition = app.defaultInputDefinition()
	}

	if app.SingleCommand {
		inputDefinition := app.definition
		inputDefinition.SetArguments(nil)

		return inputDefinition
	}

	return app.definition
}

func (app *Application) Help() string {
	version := app.version()

	if version != "" {
		return version
	}

	return "Console Tool"
}

func (app *Application) version() string {
	if app.LongVersion != "" {
		return app.LongVersion
	}

	if app.Name != "" {
		if app.Version != "" {
			return fmt.Sprintf("%s <accent>%s</accent>", app.Name, app.Version)
		}

		return app.Name
	}

	return ""
}

func (app *Application) AddCommands(commands []*Command) {
	for _, c := range commands {
		app.Add(c)
	}
}

func (app *Application) Add(commands ...*Command) {
	app.init()

	for _, c := range commands {
		c.SetApplicationDefinition(app.Definition())

		if !c.IsEnabled() {
			c.SetApplicationDefinition(nil)
			continue
		}

		if c.Name == "" {
			panic("Commands must have a name.")
		}

		app.commands[c.Name] = c

		for _, alias := range c.Aliases {
			app.commands[alias] = c
		}
	}
}

func (app *Application) Get(name string) (*Command, error) {
	app.init()

	if !app.Has(name) {
		return nil, CommandNotFound(fmt.Sprintf("The command \"%s\" does not exist.", name), nil)
	}

	if app.commands[name] == nil {
		return nil, CommandNotFound(fmt.Sprintf("The \"%s\" command cannot be found because it is registered under multiple names. Make sure you don't set a different name.", name), nil)
	}

	c := app.commands[name]

	return c, nil
}

func (app *Application) Has(name string) bool {
	app.init()

	return app.commands[name] != nil
}

func (app *Application) Namespaces() []string {
	namespacesMap := make(map[string]int)

	for _, command := range app.All("") {
		if command.Hidden || command.Name == "" {
			continue
		}

		for _, namespace := range app.extractAllNamespace(command.Name) {
			namespacesMap[namespace] = 0
		}

		if command.Aliases != nil {
			for _, alias := range command.Aliases {
				namespacesMap[app.ExtractNamespace(alias, -1)] = 0
			}
		}
	}

	namespaces := make([]string, 0, len(namespacesMap))
	for ns := range namespacesMap {
		namespaces = append(namespaces, ns)
	}

	return namespaces
}

func (app *Application) FindNamespace(namespace string) (string, error) {
	allNamespaces := app.Namespaces()

	parts := strings.Split(namespace, ":")
	re := regexp.MustCompile(`[-\/\\^$*+?.()|[\]{}]`)
	for i, part := range parts {
		parts[i] = re.ReplaceAllString(part, "$1")
	}

	expr := regexp.MustCompile("^" + strings.Join(allNamespaces, "[^:]*:") + "[^:]*")
	namespaces := make([]string, 0, len(allNamespaces))
	for _, ns := range allNamespaces {
		if expr.MatchString(ns) {
			namespaces = append(namespaces, ns)
		}
	}

	if len(namespaces) == 0 {
		var sb strings.Builder
		_, err := sb.WriteString(fmt.Sprintf("There are no commands defined in the \"%s\" namespace.", namespace))
		if err != nil {
			return "", err
		}

		alternatives := app.findAlternatives(namespace, allNamespaces)

		if len(alternatives) > 0 {
			if len(alternatives) == 1 {
				_, err := sb.WriteString("\n\nDid you mean this?\n    ")
				if err != nil {
					return "", err
				}
			} else {
				_, err := sb.WriteString("\n\nDid you mean one of these?\n    ")
				if err != nil {
					return "", err
				}
			}

			_, err := sb.WriteString(strings.Join(alternatives, "\n    "))
			if err != nil {
				return "", err
			}
		}

		return "", NamespaceNotFound(sb.String(), alternatives)
	}

	var exact bool
	for _, ns := range namespaces {
		if ns == namespace {
			exact = true
			break
		}
	}

	if len(namespaces) > 1 && !exact {
		return "", NamespaceNotFound(fmt.Sprintf("The namespace \"%s\" is ambiguous.\nDid you mean one of these?\n%s", namespace, app.abbreviationSuggestions(namespaces)), namespaces)
	}

	if exact {
		return namespace, nil
	}

	return namespaces[0], nil
}

func (app *Application) Find(name string) (*Command, error) {
	app.init()

	for _, command := range app.commands {
		if command.Aliases != nil {
			for _, alias := range command.Aliases {
				if !app.Has(alias) {
					app.commands[alias] = command
				}
			}
		}
	}

	if app.Has(name) {
		return app.Get(name)
	}

	parts := strings.Split(name, ":")
	re := regexp.MustCompile(`[-/\\^$*+?.()|[\]{}]`)
	for i, part := range parts {
		parts[i] = re.ReplaceAllString(part, "$&")
	}
	expr := strings.Join(parts, "[^:]*:") + "[^:]*"
	re2 := regexp.MustCompile(expr)

	commands := make([]string, 0, len(app.commands))
	for cmd := range app.commands {
		if re2.MatchString(cmd) {
			commands = append(commands, cmd)
		}
	}

	if len(commands) == 0 {
		caseInsensitiveRegex := regexp.MustCompile("(i?)^" + expr)
		commands = make([]string, 0, len(app.commands))
		for cmd := range app.commands {
			if caseInsensitiveRegex.MatchString(cmd) {
				commands = append(commands, cmd)
			}
		}
	}

	// if no commands matched or we just matched namespaces
	grepRegex := regexp.MustCompile("(i?){^" + expr + "}")
	var grepFilteredCommands []string
	if len(commands) > 0 {
		grepFilteredCommands = make([]string, 0, len(commands))
		for _, cmd := range commands {
			if grepRegex.MatchString(cmd) {
				grepFilteredCommands = append(grepFilteredCommands, cmd)
			}
		}
	}

	if len(commands) == 0 || (grepFilteredCommands != nil && len(grepFilteredCommands) < 1) {
		pos := strings.Index(name, ":")
		if pos != -1 {
			_, err := app.FindNamespace(name[0:pos])
			if err != nil {
				return nil, err
			}
		}

		var sb strings.Builder
		if _, err := sb.WriteString(fmt.Sprintf("command \"%s\" is not defined", name)); err != nil {
			return nil, err
		}

		alternatives := app.findAlternatives(name, maps.Keys(app.commands))
		if len(alternatives) > 0 {
			var ptr int
			for _, v := range alternatives {
				cmd, err := app.Get(v)
				if err == nil && !cmd.Hidden {
					alternatives[ptr] = v
					ptr++
				}
			}
			alternatives = alternatives[:ptr]

			if len(alternatives) == 1 {
				if _, err := sb.WriteString("\n\nDid you mean this?\n    "); err != nil {
					return nil, err
				}
			} else {
				if _, err := sb.WriteString("\n\nDid you mean one of these?\n    "); err != nil {
					return nil, err
				}
			}

			if _, err := sb.WriteString(strings.Join(alternatives, "\n    ")); err != nil {
				return nil, err
			}
		}

		return nil, CommandNotFound(sb.String(), alternatives)
	}

	aliases := make(map[string]string)
	if len(commands) > 0 {
		commandsMap := make(map[string]int)
		for _, cmd := range commands {
			item, exists := app.commands[cmd]
			if !exists {
				continue
			}
			aliases[cmd] = item.Name

			if item.Name == cmd {
				commandsMap[cmd] = 0
				continue
			}

			if item.Name == "" {
				continue
			}

			for _, c := range commands {
				if c == item.Name {
					continue
				}
			}

			commandsMap[cmd] = 0
		}

		newSlice := make([]string, 0, len(commandsMap))
		for k := range commandsMap {
			newSlice = append(newSlice, k)
		}
		commands = newSlice
	}

	if len(commands) > 0 {
		terminalWidth, _ := TerminalWidth()
		usableWidth := terminalWidth - 10
		abbrevs := commands
		maxLen := 0

		for _, abbrev := range abbrevs {
			maxLen = max(maxLen, helper.Width(abbrev))

			filteredAbbrevs := make([]string, 0, len(abbrevs))
			for i, cmd := range commands {
				if app.commands[cmd].Hidden {
					commands[i] = commands[len(commands)-1]
					commands = commands[:len(commands)-1]
					continue
				}

				abbrev = PadStart(cmd, maxLen, "") + " " + app.commands[cmd].Description

				if helper.Width(abbrev) > usableWidth {
					filteredAbbrevs = append(filteredAbbrevs, abbrev[:usableWidth-3]+"...")
				} else {
					filteredAbbrevs = append(filteredAbbrevs, abbrev)
				}
			}

			if len(commands) > 1 {
				suggestions := app.abbreviationSuggestions(filteredAbbrevs)
				return nil, CommandNotFound(fmt.Sprintf("command \"%s\" is ambiguous.\nDid you mean one of these?\n%s", name, suggestions), commands)
			}
		}
	}

	cmd, err := app.Get(commands[0])
	if err != nil {
		return nil, err
	}

	if cmd.Hidden {
		return nil, CommandNotFound(fmt.Sprintf("the command \"%s\" does not exist", name), nil)
	}

	return cmd, nil
}

func (app *Application) All(namespace string) map[string]*Command {
	app.init()

	if namespace == "" {
		return app.commands
	}

	re := regexp.MustCompile(`\:`)
	commands := make(map[string]*Command)

	for name, command := range app.commands {
		limit := len(re.FindAllStringSubmatchIndex(name, -1)) + 1

		if namespace == app.ExtractNamespace(name, limit) {
			commands[name] = command
		}
	}

	return commands
}

func (app *Application) Abbreviations(names []string) map[string][]string {
	abbrevs := make(map[string][]string)
	for _, name := range names {
		for len := len(name); len > 0; len-- {
			abbrev := name[0:len]
			arr := abbrevs[abbrev]
			if arr == nil {
				arr = []string{name}
			} else {
				arr = append(arr, name)
			}
			abbrevs[abbrev] = arr
		}
	}

	return abbrevs
}

func (app *Application) RenderError(o *Output, err error) {
	theme, _ := GetTheme("error")

	if theme.Padding {
		o.Writeln("", VerbosityQuiet)
	}

	o.Err(err)

	if !theme.Padding {
		o.Writeln("", VerbosityQuiet)
	}

	if app.runningCommand != nil {
		o.Writeln(
			fmt.Sprintf("<accent>%s %s</accent>", app.Name, app.runningCommand.Synopsis(false)),
			VerbosityQuiet,
		)

		if theme.Padding {
			o.Writeln("", VerbosityQuiet)
		}
	}
}

func (app *Application) configureIO(i *Input, o *Output) {
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
			o.SetVerbosity(VerbosityVerbose)
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

func (app *Application) commandName(input *Input) string {
	if app.SingleCommand {
		return app.DefaultCommand
	}

	return input.FirstArgument()
}

func (app *Application) defaultInputDefinition() *InputDefinition {
	commandArgument := &StringArg{
		Name:        "command",
		Required:    true,
		Description: "The command to execute",
	}

	arguments := []Arg{commandArgument}

	var helpDescription string
	if app.SingleCommand {
		helpDescription = fmt.Sprintf("Display help for the <accent>%s</accent> command", app.DefaultCommand)
	} else {
		helpDescription = "Display help the application or a given command"
	}

	helpFlag := &BoolFlag{
		Name:        "help",
		Shortcuts:   []string{"h"},
		Description: helpDescription,
	}
	versionFlag := &BoolFlag{
		Name:        "version",
		Shortcuts:   []string{"V"},
		Description: "Display the application version",
	}
	quietFlag := &BoolFlag{
		Name:        "quiet",
		Shortcuts:   []string{"q"},
		Description: "Do not output any message",
	}
	verboseflag := &OptionalStringFlag{
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
	definition.SetArguments(arguments)
	definition.SetFlags(flags)

	return definition
}

func (app *Application) printHelp(output *Output, command *Command) {
	if command == nil {
		d := TextDescriptor{output}
		d.DescribeApplication(app, &DescriptorOptions{})
		return
	}

	d := TextDescriptor{output}
	d.DescribeCommand(command, &DescriptorOptions{})
}

func (app *Application) abbreviationSuggestions(abbrevs []string) string {
	return "    " + strings.Join(abbrevs, "\n    ")
}

func (app *Application) ExtractNamespace(name string, limit int) string {
	parts := strings.Split(name, ":")
	parts = parts[0 : len(parts)-1]

	if limit < 0 {
		return strings.Join(parts, ":")
	}

	limit = max(len(parts), limit)

	return strings.Join(parts[0:limit], ":")
}

func (app *Application) findAlternatives(name string, collection []string) []string {
	treshold := int(1e3)
	alternatives := make(map[string]int)
	collectionParts := make(map[string][]string)

	for _, item := range collection {
		collectionParts[item] = strings.Split(item, ":")
	}

	slice := strings.Split(name, ":")
	for i := 0; i < len(slice); i++ {
		subname := slice[i]
		for collectionName, parts := range collectionParts {
			_, exists := alternatives[collectionName]

			if parts[i] == "" && exists {
				alternatives[collectionName] += treshold
				continue
			} else if parts[i] == "" {
				continue
			}

			lev := levenshtein(subname, parts[i])
			if lev <= len(subname)/3 || (subname != "" && strings.Contains(parts[i], subname)) {
				if exists {
					alternatives[collectionName] += lev
				} else {
					alternatives[collectionName] += treshold
				}
			} else if exists {
				alternatives[collectionName] += treshold
			}
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

func (app *Application) SetDefaultCommand(commandName string, isSingleCommand bool) error {
	for strings.HasPrefix(commandName, "|") {
		commandName = commandName[1:]
	}

	app.DefaultCommand = strings.Split(commandName, "|")[0]

	if app.SingleCommand {
		_, e := app.Find(commandName)
		app.SingleCommand = true

		if e != nil {
			return e
		}
	}

	return nil
}

func (app *Application) extractAllNamespace(name string) []string {
	parts := strings.SplitN(name, ":", 1)
	namespaces := make([]string, 0, len(parts))

	for _, part := range parts {
		if len(namespaces) > 0 {
			namespaces = append(namespaces, namespaces[len(namespaces)-1]+":"+part)
		} else {
			namespaces = append(namespaces, part)
		}
	}

	return namespaces
}

func (app *Application) init() {
	if app.initialized {
		return
	}

	app.initialized = true

	if app.DefaultCommand == "" {
		app.DefaultCommand = "help"
	}

	if app.commands == nil {
		app.commands = make(map[string]*Command)
	}
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

func (app *Application) ExitCode() int {
	return app.code
}
