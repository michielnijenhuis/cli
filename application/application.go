package application

import (
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/michielnijenhuis/cli/command"
	"github.com/michielnijenhuis/cli/descriptor"
	"github.com/michielnijenhuis/cli/formatter"
	"github.com/michielnijenhuis/cli/helper"
	"github.com/michielnijenhuis/cli/helper/maps"
	"github.com/michielnijenhuis/cli/input"
	"github.com/michielnijenhuis/cli/output"
	"github.com/michielnijenhuis/cli/style"
	"github.com/michielnijenhuis/cli/terminal"
	"github.com/michielnijenhuis/cli/types"
)

type Application struct {
	Name           string
	Version        string
	LongVersion    string
	DefaultCommand string
	wantsHelp      bool
	CatchErrors    bool
	AutoExit       bool
	SingleCommand  bool
	initialized    bool
	runningCommand *command.Command
	definition     *input.InputDefinition
	commands       map[string]*command.Command
}

func (app *Application) Run() (exitCode int, err error) {
	return app.RunWith(nil, nil)
}

func (app *Application) RunWith(i input.InputInterface, o output.OutputInterface) (exitCode int, err error) {
	width, height, err := terminal.Size()
	if err == nil {
		os.Setenv("LINES", fmt.Sprint(height))
		os.Setenv("COLUMNS", fmt.Sprint(width))
	}

	if i == nil {
		i, err = input.NewArgvInput(nil, nil)
		if err != nil {
			panic("Failed to initialize application input.")
		}
	}

	if o == nil {
		o = output.NewConsoleOutput(0, true, nil)
	}

	if app.DefaultCommand == "" {
		app.DefaultCommand = "list"
	}

	if app.CatchErrors {
		defer func() {
			if r := recover(); r != nil {
				recoveredErr, isErr := r.(error)
				exitCode = 1

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
			}
		}()
	}

	app.configureIO(i, o)

	code, err := app.doRun(i, o)

	if err != nil {
		if !app.CatchErrors {
			return 1, err
		}

		app.RenderError(o, err)
		exitCode = 1
	} else {
		exitCode = code
	}

	if app.AutoExit {
		if exitCode > 255 {
			exitCode = 255
		}

		os.Exit(exitCode)
	}

	return exitCode, err
}

func (app *Application) doRun(i input.InputInterface, o output.OutputInterface) (int, error) {
	if i.HasParameterOption("--version", true) || i.HasParameterOption("-V", true) {
		o.Writeln(app.version(), 0)

		return 0, nil
	}

	// Makes ArgvInput.FirstArgument() able to distinguish an option from an argument.
	// Errors must be ignored, full binding/validation happens later when the command is known.
	i.Bind(app.Definition())
	i.Parse()

	name := app.commandName(i)
	if i.HasParameterOption("--help", true) || i.HasParameterOption("-h", true) {
		if name == "" {
			name = "help"
			params := map[string]input.InputType{"command_name": app.DefaultCommand}
			objectInput, err := input.NewObjectInput(params, nil)
			i = objectInput

			if err != nil {
				return 1, err
			}
		} else {
			app.wantsHelp = true
		}
	}

	if name == "" {
		name = app.DefaultCommand
		definition := app.Definition()
		definition.SetArguments(nil)
	}

	app.runningCommand = nil
	c, findCommandErr := app.Find(name)

	if findCommandErr != nil {
		notFound, ok := findCommandErr.(*command.CommandNotFoundError)
		var alternatives []string
		if ok {
			alternatives = notFound.Alternatives()
		}
		interactive := i.IsInteractive()

		if ok && len(alternatives) == 1 && interactive {
			o.Writeln("", 0)
			formattedBlock := formatter.FormatBlock([]string{fmt.Sprintf("command \"%s\" is not defined", name)}, "error", true)
			o.Writeln(formattedBlock+"\n", 0)

			style := style.NewStyle(i, o)
			alternative := alternatives[0]

			runAlternative, err := style.Confirm(fmt.Sprintf("Do you want to run \"%s\" instead?", alternative), false)
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

			if _, ok := findCommandErr.(types.ErrorWithAlternatives); ok && err == nil {
				outputter := o
				if consoleOutput, ok := outputter.(*output.ConsoleOutput); ok {
					outputter = consoleOutput
				}

				d := descriptor.NewTextDescriptor(outputter)
				d.DescribeApplication(app, descriptor.NewDescriptorOptions(namespace, false, false, 0))
			}

			return 1, findCommandErr
		}
	}

	app.runningCommand = c
	exitCode, runCommandErr := c.Run(i, o)
	app.runningCommand = nil

	return exitCode, runCommandErr
}

