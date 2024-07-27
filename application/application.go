package application

import (
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	Command "github.com/michielnijenhuis/cli/command"
	Descriptor "github.com/michielnijenhuis/cli/descriptor"
	Error "github.com/michielnijenhuis/cli/error"
	Formatter "github.com/michielnijenhuis/cli/formatter"
	Input "github.com/michielnijenhuis/cli/input"
	Output "github.com/michielnijenhuis/cli/output"
	Terminal "github.com/michielnijenhuis/cli/terminal"
)

type Application struct {
	name           string
	version        string
	defaultCommand string
	wantsHelp      bool
	catchErrors    bool
	autoExit       bool
	singleCommand  bool
	initialized    bool
	runningCommand *Command.Command
	definition     *Input.InputDefinition
	commands       map[string]*Command.Command
}

func NewApplication(name string, version string) *Application {
	if name == "" {
		name = "UNKNOWN"
	}

	if version == "" {
		version = "UNKNOWN"
	}

	return &Application{
		name:           name,
		version:        version,
		defaultCommand: "list",
		wantsHelp:      false,
		catchErrors:    false,
		autoExit:       true,
		singleCommand:  false,
		initialized:    false,
		runningCommand: nil,
		definition:     nil,
		commands:       make(map[string]*Command.Command),
	}
}

func (app *Application) Run(input Input.InputInterface, output Output.OutputInterface) (exitCode int, err error) {
	width, height, err := Terminal.GetSize()
	if err == nil {
		os.Setenv("LINES", fmt.Sprint(height))
		os.Setenv("COLUMNS", fmt.Sprint(width))
	}

	if input == nil {
		input, err = Input.NewArgvInput(nil, nil)
		if err != nil {
			panic("Failed to initialize application input.")
		}
	}

	if output == nil {
		output = Output.NewConsoleOutput(0, true, nil)
	}

	if app.catchErrors {
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
				app.RenderError(output, err)
			}
		}()
	}

	app.configureIO(input, output)

	code, err := app.doRun(input, output)

	if err != nil {
		if !app.catchErrors {
			return 1, err
		}

		app.RenderError(output, err)
		exitCode = 1
	} else {
		exitCode = code
	}

	if app.autoExit {
		if exitCode > 255 {
			exitCode = 255
		}

		os.Exit(exitCode)
	}

	return exitCode, err
}

func (app *Application) doRun(input Input.InputInterface, output Output.OutputInterface) (int, error) {
	if input.HasParameterOption("--version", true) || input.HasParameterOption("-V", true) {
		output.Writeln(app.getLongVersion(), 0)

		return 0, nil
	}

	// Makes ArgvInput.GetFirstArgument() able to distinguish an option from an argument.
	// Errors must be ignored, full binding/validation happens later when the command is known.
	input.Bind(app.GetDefinition())
	input.Parse()

	name := app.getCommandName(input)
	if input.HasParameterOption("--help", true) || input.HasParameterOption("-h", true) {
		if name == "" {
			name = "help"
			params := map[string]Input.InputType{"command_name": app.defaultCommand}
			objectInput, err := Input.NewObjectInput(params, nil)
			input = objectInput

			if err != nil {
				return 1, err
			}
		} else {
			app.wantsHelp = true
		}
	}

	if name == "" {
		name = app.defaultCommand
		definition := app.GetDefinition()
		definition.SetArguments(nil)
	}

	app.runningCommand = nil
	command, findCommandErr := app.Find(name)

	if findCommandErr != nil {
		// TODO: suggest alternatives

		return 1, findCommandErr
	}

	app.runningCommand = command
	exitCode, runCommandErr := command.Run(input, output)
	app.runningCommand = nil

	return exitCode, runCommandErr
}

func (app *Application) SetDefinition(definition *Input.InputDefinition) {
	app.definition = definition
}

func (app *Application) GetDefinition() *Input.InputDefinition {
	if app.definition == nil {
		app.definition = app.getDefaultInputDefinition()
	}

	if app.singleCommand {
		inputDefinition := app.definition
		inputDefinition.SetArguments(nil)

		return inputDefinition
	}

	return app.definition
}

func (app *Application) GetHelp() string {
	return app.getLongVersion()
}

func (app *Application) AreErrorsCaught() bool {
	return app.catchErrors
}

func (app *Application) SetCatchErrors(boolean bool) {
	app.catchErrors = boolean
}

func (app *Application) IsAutoExitEnabled() bool {
	return app.autoExit
}

