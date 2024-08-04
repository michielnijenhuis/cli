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
	stream      *os.File
	options     map[string]InputType
	arguments   map[string]InputType
	interactive bool
	Tokens      []string
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
		Tokens:      args,
		parsed:      make([]string, 0),
		definition:  &InputDefinition{},
		stream:      os.Stdin,
		options:     make(map[string]InputType),
		arguments:   make(map[string]InputType),
		interactive: TerminalIsInteractive(),
	}

	return i
}

func (i *Input) SetDefinition(definition *InputDefinition) error {
	if definition == nil {
		i.definition = &InputDefinition{}
		return nil
	} else {
		i.Bind(definition)
		err := i.Parse()
		if err != nil {
			return err
		}

		return i.Validate()
	}
}

func (i *Input) Bind(definition *InputDefinition) {
	i.arguments = make(map[string]InputType)
	i.options = make(map[string]InputType)
	i.definition = definition
}

func (i *Input) Parse() error {
	parseOptions := true
	i.parsed = make([]string, 0, len(i.Tokens))
	i.parsed = append(i.parsed, i.Tokens...)
	var token string
	var err error

	for {
		if len(i.parsed) == 0 {
			break
		}

		token = helper.Shift(&i.parsed)
		parseOptions, err = i.parseToken(token, parseOptions)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Input) Validate() error {
	definition := i.definition
	if definition == nil {
		return errors.New("inputDefinition not found")
	}

	givenArguments := i.arguments
	if givenArguments == nil {
		return errors.New("given arguments not found")
	}

	arguments := definition.Arguments
	if arguments != nil {
		missingArguments := make([]string, 0, len(arguments))
		for _, arg := range arguments {
			name := arg.Name
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

	validationError := i.runOptionValidators()
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

func (i *Input) Arguments() map[string]InputType {
	definition := i.definition
	args := make(map[string]InputType)

	if definition != nil {
		for k, v := range definition.ArgumentDefaults() {
			args[k] = v
		}
	}

	for k, v := range i.arguments {
		args[k] = v
	}

	return args
}

func (i *Input) StringArgument(name string) (string, error) {
	definition := i.definition
	if definition == nil {
		return "", errors.New("no input definition found")
	}

	if !definition.HasArgument(name) {
		return "", fmt.Errorf("the \"%s\" argument does not exist", name)
	}

	if i.arguments[name] == nil {
		arg, e := definition.Argument(name)
		if e != nil {
			return "", e
		}

		value := arg.DefaultValue
		str, isStr := value.(string)
		if !isStr {
			return "", nil
		}

		return str, nil
	}

	value := i.arguments[name]
	str, isStr := value.(string)
	if !isStr {
		return "", nil
	}

	return str, nil
}

func (i *Input) ArrayArgument(name string) ([]string, error) {
	definition := i.definition
	if definition == nil {
		return []string{}, errors.New("no input definition found")
	}

	if !definition.HasArgument(name) {
		return []string{}, fmt.Errorf("the \"%s\" argument does not exist", name)
	}

	if i.arguments[name] == nil {
		arg, e := definition.Argument(name)
		if e != nil {
			return []string{}, e
		}

		value := arg.DefaultValue
		arr, isArr := value.([]string)
		if !isArr {
			return []string{}, nil
		}

		return arr, nil
	}

	value := i.arguments[name]
	arr, isArr := value.([]string)
	if !isArr {
		return []string{}, nil
	}

	return arr, nil
}

func (i *Input) SetArgument(name string, value InputType) error {
	definition := i.definition
	if definition == nil {
		return errors.New("no input definition found")
	}

	if !definition.HasArgument(name) {
		return fmt.Errorf("the \"%s\" argument does not exist", name)
	}

	i.arguments[name] = value
	return nil
}

func (i *Input) HasArgument(name string) bool {
	definition := i.definition
	if definition == nil {
		return false
	}

	return definition.HasArgument(name)
}

func (i *Input) Options() map[string]InputType {
	definition := i.definition
	opts := make(map[string]InputType)

	if definition != nil {
		for k, v := range definition.OptionDefaults() {
			opts[k] = v
		}
	}

	for k, v := range i.options {
		opts[k] = v
	}

	return opts
}

func (i *Input) BoolOption(name string) (bool, error) {
	definition := i.definition
	if definition == nil {
		return false, errors.New("no input definition found")
	}

	if !definition.HasOption(name) {
		return false, fmt.Errorf("the \"%s\" option does not exist", name)
	}

	if i.options[name] == nil {
		opt, e := definition.Option(name)
		if e != nil {
			return false, e
		}

		value := opt.DefaultValue
		boolval, isBool := value.(bool)
		if !isBool {
			return false, nil
		}

		if definition.HasNegation(name) {
			return !boolval, nil
		}

		return boolval, nil
	}

	value := i.options[name]
	boolval, isBool := value.(bool)
	if !isBool {
		return false, nil
	}

	if definition.HasNegation(name) {
		return !boolval, nil
	}

	return boolval, nil
}

func (i *Input) StringOption(name string) (string, error) {
	definition := i.definition
	if definition == nil {
		return "", errors.New("no input definition found")
	}

	if !definition.HasOption(name) {
		return "", fmt.Errorf("the \"%s\" option does not exist", name)
	}

	if i.options[name] == nil {
		opt, e := definition.Option(name)
		if e != nil {
			return "", e
		}

		value := opt.DefaultValue
		str, isStr := value.(string)
		if !isStr {
			return "", nil
		}

		return str, nil
	}

	value := i.options[name]
	str, isStr := value.(string)
	if !isStr {
		return "", nil
	}

	return str, nil
}

func (i *Input) ArrayOption(name string) ([]string, error) {
	definition := i.definition
	if definition == nil {
		return []string{}, errors.New("no input definition found")
	}

	if !definition.HasOption(name) {
		return []string{}, fmt.Errorf("the \"%s\" option does not exist", name)
	}

	if i.options[name] == nil {
		opt, e := definition.Option(name)
		if e != nil {
			return []string{}, e
		}

		value := opt.DefaultValue
		arr, isArr := value.([]string)
		if !isArr {
			return []string{}, nil
		}

		return arr, nil
	}

	value := i.options[name]
	arr, isArr := value.([]string)
	if !isArr {
		return []string{}, nil
	}

	return arr, nil
}

func (i *Input) SetOption(name string, value InputType) error {
	definition := i.definition
	if definition == nil {
		return errors.New("no input definition found")
	}

	if definition.HasNegation(name) {
		i.options[definition.NegationToName(name)] = value
		return nil
	}

	if !definition.HasOption(name) {
		return fmt.Errorf("the \"%s\" option does not exist", name)
	}

	i.options[name] = value
	return nil
}

func (i *Input) HasOption(name string) bool {
	definition := i.definition
	if definition == nil {
		return false
	}

	return definition.HasOption(name) || definition.HasNegation(name)
}

func (i *Input) SetStream(stream *os.File) {
	i.stream = stream
}

func (i *Input) Stream() *os.File {
	return i.stream
}

func (i *Input) runArgumentValidators() error {
	definition := i.definition
	if definition == nil {
		return errors.New("no input definition found")
	}

	args := definition.Arguments
	for _, arg := range args {
		value := i.arguments[arg.Name]
		if value == arg.DefaultValue {
			continue
		}

		if !arg.IsRequired() && value == nil {
			continue
		}

		validator := arg.Validator
		if validator == nil {
			continue
		}

		validationError := validator(value)
		if validationError != nil {
			return validationError
		}
	}

	return nil
}

func (i *Input) runOptionValidators() error {
	definition := i.definition
	if definition == nil {
		return errors.New("no input definition found")
	}

	opts := definition.Options
	for _, opt := range opts {
		value := i.options[opt.Name]
		if value == opt.DefaultValue {
			continue
		}

		if !opt.IsValueRequired() && value == nil {
			continue
		}

		validator := opt.Validator
		if validator == nil {
			continue
		}

		validationError := validator(value)
		if validationError != nil {
			return validationError
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

func (i *Input) parseToken(token string, parseOptions bool) (bool, error) {
	if parseOptions && token == "" {
		err := i.parseArgument(token)

		if err != nil {
			return false, err
		}
	} else if parseOptions && token == "" {
		return false, nil
	} else if parseOptions && strings.HasPrefix(token, "--") {
		err := i.parseLongOption(token)
		if err != nil {
			return false, err
		}
	} else if parseOptions && strings.HasPrefix(token, "-") && token != "-" {
		err := i.parseShortOption(token)
		if err != nil {
			return false, err
		}
	} else {
		err := i.parseArgument(token)
		if err != nil {
			return false, err
		}
	}

	return parseOptions, nil
}

func (i *Input) parseArgument(token string) error {
	definition := i.definition
	currentCount := uint(len(i.arguments))
	argsCount := uint(len(definition.Arguments))

	if currentCount < argsCount {
		// if input is expecting another argument, add it
		arg, err := definition.ArgumentByIndex(currentCount)
		if err != nil {
			return err
		}

		if arg.IsArray() {
			i.arguments[arg.Name] = []string{token}
		} else {
			i.arguments[arg.Name] = token
		}

		return nil
	} else {
		if currentCount == argsCount {
			arg, err := definition.ArgumentByIndex(currentCount - 1)
			if err != nil {
				return err
			}

			// if last argument isArray(), append token to last argument
			if arg.IsArray() {
				name := arg.Name
				current := i.arguments[name]
				arr, isArr := current.([]string)
				if isArr {
					i.arguments[name] = append(arr, token)
				} else {
					i.arguments[name] = []string{token}
				}

				return nil
			}
		}

		all := definition.Arguments

		var commandName string
		inputArgument := all[0]

		if inputArgument != nil && inputArgument.Name == "command" {
			commandValue := i.arguments["command"]
			str, isStr := commandValue.(string)

			if isStr {
				commandName = str
			}

			delete(i.arguments, "command")
		}

		var message string
		if len(all) > 0 {
			allCommands := make([]string, 0, len(all))
			for _, arg := range all {
				allCommands = append(allCommands, arg.Name)
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

func (i *Input) parseLongOption(token string) error {
	name := token[2:]
	pos := strings.Index(name, "=")

	if pos != -1 {
		value := name[pos+1:]
		if value == "" {
			helper.Unshift(&i.parsed, value)
		}
		return i.addLongOption(name[0:pos], value)
	} else {
		return i.addLongOption(name, nil)
	}
}

func (i *Input) parseShortOption(token string) error {
	name := token[1:]

	if len(name) > 1 {
		short := name[0:1]
		if i.definition.HasShortcut(short) {
			opt, err := i.definition.OptionForShortcut(short)
			if err != nil {
				return err
			}

			if opt.AcceptValue() {
				// an option with a value (with no space)
				return i.addShortOption(short, name[1:])
			}
		}

		return i.parseShortOptionSet(name)
	} else {
		return i.addShortOption(name, nil)
	}
}

func (i *Input) parseShortOptionSet(name string) error {
	length := len(name)
	for idx := 0; idx < length; idx++ {
		char := name[idx : idx+1]
		if !i.definition.HasShortcut(char) {
			return fmt.Errorf("the \"-%s\" option does not exist", char)
		}

		opt, err := i.definition.OptionForShortcut(char)
		if err != nil {
			return err
		}

		if opt.AcceptValue() {
			if length-1 == idx {
				err := i.addLongOption(opt.Name, nil)
				if err != nil {
					return err
				}
			} else {
				err := i.addLongOption(opt.Name, name[idx+1:])
				if err != nil {
					return err
				}
			}

			break
		} else {
			err := i.addLongOption(opt.Name, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *Input) addShortOption(shortcut string, value InputType) error {
	if !i.definition.HasShortcut(shortcut) {
		return fmt.Errorf("the \"-%s\" option does not exist", shortcut)
	}

	opt, err := i.definition.OptionForShortcut(shortcut)
	if err != nil {
		return err
	}

	return i.addLongOption(opt.Name, value)
}

func (i *Input) addLongOption(name string, value InputType) error {
	if !i.definition.HasOption(name) {
		if !i.definition.HasNegation(name) {
			return fmt.Errorf("the \"--%s\" option does not exist", name)
		}

		optName := i.definition.NegationToName(name)
		i.options[optName] = false

		if value != nil && value != "" {
			return fmt.Errorf("the \"--%s\" option does not accept a value", name)
		}

		return nil
	}

	opt, e := i.definition.Option(name)
	if e != nil {
		return e
	}

	if value != nil && value != "" && !opt.AcceptValue() {
		return fmt.Errorf("the \"--%s\" option does not accept a value", name)
	}

	if (value == nil || value == "") && opt.AcceptValue() && len(i.parsed) > 0 {
		// if option accepts an optional or mandatory argument
		// let's see if there is one provided
		next := helper.Shift(&i.parsed)
		if (next != "" && next[0:1] != "-") || next == "" {
			value = next
		} else {
			helper.Unshift(&i.parsed, next)
		}
	}

	if value == nil || value == "" {
		if opt.IsValueOptional() {
			return fmt.Errorf("the \"--%s\" option requires a value", name)
		}

		if !opt.IsArray() && !opt.IsValueOptional() {
			value = true
		}
	}

	if opt.IsArray() {
		cur := i.options[name]
		arr, isArr := cur.([]string)
		if !isArr {

		} else {
			arr = append(arr, value.(string))
		}
		i.options[name] = arr
	} else {
		i.options[name] = value
	}

	return nil
}

func (i *Input) FirstArgument() InputType {
	isOption := false
	tokenCount := len(i.Tokens)

	for idx, token := range i.Tokens {
		if token != "" && strings.HasPrefix(token, "-") {
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

			value := i.options[name]
			if (value == nil || value == "") && !i.definition.HasShortcut(name) {
				// noop
			} else if value != "" && value != nil && i.Tokens[idx+1] == value {
				isOption = true
			} else {
				name = i.definition.ShortcutToName(name)
				value = i.options[name]

				if value != "" && value != nil && i.Tokens[idx+1] == value {
					isOption = true
				}
			}

			continue
		}

		if isOption {
			isOption = false
			continue
		}

		return token
	}

	return ""
}

func (i *Input) HasParameterOption(value string, onlyParams bool) bool {
	for _, token := range i.Tokens {
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

func (i *Input) ParameterOption(value string, defaultValue InputType, onlyParams bool) InputType {
	tokens := make([]string, 0, len(i.Tokens))
	copy(tokens, i.Tokens)

	for len(tokens) > 0 {
		token := helper.Shift(&tokens)
		if onlyParams && token == "--" {
			return defaultValue
		}

		if token == value {
			return helper.Shift(&tokens)
		}

		// Options with values:
		//   For long options, test for '--option=' at beginning
		//   For short options, test for '-o' at beginning
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

func (i *Input) String() string {
	re := regexp.MustCompile(`{^(-[^=]+=)(.+)}`)
	tokens := make([]string, 0, len(i.Tokens))

	for _, token := range i.Tokens {
		match := re.FindStringSubmatch(token)

		if match != nil {
			tokens = append(tokens, match[1]+i.EscapeToken(match[2]))
		} else {
			if token != "" && token[0] != '-' {
				tokens = append(tokens, i.EscapeToken(token))
			} else {
				tokens = append(tokens, token)
			}
		}
	}

	return strings.Join(tokens, " ")
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