func (app *Application) SetDefinition(definition *input.InputDefinition) {
	app.definition = definition
}

func (app *Application) Definition() *input.InputDefinition {
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
			return fmt.Sprintf("%s <highlight>%s</highlight>", app.Name, app.Version)
		}

		return app.Name
	}

	return ""
}

func (app *Application) AddCommands(commands []*command.Command) {
	for _, c := range commands {
		app.Add(c)
	}
}

func (app *Application) Add(commands ...*command.Command) {
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

		if c.Aliases != nil {
			for _, alias := range c.Aliases {
				app.commands[alias] = c
			}
		}
	}
}

func (app *Application) Get(name string) (*command.Command, error) {
	app.init()

	if !app.Has(name) {
		return nil, command.NotFound(fmt.Sprintf("The command \"%s\" does not exist.", name), nil)
	}

	if app.commands[name] == nil {
		return nil, command.NotFound(fmt.Sprintf("The \"%s\" command cannot be found because it is registered under multiple names. Make sure you don't set a different name.", name), nil)
	}

	c := app.commands[name]

	if app.wantsHelp {
		app.wantsHelp = false

		helpCommand, _ := app.Get("help")
		helpCommand.SetMeta(map[string]*command.Command{
			"command": c,
		})

		return helpCommand, nil
	}

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
		message := fmt.Sprintf("There are no commands defined in the \"%s\" namespace.", namespace)
		alternatives := app.findAlternatives(namespace, allNamespaces)

		if len(alternatives) > 0 {
			if len(alternatives) == 1 {
				message += "\n\nDid you mean this?\n    "
			} else {
				message += "\n\nDid you mean one of these?\n    "
			}

			message += strings.Join(alternatives, "\n    ")
		}

		return "", NamespaceNotFound(message, alternatives)
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

func (app *Application) Find(name string) (*command.Command, error) {
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

		message := fmt.Sprintf("command \"%s\" is not defined", name)
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
				message += "\n\nDid you mean this?\n    "
			} else {
				message += "\n\nDid you mean one of these?\n    "
			}

			message += strings.Join(alternatives, "\n    ")
		}

		return nil, command.NotFound(message, alternatives)
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
		terminalWidth, _ := terminal.Width()
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

				abbrev = helper.PadStart(cmd, maxLen, ' ') + " " + app.commands[cmd].Description

				if helper.Width(abbrev) > usableWidth {
					filteredAbbrevs = append(filteredAbbrevs, abbrev[:usableWidth-3]+"...")
				} else {
					filteredAbbrevs = append(filteredAbbrevs, abbrev)
				}
			}

			if len(commands) > 1 {
				suggestions := app.abbreviationSuggestions(filteredAbbrevs)
				return nil, command.NotFound(fmt.Sprintf("command \"%s\" is ambiguous.\nDid you mean one of these?\n%s", name, suggestions), commands)
			}
		}
	}

	cmd, err := app.Get(commands[0])
	if err != nil {
		return nil, err
	}

	if cmd.Hidden {
		return nil, command.NotFound(fmt.Sprintf("the command \"%s\" does not exist", name), nil)
	}

	return cmd, nil
}