func (app *Application) SetAutoExit(boolean bool) {
	app.autoExit = boolean
}

func (app *Application) GetName() string {
	return app.name
}

func (app *Application) SetName(name string) {
	app.name = name
}

func (app *Application) GetVersion() string {
	return app.version
}

func (app *Application) SetVersion(version string) {
	app.version = version
}

func (app *Application) getLongVersion() string {
	if app.GetName() == "" || app.GetName() == "UNKNOWN" {
		if app.GetVersion() == "" || app.GetVersion() == "UNKNOWN" {
			return fmt.Sprintf("%s <highlight>%s</highlight>", app.GetName(), app.GetVersion())
		}

		return app.GetName()
	}

	return "Console Tool"
}

func (app *Application) AddCommands(commands []*Command.Command) {
	for _, c := range commands {
		app.Add(c)
	}
}

func (app *Application) Add(command *Command.Command) *Command.Command {
	app.init()

	command.SetApplicationDefinition(app.GetDefinition())

	if !command.IsEnabled() {
		command.SetApplicationDefinition(nil)
		return nil
	}

	if command.GetName() == "" {
		panic("Commands must have a name.")
	}

	app.commands[command.GetName()] = command

	for _, alias := range command.GetAliases() {
		app.commands[alias] = command
	}

	return command
}

func (app *Application) Get(name string) (*Command.Command, error) {
	app.init()

	if !app.Has(name) {
		return nil, Error.NewCommandNotFoundError(fmt.Sprintf("The command \"%s\" does not exist.", name), nil)
	}

	if app.commands[name] == nil {
		return nil, Error.NewCommandNotFoundError(fmt.Sprintf("The \"%s\" commandcannot be found because it is registered under multiple names. Make sure you don't set a different name.", name), nil)
	}

	command := app.commands[name]

	if app.wantsHelp {
		app.wantsHelp = false

		helpCommand, _ := app.Get("help")
		helpCommand.SetMeta(map[string]any{
			"command": command,
		})

		return helpCommand, nil
	}

	return command, nil
}

func (app *Application) Has(name string) bool {
	app.init()

	return app.commands[name] != nil
}

func (app *Application) GetNamespaces() []string {
	namespacesMap := make(map[string]int)

	for _, command := range app.All("") {
		if command.IsHidden() || command.GetName() == "" {
			continue
		}

		for _, namespace := range app.extractAllNamespace(command.GetName()) {
			namespacesMap[namespace] = 0
		}

		for _, alias := range command.GetAliases() {
			namespacesMap[app.ExtractNamespace(alias, -1)] = 0
		}
	}

	namespaces := make([]string, len(namespacesMap))
	for ns := range namespacesMap {
		namespaces = append(namespaces, ns)
	}

	return namespaces
}

func (app *Application) FindNamespaces(namespace string) string {
	panic("TODO: Application.FindNamespaces()")
}

func (app *Application) Find(name string) (*Command.Command, error) {
	app.init()

	for _, command := range app.commands {
		for _, alias := range command.GetAliases() {
			if !app.Has(alias) {
				app.commands[alias] = command
			}
		}
	}

	if app.Has(name) {
		return app.Get(name)
	}

	panic("TODO: alternatives support in app.Find()")
}

func (app *Application) All(namespace string) map[string]*Command.Command {
	app.init()

	if namespace == "" {
		return app.commands
	}

	re := regexp.MustCompile(`\:`)
	commands := make(map[string]*Command.Command)

	for name, command := range app.commands {
		limit := len(re.FindAllStringSubmatchIndex(name, -1)) + 1

		if namespace == app.ExtractNamespace(name, limit) {
			commands[name] = command
		}
	}

	return commands
}

func (app *Application) GetAbbreviations(names []string) map[string][]string {
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

func (app *Application) RenderError(output Output.OutputInterface, err error) {
	output.Writeln("", Output.VERBOSITY_QUIET)

	app.doRenderError(output, err)

	if app.runningCommand != nil {
		output.Writeln(
			fmt.Sprintf("<highlight>%s %s</highlight>", app.GetName(), app.runningCommand.GetSynopsis(false)),
			Output.VERBOSITY_QUIET,
		)
		output.Writeln("", Output.VERBOSITY_QUIET)
	}
}

func (app *Application) doRenderError(output Output.OutputInterface, err error) {
	message := strings.TrimSpace(err.Error())
	length := 0
	width, _ := Terminal.GetWidth()
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
		formattedLine := fmt.Sprintf("<error>  %s  %s</error>", Formatter.Escape(line), strings.Repeat(" ", length-linesLength[i]))
		messages = append(messages, formattedLine)
	}

	messages = append(messages, emptyLine, "")

	output.Writelns(messages, Output.VERBOSITY_QUIET)
}

