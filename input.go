package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
	"github.com/michielnijenhuis/cli/terminal"
)

type InputParser func(self any) error

type Input struct {
	definition      *InputDefinition
	Stream          *os.File
	flags           map[string]Flag
	arguments       map[string]Arg
	interactive     bool
	givenArguments  []string
	Args            []string
	tokens          []string
	parsed          []string
	initialSttyMode string
	Strict          bool
}

type ErrMissingArguments interface {
	error
	MissingArguments() []string
}

type errMissingArguments struct {
	err     string
	missing []string
}

func (e *errMissingArguments) Error() string {
	return e.err
}

func (e *errMissingArguments) MissingArguments() []string {
	return e.missing
}

func NewInput(args ...string) *Input {
	if len(args) == 1 {
		// accept single string that includes all args
		args = StringToInputArgs(args[0])
	} else if args == nil {
		args = os.Args[1:]
	}

	tokens := make([]string, len(args))
	copy(tokens, args)

	i := &Input{
		Args:           args,
		tokens:         tokens,
		parsed:         make([]string, 0),
		definition:     &InputDefinition{},
		Stream:         os.Stdin,
		flags:          make(map[string]Flag),
		givenArguments: make([]string, 0),
		arguments:      make(map[string]Arg),
		interactive:    terminal.IsInteractive(),
	}

	return i
}

func (i *Input) SetDefinition(definition *InputDefinition) error {
	if definition == nil {
		i.definition = &InputDefinition{}
		return nil
	} else {
		err := i.Bind(definition)
		if err != nil {
			return err
		}

		return i.Validate()
	}
}

func (i *Input) Bind(definition *InputDefinition) error {
	i.arguments = make(map[string]Arg)
	i.givenArguments = make([]string, 0)
	i.flags = make(map[string]Flag)
	i.definition = definition
	return i.parse(i.tokens, nil)
}

func (i *Input) parse(tokens []string, inspector *InputInspector) error {
	parseFlags := true
	i.parsed = make([]string, 0, len(tokens))
	i.parsed = append(i.parsed, tokens...)
	var (
		token       string
		err         error
		keepParsing bool
	)

	for {
		if len(i.parsed) == 0 {
			break
		}

		token = helper.Shift(&i.parsed)
		parseFlags, err, keepParsing = i.parseToken(inspector, token, parseFlags)
		if err != nil {
			return err
		} else if !keepParsing {
			break
		}
	}

	return nil
}

func (i *Input) Validate() error {
	definition := i.definition
	givenArguments := i.arguments

	arguments := definition.arguments
	if arguments != nil {
		missingArguments := make([]string, 0, len(arguments))
		for _, arg := range arguments {
			name := arg.GetName()
			if givenArguments[name] != nil {
				continue
			}

			argument, _ := definition.Argument(name)
			if argument != nil && argument.IsRequired() {
				missingArguments = append(missingArguments, name)
			}
		}

		if len(missingArguments) > 0 {
			return &errMissingArguments{
				err:     fmt.Sprintf("not enough arguments (missing: \"%s\")", strings.Join(missingArguments, ", ")),
				missing: missingArguments,
			}
		}
	}

	validationError := i.runFlagValidators()
	if validationError != nil {
		return validationError
	}

	return i.runArgumentValidators()
}

func (i *Input) IsInteractive() bool {
	return i.interactive
}

func (i *Input) SetInteractive(interactive bool) {
	i.interactive = interactive
}

func (i *Input) Arguments() map[string]Arg {
	args := make(map[string]Arg)
	for k, v := range i.arguments {
		args[k] = v
	}

	return args
}

func (i *Input) String(name string) (string, error) {
	arg, _ := i.definition.Argument(name)
	if arg != nil {
		return GetArgStringValue(arg), nil
	}

	flag, _ := i.definition.Flag(name)
	if flag != nil {
		return GetFlagStringValue(flag), nil
	}

	return "", fmt.Errorf("the \"%s\" argument or flag does not exist", name)
}

func (i *Input) Bool(name string) (bool, error) {
	flag, err := i.definition.Flag(name)
	if err != nil {
		return false, err
	}
	return GetFlagBoolValue(flag), nil
}

func (i *Input) Array(name string) ([]string, error) {
	arg, _ := i.definition.Argument(name)
	if arg != nil {
		return GetArgArrayValue(arg), nil
	}

	flag, _ := i.definition.Flag(name)
	if flag != nil {
		return GetFlagArrayValue(flag), nil
	}

	return nil, fmt.Errorf("the \"%s\" argument or flag does not exist", name)
}

func (i *Input) SetArgument(name string, token string) error {
	arg, err := i.definition.Argument(name)

	if err != nil {
		return fmt.Errorf("the \"%s\" argument does not exist", name)
	}

	i.arguments[name] = arg
	ArgSetValue(arg, token)
	return nil
}

