package cli

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/michielnijenhuis/cli/helper"
)

type InputParser func(self any) error

type Input struct {
	definition  *InputDefinition
	Stream      *os.File
	flags       map[string]Flag
	arguments   map[string]Arg
	interactive bool
	Args        []string
	parsed      []string
}

func NewInput(args ...string) *Input {
	if len(args) == 1 {
		// accept single string that includes all args
		args = StringToInputArgs(args[0])
	} else if args == nil {
		args = os.Args[1:]
	}

	i := &Input{
		Args:        args,
		parsed:      make([]string, 0),
		definition:  &InputDefinition{},
		Stream:      os.Stdin,
		flags:       make(map[string]Flag),
		arguments:   make(map[string]Arg),
		interactive: TerminalIsInteractive(),
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
	i.flags = make(map[string]Flag)
	i.definition = definition
	return i.parse()
}

func (i *Input) parse() error {
	parseFlags := true
	i.parsed = make([]string, 0, len(i.Args))
	i.parsed = append(i.parsed, i.Args...)
	var token string
	var err error

	for {
		if len(i.parsed) == 0 {
			break
		}

		token = helper.Shift(&i.parsed)
		parseFlags, err = i.parseToken(token, parseFlags)
		if err != nil {
			return err
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
			return fmt.Errorf("not enough arguments (missing: \"%s\")", strings.Join(missingArguments, ", "))
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

func (i *Input) parseToken(token string, parseFlags bool) (bool, error) {
	if parseFlags && token == "" {
		err := i.parseArgument(token)

		if err != nil {
			return false, err
		}
	} else if parseFlags && token == "" {
		return false, nil
	} else if parseFlags && strings.HasPrefix(token, "--") {
		err := i.parseLongFlag(token)
		if err != nil {
			return false, err
		}
	} else if parseFlags && strings.HasPrefix(token, "-") && token != "-" {
		err := i.parseShortFlag(token)
		if err != nil {
			return false, err
		}
	} else {
		err := i.parseArgument(token)
		if err != nil {
			return false, err
		}
	}

	return parseFlags, nil
}

func (i *Input) parseArgument(token string) error {
	definition := i.definition
	currentCount := uint(len(i.arguments))
	argsCount := uint(len(definition.arguments))

	if currentCount < argsCount {
		// if input is expecting another argument, add it
		arg, err := definition.ArgumentByIndex(currentCount)
		if err != nil {
			return err
		}
		i.arguments[arg.GetName()] = arg

		ArgSetValue(arg, token)
		return nil
	} else {
		if currentCount == argsCount {
			arg, err := definition.ArgumentByIndex(currentCount - 1)
			if err != nil {
				return err
			}
			i.arguments[arg.GetName()] = arg

			// if last argument isArray(), append token to last argument
			if a, ok := arg.(*ArrayArg); ok {
				ArgSetValue(a, token)
				return nil
			}
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

		return errors.New(message)
	}
}

func (i *Input) parseLongFlag(token string) error {
	name := token[2:]
	pos := strings.Index(name, "=")

	if pos != -1 {
		value := name[pos+1:]
		if value == "" {
			helper.Unshift(&i.parsed, value)
		}
		return i.addLongFlag(name[0:pos], value)
	} else {
		return i.addLongFlag(name, "")
	}
}

func (i *Input) parseShortFlag(token string) error {
	name := token[1:]

	if len(name) > 1 {
		short := name[0:1]
		if i.definition.HasShortcut(short) {
			flag, err := i.definition.FlagForShortcut(short)
			if err != nil {
				return err
			}

			if FlagAcceptsValue(flag) {
				// a flag with a value (with no space)
				return i.addShortFlag(short, name[1:])
			}
		}

		return i.parseShortFlagSet(name)
	} else {
		return i.addShortFlag(name, "")
	}
}

func (i *Input) parseShortFlagSet(name string) error {
	length := len(name)
	for idx := 0; idx < length; idx++ {
		char := name[idx : idx+1]
		if !i.definition.HasShortcut(char) {
			return fmt.Errorf("the \"-%s\" flag does not exist", char)
		}

		flag, err := i.definition.FlagForShortcut(char)
		if err != nil {
			return err
		}

		if FlagAcceptsValue(flag) {
			if length-1 == idx {
				err := i.addLongFlag(flag.GetName(), "")
				if err != nil {
					return err
				}
			} else {
				err := i.addLongFlag(flag.GetName(), name[idx+1:])
				if err != nil {
					return err
				}
			}

			break
		} else {
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
		return fmt.Errorf("the \"-%s\" flag does not exist", shortcut)
	}

	opt, err := i.definition.FlagForShortcut(shortcut)
	if err != nil {
		return err
	}

	return i.addLongFlag(opt.GetName(), token)
}

func (i *Input) addLongFlag(name string, token string) error {
	boolean := token != ""
	isNegation := false

	if !i.definition.HasFlag(name) {
		if !i.definition.HasNegation(name) {
			return fmt.Errorf("the \"--%s\" flag does not exist", name)
		}

		name = i.definition.NegationToName(name)
		isNegation = true
	}

	flag, e := i.definition.Flag(name)
	if e != nil {
		return e
	}

	if isNegation {
		boolean = false
	}

	i.flags[name] = flag

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

	if token == "" {
		if FlagRequiresValue(flag) {
			return fmt.Errorf("the \"--%s\" flag requires a value", name)
		}

		if !FlagIsArray(flag) && !FlagValueIsOptional(flag) && (!FlagIsNegatable(flag) || !isNegation) {
			boolean = true
		}
	}

	SetFlagValue(flag, token, boolean)

	return nil
}

func (i *Input) FirstArgument() string {
	isOption := false
	tokenCount := len(i.Args)

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
		}

		// Is value for option
		if isOption {
			isOption = false
			continue
		}

		return token
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