func (app *Application) configureIO(input Input.InputInterface, output Output.OutputInterface) {
	if input.HasParameterOption("--ansi", true) {
		output.SetDecorated(true)
	} else if input.HasParameterOption("--no-ansi", true) {
		output.SetDecorated(false)
	}

	if input.HasParameterOption("--no-interaction", true) || input.HasParameterOption("-n", true) {
		input.SetInteractive(false)
	}

	shellVerbosity, err := strconv.Atoi(os.Getenv("SHELL_VERBOSITY"))
	if err != nil {
		shellVerbosity = 0
	}

	switch shellVerbosity {
	case -1:
		output.SetVerbosity(Output.VERBOSITY_QUIET)
	case 1:
		output.SetVerbosity(Output.VERBOSITY_VERBOSE)
	case 2:
		output.SetVerbosity(Output.VERBOSITY_VERY_VERBOSE)
	case 3:
		output.SetVerbosity(Output.VERBOSITY_DEBUG)
	default:
		shellVerbosity = 0
	}

	if input.HasParameterOption("--quiet", true) || input.HasParameterOption("-q", true) {
		output.SetVerbosity(Output.VERBOSITY_QUIET)
		shellVerbosity = -1
	} else {
		if input.HasParameterOption("-vvv", true) ||
			input.HasParameterOption("--verbose=3", true) ||
			input.GetParameterOption("--verbose", false, true) == "3" {
			output.SetVerbosity(Output.VERBOSITY_DEBUG)
			shellVerbosity = 3
		} else if input.HasParameterOption("-vv", true) ||
			input.HasParameterOption("--verbose=2", true) ||
			input.GetParameterOption("--verbose", false, true) == "2" {
			output.SetVerbosity(Output.VERBOSITY_VERBOSE)
			shellVerbosity = 2
		} else if input.HasParameterOption("-v", true) ||
			input.HasParameterOption("--verbose=1", true) ||
			input.HasParameterOption("--verbose", true) {
			output.SetVerbosity(Output.VERBOSITY_VERBOSE)
			shellVerbosity = 1
		}

		if shellVerbosity == -1 {
			input.SetInteractive(false)
		}

		os.Setenv("SHELL_VERBOSITY", fmt.Sprint(shellVerbosity))
	}
}

func (app *Application) getCommandName(input Input.InputInterface) string {
	if app.singleCommand {
		return app.defaultCommand
	}

	first := input.GetFirstArgument()
	return first.(string)
}

func (app *Application) getDefaultInputDefinition() *Input.InputDefinition {
	commandArgument := Input.NewInputArgument("command", Input.INPUT_ARGUMENT_REQUIRED, "The command to execute", "", nil)
	arguments := []*Input.InputArgument{commandArgument}

	helpOption := Input.NewInputOption("--help", "-h", Input.INPUT_OPTION_BOOLEAN, fmt.Sprintf("Display help for the given command, or the <highlight>%s</highlight> command (if no command is given)", app.defaultCommand), nil, nil)
	quietOption := Input.NewInputOption("--quiet", "-q", Input.INPUT_OPTION_BOOLEAN, "Do not output any message", nil, nil)
	verboseoption := Input.NewInputOption("--verbose", "-v|vv|vvv", Input.INPUT_OPTION_BOOLEAN, "Increase the verbosity of messages: normal (1), verbose (2) or debug (3)", nil, nil)
	versionOption := Input.NewInputOption("--version", "-V", Input.INPUT_OPTION_BOOLEAN, "Display this application version", nil, nil)
	ansiOption := Input.NewInputOption("--ansi", "", Input.INPUT_OPTION_NEGATABLE, "Force (or disable --no-ansi) ANSI output", nil, nil)
	noInteractionOption := Input.NewInputOption("--no-interaction", "-n", Input.INPUT_OPTION_BOOLEAN, "Do not ask any interactive question", nil, nil)

	options := []*Input.InputOption{
		helpOption,
		quietOption,
		verboseoption,
		versionOption,
		ansiOption,
		noInteractionOption,
	}

	definition := Input.NewInputDefinition(arguments, options)

	return definition
}