func (i *Input) HasArgument(name string) bool {
	definition := i.definition
	if definition == nil {
		return false
	}
	return definition.HasArgument(name)
}

func (i *Input) SetFlag(name string, str string, boolean bool) error {
	definition := i.definition
	flag, err := definition.Flag(name)
	if err != nil {
		return err
	}
	i.flags[name] = flag
	SetFlagValue(flag, str, boolean)
	return nil
}

func (i *Input) HasFlag(name string) bool {
	definition := i.definition
	if definition == nil {
		return false
	}

	return definition.HasFlag(name) || definition.HasNegation(name)
}

func (i *Input) runArgumentValidators() error {
	definition := i.definition

	args := definition.arguments
	for _, arg := range args {
		_, given := i.arguments[arg.GetName()]
		if !given {
			continue
		}

		err := ValidateArg(arg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Input) runFlagValidators() error {
	for _, flag := range i.definition.flags {
		_, given := i.flags[flag.GetName()]
		if !given {
			continue
		}

		err := ValidateFlag(flag)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Input) EscapeToken(token string) string {
	re := regexp.MustCompile(`{^[\w-]+}`)
	if re.MatchString(token) {
		return token
	}
	re2 := regexp.MustCompile(`'`)
	return re2.ReplaceAllString(token, "'\\''")
}

func (i *Input) parseToken(inspector *InputInspector, token string, parseFlags bool) (bool, error, bool) {
	if parseFlags && token == "" {
		err, keepParsing := i.parseArgument(inspector, token)

		if err != nil {
			return false, err, true
		} else if !keepParsing {
			return false, nil, false
		}
	} else if parseFlags && token == "" {
		return false, nil, true
	} else if parseFlags && strings.HasPrefix(token, "--") {
		err := i.parseLongFlag(inspector, token)
		if err != nil {
			return false, err, true
		}
	} else if parseFlags && strings.HasPrefix(token, "-") && token != "-" {
		err := i.parseShortFlag(inspector, token)
		if err != nil {
			return false, err, true
		}
	} else {
		err, keepParsing := i.parseArgument(inspector, token)
		if err != nil {
			return false, err, keepParsing
		} else if !keepParsing {
			return false, nil, false
		}
	}

	return parseFlags, nil, true
}

func (i *Input) parseArgument(inspector *InputInspector, token string) (error, bool) {
	definition := i.definition
	currentCount := uint(len(i.arguments))
	argsCount := uint(len(definition.arguments))

	if inspector != nil {
		inspector.Args = append(inspector.Args, token)
		return nil, true
	}

	if currentCount < argsCount {
		// if input is expecting another argument, add it
		arg, err := definition.ArgumentByIndex(currentCount)
		if err != nil {
			return err, true
		}
		i.arguments[arg.GetName()] = arg

		ArgSetValue(arg, token)
		return nil, true
	} else {
		if currentCount == 0 {
			if !i.Strict {
				return nil, false
			} else {
				return fmt.Errorf("no arguments expected"), false
			}
		}

		if currentCount == argsCount {
			arg, err := definition.ArgumentByIndex(currentCount - 1)
			if err != nil {
				return err, true
			}
			i.arguments[arg.GetName()] = arg

			// if last argument isArray(), append token to last argument
			if a, ok := arg.(*ArrayArg); ok {
				ArgSetValue(a, token)
				return nil, true
			}
		}

		if !i.Strict {
			return nil, false
		}

		all := definition.arguments

		var commandName string
		inputArgument := all[0]

		if inputArgument != nil && inputArgument.GetName() == "command" {
			cmdArg := i.arguments["command"]
			str := GetArgStringValue(cmdArg)

			if str != "" {
				commandName = str
			}

			delete(i.arguments, "command")
		}

		var message string
		if len(all) > 0 {
			allCommands := make([]string, 0, len(all))
			for _, arg := range all {
				allCommands = append(allCommands, arg.GetName())
			}
			allCommandsString := strings.Join(allCommands, " ")
			if commandName != "" {
				message = fmt.Sprintf("too many arguments to \"%s\" command, expected arguments \"%s\"", commandName, allCommandsString)
			} else {
				message = fmt.Sprintf("too many arguments, expected arguments \"%s\"", allCommandsString)
			}
		} else if commandName != "" {
			message = fmt.Sprintf("no arguments expected for \"%s\" command, got \"%s\"", commandName, token)
		} else {
			message = fmt.Sprintf("no arguments expected, got \"%s\"", token)
		}

		return errors.New(message), false
	}
}

func (i *Input) parseLongFlag(inspector *InputInspector, token string) error {
	name := token[2:]
	pos := strings.Index(name, "=")

	if pos != -1 {
		flagName := name[0:pos]
		value := name[pos+1:]

		if value == "" {
			helper.Unshift(&i.parsed, value)
		}

		if inspector != nil {
			inspector.AddFlag(flagName, value)
			return nil
		}

		return i.addLongFlag(flagName, value)
	} else {
		if inspector != nil {
			inspector.AddFlag(name, "")
			return nil
		}

		return i.addLongFlag(name, "")
	}
}

func (i *Input) parseShortFlag(inspector *InputInspector, token string) error {
	name := token[1:]

	if len(name) > 1 {
		short := name[0:1]
		if i.definition.HasShortcut(short) {
			flag, err := i.definition.FlagForShortcut(short)
			if err != nil {
				if !i.Strict {
					return nil
				}

				return err
			}

			if FlagAcceptsValue(flag) {
				if inspector != nil {
					inspector.AddFlag(short, name[1:])
					return nil
				}

				// a flag with a value (with no space)
				return i.addShortFlag(short, name[1:])
			}
		}

		return i.parseShortFlagSet(inspector, name)
	} else {
		if inspector != nil {
			inspector.AddFlag(name, "")
			return nil
		}

		return i.addShortFlag(name, "")
	}
}

func (i *Input) parseShortFlagSet(inspector *InputInspector, name string) error {
	length := len(name)
	for idx := 0; idx < length; idx++ {
		char := name[idx : idx+1]
		if !i.definition.HasShortcut(char) {
			if !i.Strict {
				continue
			}

			return fmt.Errorf("the \"-%s\" flag does not exist", char)
		}

		flag, err := i.definition.FlagForShortcut(char)
		if err != nil {
			if !i.Strict {
				continue
			}

			return err
		}

		if FlagAcceptsValue(flag) {
			if length-1 == idx {
				if inspector != nil {
					inspector.AddFlag(flag.GetName(), "")
					return nil
				}

				err := i.addLongFlag(flag.GetName(), "")
				if err != nil {
					return err
				}
			} else {
				if inspector != nil {
					inspector.AddFlag(flag.GetName(), name[idx+1:])
					return nil
				}

				err := i.addLongFlag(flag.GetName(), name[idx+1:])
				if err != nil {
					return err
				}
			}

			break
		} else {
			if inspector != nil {
				inspector.AddFlag(flag.GetName(), "")
				return nil
			}

			err := i.addLongFlag(flag.GetName(), "")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *Input) addShortFlag(shortcut string, token string) error {
	if !i.definition.HasShortcut(shortcut) {
		if !i.Strict {
			return nil
		}

		return fmt.Errorf("the \"-%s\" flag does not exist", shortcut)
	}

	opt, err := i.definition.FlagForShortcut(shortcut)
	if err != nil {
		return err
	}

	return i.addLongFlag(opt.GetName(), token)
}

func (i *Input) addLongFlag(name string, token string) error {
	boolean := true

	if !i.definition.HasFlag(name) {
		if !i.definition.HasNegation(name) {
			if !i.Strict {
				return nil
			}

			return fmt.Errorf("the \"--%s\" flag does not exist", name)
		}

		name = i.definition.NegationToName(name)
		boolean = false
	}

	flag, e := i.definition.Flag(name)
	if e != nil {
		if !i.Strict {
			return nil
		}

		return e
	}

	if token != "" && !FlagAcceptsValue(flag) {
		return fmt.Errorf("the \"--%s\" flag does not accept a value", name)
	}

	if token == "" && FlagAcceptsValue(flag) && len(i.parsed) > 0 {
		// if flag accepts an flagal or mandatory argument
		// let's see if there is one provided
		next := helper.Shift(&i.parsed)
		if (next != "" && next[0:1] != "-") || next == "" {
			token = next
		} else {
			helper.Unshift(&i.parsed, next)
		}
	}

	if token == "" && FlagRequiresValue(flag) {
		return fmt.Errorf("the \"--%s\" flag requires a value", name)
	}

	i.flags[name] = flag
	SetFlagValue(flag, token, boolean)

	return nil
}

func (i *Input) FirstArgument() string {
	args := i.GivenArguments()
	if len(args) > 0 {
		return args[0]
	}
	return ""
}

func (i *Input) HasParameterFlag(value string, onlyParams bool) bool {
	for _, token := range i.Args {
		if onlyParams && token == "--" {
			return false
		}

		leading := value
		if strings.HasPrefix(value, "--") {
			leading = value + "="
		}

		if token == value || (leading != "" && strings.HasPrefix(token, leading)) {
			return true
		}
	}

	return false
}

func (i *Input) ParameterFlag(value string, defaultValue InputType, onlyParams bool) InputType {
	tokens := make([]string, 0, len(i.Args))
	copy(tokens, i.Args)

	for len(tokens) > 0 {
		token := helper.Shift(&tokens)
		if onlyParams && token == "--" {
			return defaultValue
		}

		if token == value {
			return helper.Shift(&tokens)
		}

		// Flags with values:
		//   For long flags, test for '--flag=' at beginning
		//   For short flags, test for '-o' at beginning
		leading := value
		if strings.HasPrefix(value, "--") {
			leading = value + "="
		}

		if leading != "" && strings.HasPrefix(token, leading) {
			return token[len(leading):]
		}
	}

	return nil
}

func StringToInputArgs(cmd string) []string {
	segments := strings.Split(cmd, " ")
	out := make([]string, 0)
	stack := make([]string, 0)
	var i int
	var collectingBy string

	for i < len(segments) {
		current := segments[i]
		i++

		if collectingBy == "" {
			for _, char := range []string{"'", `"`} {
				if strings.HasPrefix(current, char) && !strings.HasSuffix(current, char) {
					stack = append(stack, current[1:])
					collectingBy = char
					break
				}
			}

			if collectingBy == "" {
				out = append(out, current)
			}

		} else if strings.HasSuffix(current, collectingBy) {
			stack = append(stack, current[:len(current)-1])
			out = append(out, strings.Join(stack, " "))
			stack = make([]string, 0)
			collectingBy = ""
		} else {
			stack = append(stack, current)
		}
	}

	if len(stack) > 0 {
		out = append(out, stack...)
	}

	return out
}

func (i *Input) SetTty(mode string) (string, error) {
	if i.initialSttyMode == "" {
		c := exec.Command("stty", "-g")
		c.Stdin = i.Stream

		out, err := c.Output()
		if err != nil {
			return "", err
		}

		i.initialSttyMode = string(out)
	}

	c := exec.Command("stty", StringToInputArgs(mode)...) // #nosec G204
	c.Stdin = i.Stream

	out, err := c.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (i *Input) RestoreTty() error {
	if i.initialSttyMode != "" {
		c := exec.Command("stty", StringToInputArgs(i.initialSttyMode)...) // #nosec G204

		err := c.Run()
		if err != nil {
			return err
		}

		i.initialSttyMode = ""
	}

	return nil
}

type InputInspector struct {
	Args  []string
	Flags map[string][]string
}

func (i InputInspector) String() string {
	var sb strings.Builder

	sb.WriteString("Input inspection:\n")

	if i.ArgsCount() > 0 {
		sb.WriteString("  Args:\n")
		for i, arg := range i.Args {
			sb.WriteString(fmt.Sprintf("    %d: %s\n", i, arg))
		}
	}

	if i.FlagsCount() > 0 {
		sb.WriteString("  Flags:\n")
		for name, values := range i.Flags {
			if len(values) == 0 {
				sb.WriteString(fmt.Sprintf("    %s\n", name))
			} else {
				sb.WriteString(fmt.Sprintf("    %s: %s\n", name, strings.Join(values, ", ")))
			}
		}
	}

	return sb.String()
}

func (i *InputInspector) AddArg(value string) {
	if i.Args == nil {
		i.Args = make([]string, 0)
	}

	i.Args = append(i.Args, value)
}

func (i *InputInspector) AddFlag(name string, value string) {
	if i.Flags == nil {
		i.Flags = make(map[string][]string)
	}

	if i.Flags[name] == nil {
		i.Flags[name] = make([]string, 0)
	}

	if value != "" {
		i.Flags[name] = append(i.Flags[name], value)
	}
}

func (i InputInspector) ArgsCount() int {
	return len(i.Args)
}

func (i InputInspector) FlagsCount() int {
	return len(i.Flags)
}

func (i InputInspector) FlagIsGiven(flag Flag) bool {
	for name := range i.Flags {
		if name == flag.GetName() {
			return true
		}

		for _, short := range flag.GetShortcuts() {
			if name == short {
				return true
			}
		}
	}

	return false
}

func (i *Input) Inspect(tokens []string) (InputInspector, error) {
	inspector := InputInspector{}
	err := i.parse(tokens, &inspector)

	return inspector, err
}

// TODO: refactor, use Inspect()
func (i *Input) GivenArguments() []string {
	isOption := false
	tokenCount := len(i.Args)
	arguments := make([]string, 0)

	for idx, token := range i.Args {
		// Is option
		if strings.HasPrefix(token, "-") {
			// Has value, or is last token
			if strings.Contains(token, "=") || idx+1 >= tokenCount {
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

			flag, _ := i.definition.Flag(name)
			if flag == nil {
				// Try again with the shortcut
				flag, _ = i.definition.FlagForShortcut(name)

				if flag == nil {
					continue
				}
			}

			// If flag accepts a value, check if the next token is not an option value
			if FlagAcceptsValue(flag) && !strings.HasPrefix(i.Args[idx+1], "-") {
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

	return arguments
}