func (app *Application) All(namespace string) map[string]*command.Command {
	app.init()

	if namespace == "" {
		return app.commands
	}

	re := regexp.MustCompile(`\:`)
	commands := make(map[string]*command.Command)

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

func (app *Application) RenderError(o output.OutputInterface, err error) {
	o.Writeln("", output.VERBOSITY_QUIET)

	app.doRenderError(o, err)

	if app.runningCommand != nil {
		o.Writeln(
			fmt.Sprintf("<highlight>%s %s</highlight>", app.Name, app.runningCommand.Synopsis(false)),
			output.VERBOSITY_QUIET,
		)
		o.Writeln("", output.VERBOSITY_QUIET)
	}
}

func (app *Application) doRenderError(o output.OutputInterface, err error) {
	message := strings.TrimSpace(err.Error())
	length := 0
	width, _ := terminal.Width()
	lines := make([]string, 0)
	linesLength := make([]int, 0)
	messageLines := strings.Split(strings.ReplaceAll(message, "\r\n", "\n"), "\n")

	for i := 0; i < len(messageLines); i++ {
		message := messageLines[i]
		splitMessage := splitStringByWidth(message, width-4)

		for _, line := range splitMessage {
			lineLength := len(line) + 4
			lines = append(lines, line)
			linesLength = append(linesLength, lineLength)

			length = int(math.Max(float64(lineLength), float64(length)))
		}
	}

	messages := make([]string, 0)
	emptyLine := fmt.Sprintf("<error>%s</error>", strings.Repeat(" ", length))
	messages = append(messages, emptyLine)

	for i, line := range lines {
		formattedLine := fmt.Sprintf("<error>  %s  %s</error>", formatter.Escape(line), strings.Repeat(" ", length-linesLength[i]))
		messages = append(messages, formattedLine)
	}

	messages = append(messages, emptyLine, "")

	o.Writelns(messages, output.VERBOSITY_QUIET)
}

func (app *Application) configureIO(i input.InputInterface, o output.OutputInterface) {
	if i.HasParameterOption("--ansi", true) {
		o.SetDecorated(true)
	} else if i.HasParameterOption("--no-ansi", true) {
		o.SetDecorated(false)
	}

	if i.HasParameterOption("--no-interaction", true) || i.HasParameterOption("-n", true) {
		i.SetInteractive(false)
	}

	shellVerbosity, err := strconv.Atoi(os.Getenv("SHELL_VERBOSITY"))
	if err != nil {
		shellVerbosity = 0
	}

	switch shellVerbosity {
	case -1:
		o.SetVerbosity(output.VERBOSITY_QUIET)
	case 1:
		o.SetVerbosity(output.VERBOSITY_VERBOSE)
	case 2:
		o.SetVerbosity(output.VERBOSITY_VERY_VERBOSE)
	case 3:
		o.SetVerbosity(output.VERBOSITY_DEBUG)
	default:
		shellVerbosity = 0
	}

	if i.HasParameterOption("--quiet", true) || i.HasParameterOption("-q", true) {
		o.SetVerbosity(output.VERBOSITY_QUIET)
		shellVerbosity = -1
	} else {
		if i.HasParameterOption("-vvv", true) ||
			i.HasParameterOption("--verbose=3", true) ||
			i.ParameterOption("--verbose", false, true) == "3" {
			o.SetVerbosity(output.VERBOSITY_DEBUG)
			shellVerbosity = 3
		} else if i.HasParameterOption("-vv", true) ||
			i.HasParameterOption("--verbose=2", true) ||
			i.ParameterOption("--verbose", false, true) == "2" {
			o.SetVerbosity(output.VERBOSITY_VERBOSE)
			shellVerbosity = 2
		} else if i.HasParameterOption("-v", true) ||
			i.HasParameterOption("--verbose=1", true) ||
			i.HasParameterOption("--verbose", true) {
			o.SetVerbosity(output.VERBOSITY_VERBOSE)
			shellVerbosity = 1
		}
	}

	if shellVerbosity == -1 {
		i.SetInteractive(false)
	}

	os.Setenv("SHELL_VERBOSITY", fmt.Sprint(shellVerbosity))
}

func (app *Application) commandName(input input.InputInterface) string {
	if app.SingleCommand {
		return app.DefaultCommand
	}

	first := input.FirstArgument()
	str, ok := first.(string)
	if ok {
		return str
	}

	arr, ok := first.([]string)
	if ok && len(arr) > 0 {
		return arr[0]
	}

	panic("Failed to retrieve first argument from input.")
}

func (app *Application) defaultInputDefinition() *input.InputDefinition {
	commandArgument := input.NewInputArgument("command", input.INPUT_ARGUMENT_REQUIRED, "The command to execute")
	arguments := []*input.InputArgument{commandArgument}

	helpOption := input.NewInputOption("--help", "-h", input.INPUT_OPTION_BOOLEAN, fmt.Sprintf("Display help for the given command, or the <highlight>%s</highlight> command (if no command is given)", app.DefaultCommand))
	quietOption := input.NewInputOption("--quiet", "-q", input.INPUT_OPTION_BOOLEAN, "Do not output any message")
	verboseoption := input.NewInputOption("--verbose", "-v|vv|vvv", input.INPUT_OPTION_BOOLEAN, "Increase the verbosity of messages: normal (1), verbose (2) or debug (3)")
	versionOption := input.NewInputOption("--version", "-V", input.INPUT_OPTION_BOOLEAN, "Display this application version")
	ansiOption := input.NewInputOption("--ansi", "", input.INPUT_OPTION_NEGATABLE, "Force (or disable --no-ansi) ANSI output")
	noInteractionOption := input.NewInputOption("--no-interaction", "-n", input.INPUT_OPTION_BOOLEAN, "Do not ask any interactive question")

	options := []*input.InputOption{
		helpOption,
		quietOption,
		verboseoption,
		versionOption,
		ansiOption,
		noInteractionOption,
	}

	definition := input.NewInputDefinition(arguments, options)

	return definition
}

func (app *Application) defaultCommands() []*command.Command {
	return []*command.Command{
		app.newHelpCommand(),
		app.newListCommand(),
	}
}

func (app *Application) newHelpCommand() *command.Command {
	c := &command.Command{
		Name:        "help",
		Description: "Display help for a command",
		Help: `The <highlight>%command.name%</highlight> command displays help for a given command:

  <highlight>%command.full_name% list</highlight>

To display the list of available commands, please use the <highlight>list</highlight> command.
`,
		IgnoreValidationErrors: true,
		Handle: func(self *command.Command) (int, error) {
			var target *command.Command = nil

			meta := self.Meta()
			if metaMap, ok := meta.(map[string]*command.Command); ok {
				target = metaMap["command"]
			}

			if target == nil {
				commandName, err := self.StringArgument("command_name")
				if err != nil {
					return 1, err
				}

				targetCommand, err := app.Find(commandName)
				if err != nil {
					return 1, err
				}

				target = targetCommand
			}

			d := descriptor.NewTextDescriptor(self.Output())
			raw, _ := self.BoolOption("raw")
			d.DescribeCommand(target, descriptor.NewDescriptorOptions("", raw, false, 0))

			return 0, nil
		},
	}

	c.SetDefinition(input.NewInputDefinition(
		[]*input.InputArgument{
			input.NewInputArgument("command_name", input.INPUT_ARGUMENT_OPTIONAL, "The command name").SetDefaultValue("help"),
		},
		[]*input.InputOption{
			input.NewInputOption("raw", "", input.INPUT_OPTION_BOOLEAN, "To output raw command help"),
		},
	))

	return c
}

func (app *Application) newListCommand() *command.Command {
	c := &command.Command{
		Name:        "list",
		Description: "List commands",
		Help: `The <highlight>%command.name%</highlight> command lists all commands:

  <highlight>%command.full_name%</highlight>

You can also display the commands for a specific namespace:

  <highlight>%command.full_name% test</highlight>

It's also possible to get raw list of commands (useful for embedding command runner):

  <highlight>%command.full_name% --raw</highlight>`,
		IgnoreValidationErrors: true,
		Handle: func(self *command.Command) (int, error) {
			d := descriptor.NewTextDescriptor(self.Output())
			raw, _ := self.BoolOption("raw")
			short, _ := self.BoolOption("short")
			namespace, _ := self.StringArgument("namespace")
			d.DescribeApplication(app, descriptor.NewDescriptorOptions(namespace, raw, short, 0))

			return 0, nil
		},
	}

	c.SetDefinition(input.NewInputDefinition(
		[]*input.InputArgument{
			input.NewInputArgument("namespace", input.INPUT_ARGUMENT_OPTIONAL, "The namespace name"),
		},
		[]*input.InputOption{
			input.NewInputOption("raw", "", input.INPUT_OPTION_BOOLEAN, "To output raw command list"),
			input.NewInputOption("short", "", input.INPUT_OPTION_BOOLEAN, "To skip describing commands' arguments"),
		},
	))

	return c
}

func (app *Application) abbreviationSuggestions(abbrevs []string) string {
	return "    " + strings.Join(abbrevs, "\n    ")
}

func (app *Application) ExtractNamespace(name string, limit int) string {
	parts := strings.Split(name, ":")[1:]

	if limit < 0 {
		return strings.Join(parts, ":")
	}

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
					alternatives[collectionName] = alternatives[collectionName] + lev
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
				alternatives[item] = alternatives[item] - lev
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
		app.DefaultCommand = "list"
	}

	if app.commands == nil {
		app.commands = make(map[string]*command.Command)
	}

	for _, command := range app.defaultCommands() {
		app.Add(command)
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

func splitStringByWidth(s string, w int) []string {
	if w < 1 {
		w = 1
	}

	result := make([]string, 0)
	for i := 0; i < len(s); i += w {
		m := min(i+w, len(s))
		result = append(result, s[i:m])
	}

	return result
}