func (app *Application) getDefaultCommands() []*Command.Command {
	return []*Command.Command{
		app.newHelpCommand(),
		app.newListCommand(),
	}
}

func (app *Application) newHelpCommand() *Command.Command {
	command := Command.NewCommand("help", func(self *Command.Command) (int, error) {
		var target *Command.Command = nil

		meta := self.Meta()
		if metaMap, ok := meta.(map[string]*Command.Command); ok {
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

		// TODO: describe command
		self.Output().Writeln(fmt.Sprintf("TODO: Describe command \"%s\"", target.GetName()), 0)

		descriptor := Descriptor.NewTextDescriptor(self.Output())
		raw, _ := self.BoolOption("raw")
		descriptor.DescribeCommand(target, Descriptor.NewDescriptorOptions("", raw, false, 0))

		return 0, nil
	})

	command.IgnoreValidationErrors()
	command.SetDefinition(Input.NewInputDefinition(
		[]*Input.InputArgument{
			Input.NewInputArgument("command_name", Input.INPUT_ARGUMENT_OPTIONAL, "The command name", "help", nil),
		},
		[]*Input.InputOption{
			Input.NewInputOption("raw", "", Input.INPUT_OPTION_BOOLEAN, "To output raw command help", nil, nil),
		},
	))
	command.SetDescription("Display help for a command")
	command.SetHelp(`
The <highlight>%command.name%</highlight> command displays help for a given command:

  <highlight>%command.full_name% list</highlight>

You can also output the help in other formats by using the <highlight>--format</highlight> option:

  <highlight>%command.full_name% --format=json list</highlight>

To display the list of available commands, please use the <highlight>list</highlight> command.
`)

	return command
}

func (app *Application) newListCommand() *Command.Command {
	command := Command.NewCommand("list", func(self *Command.Command) (int, error) {
		descriptor := Descriptor.NewTextDescriptor(self.Output())
		raw, _ := self.BoolOption("raw")
		short, _ := self.BoolOption("short")
		namespace, _ := self.StringArgument("namespace")
		descriptor.DescribeApplication(app, Descriptor.NewDescriptorOptions(namespace, raw, short, 0))

		return 0, nil
	})

	command.IgnoreValidationErrors()
	command.SetDefinition(Input.NewInputDefinition(
		[]*Input.InputArgument{
			Input.NewInputArgument("namespace", Input.INPUT_ARGUMENT_OPTIONAL, "The namespace name", nil, nil),
		},
		[]*Input.InputOption{
			Input.NewInputOption("raw", "", Input.INPUT_OPTION_BOOLEAN, "To output raw command list", nil, nil),
			Input.NewInputOption("short", "", Input.INPUT_OPTION_BOOLEAN, "To skip describing commands' arguments", nil, nil),
		},
	))
	command.SetDescription("List commands")
	command.SetHelp(`
The <highlight>%command.name%</highlight> command lists all commands:

  <highlight>%command.full_name%</highlight>

You can also display the commands for a specific namespace:

  <highlight>%command.full_name% test</highlight>

You can also output the information in other formats by using the <highlight>--format</highlight> option:

  <highlight>%command.full_name% --format=json</highlight>

It's also possible to get raw list of commands (useful for embedding command runner):

  <highlight>%command.full_name% --raw</highlight>
`)

	return command
}

func (app *Application) getAbbreviationSuggestions(abbrevs []string) string {
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
	panic("TODO: app.findAlternatives()")
}

func (app *Application) SetDefaultCommand(commandName string, isSingleCommand bool) error {
	for strings.HasPrefix(commandName, "|") {
		commandName = commandName[1:]
	}

	app.defaultCommand = strings.Split(commandName, "|")[0]

	if app.singleCommand {
		_, e := app.Find(commandName)
		app.singleCommand = true

		if e != nil {
			return e
		}
	}

	return nil
}

func (app *Application) IsSingleCommand() bool {
	return app.singleCommand
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

	for _, command := range app.getDefaultCommands() {
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

	matrix := make([][]int, 0)

	for i := 0; i <= aLen; i++ {
		matrix[i] = []int{i}
	}

	for j := 0; j <= bLen; j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= aLen; i++ {
		for j := 1; j <= bLen; j++ {
			var cost int
			if a[i-1] != b[j-1] {
				cost = 1
			}

			matrix[i][j] = min(matrix[i-1][j]+1, matrix[i][j-1]+1, matrix[i-1][j-1]+cost)
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