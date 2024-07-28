package input

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	err "github.com/michielnijenhuis/cli/error"
	helper "github.com/michielnijenhuis/cli/helper"
)

type ArgvInput struct {
	Input
	tokens []string
	parsed []string
}

func NewArgvInput(argv []string, definition *InputDefinition) (*ArgvInput, error) {
	if argv == nil {
		argv = os.Args[1:]
	}

	baseInput, err := NewInput(definition)

	input := &ArgvInput{
		Input:  *baseInput,
		tokens: argv,
		parsed: make([]string, 0),
	}

	return input, err
}

func (input *ArgvInput) Parse() error {
	parseOptions := true
	copy(input.parsed, input.tokens)
	var token string
	var err error

	for {
		if len(input.parsed) == 0 {
			break
		}

		token = helper.Shift(&input.parsed)
		parseOptions, err = input.parseToken(token, parseOptions)
		if err != nil {
			return err
		}
	}

	return nil
}

func (input *ArgvInput) parseToken(token string, parseOptions bool) (bool, error) {
	if parseOptions && token == "" {
		err := input.parseArgument(token)

		if err != nil {
			return false, err
		}
	} else if parseOptions && token == "" {
		return false, nil
	} else if parseOptions && strings.HasPrefix(token, "--") {
		err := input.parseLongOption(token)
		if err != nil {
			return false, err
		}
	} else if parseOptions && strings.HasPrefix(token, "-") && token != "-" {
		err := input.parseShortOption(token)
		if err != nil {
			return false, err
		}
	} else {
		err := input.parseArgument(token)
		if err != nil {
			return false, err
		}
	}

	return parseOptions, nil
}

func (input *ArgvInput) parseArgument(token string) error {
	definition := input.definition
	currentCount := uint(len(input.arguments))
	argsCount := uint(len(definition.GetArguments()))

	if currentCount < argsCount {
		// if input is expecting another argument, add it
		arg, err := definition.GetArgumentByIndex(currentCount)

		if err != nil {
			return err
		}

		if arg.IsArray() {
			input.arguments[arg.GetName()] = []string{token}
		} else {
			input.arguments[arg.GetName()] = token
		}

		return nil
	} else {
		if currentCount == argsCount {
			arg, err := definition.GetArgumentByIndex(currentCount - 1)

			if err != nil {
				return err
			}

			// if last argument isArray(), append token to last argument
			if arg.IsArray() {
				name := arg.GetName()
				current := input.arguments[name]
				arr, isArr := current.([]string)
				if isArr {
					input.arguments[name] = []string{token}
				} else {
					input.arguments[name] = append(arr, token)
				}

				return nil
			}
		}

		all := definition.GetArguments()

		var key string
		for k := range all {
			key = k
			break
		}

		var commandName string
		inputArgument := all[key]

		if inputArgument != nil && inputArgument.GetName() == "command" {
			commandValue := input.arguments["command"]
			str, isStr := commandValue.(string)

			if !isStr {
				commandValue = ""
			} else {
				commandValue = str
			}

			delete(input.arguments, "command")
		}

		var message string
		if len(all) > 0 {
			allCommands := make([]string, 0, len(all))
			for k := range all {
				allCommands = append(allCommands, k)
			}
			allCommandsString := strings.Join(allCommands, " ")
			if commandName != "" {
				message = fmt.Sprintf("Too many arguments to \"%s\" command, expected arguments \"%s\".", commandName, allCommandsString)
			} else {
				message = fmt.Sprintf("Too many arguments, expected arguments \"%s\".", allCommandsString)
			}
		} else if commandName != "" {
			message = fmt.Sprintf("No arguments expected for \"%s\" command, got \"%s\".", commandName, token)
		} else {
			message = fmt.Sprintf("No arguments expected, got \"%s\".", token)
		}

		return err.NewRuntimeError(message)
	}
}

func (input *ArgvInput) parseLongOption(token string) error {
	name := token[2:]
	pos := strings.Index(name, "=")

	if pos != -1 {
		value := name[pos+1:]
		if value == "" {
			helper.Unshift(&input.parsed, value)
		}
		return input.addLongOption(name[0:pos], value)
	} else {
		return input.addLongOption(name, nil)
	}
}

func (input *ArgvInput) parseShortOption(token string) error {
	name := token[1:]

	if len(name) > 1 {
		short := name[0:1]
		if input.definition.HasShortcut(short) {
			opt, err := input.definition.GetOptionForShortcut(short)
			if err != nil {
				return err
			}

			if opt.AcceptValue() {
				// an option with a value (with no space)
				return input.addShortOption(short, name[1:])
			}
		}

		return input.parseShortOptionSet(name)
	} else {
		return input.addShortOption(name, nil)
	}
}

func (input *ArgvInput) parseShortOptionSet(name string) error {
	length := len(name)
	for i := 0; i < length; i++ {
		char := name[i : i+1]
		if !input.definition.HasShortcut(char) {
			return err.NewRuntimeError(fmt.Sprintf("The \"-%s\" option does not exist.", char))
		}

		opt, err := input.definition.GetOptionForShortcut(char)
		if err != nil {
			return err
		}

		if opt.AcceptValue() {
			if length-1 == i {
				err := input.addLongOption(opt.GetName(), nil)
				if err != nil {
					return err
				}
			} else {
				err := input.addLongOption(opt.GetName(), name[i+1:])
				if err != nil {
					return err
				}
			}

			break
		} else {
			err := input.addLongOption(opt.GetName(), nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (input *ArgvInput) addShortOption(shortcut string, value InputType) error {
	if !input.definition.HasShortcut(shortcut) {
		return err.NewRuntimeError(fmt.Sprintf("The \"-%s\" option does not exist.", shortcut))
	}

	opt, err := input.definition.GetOptionForShortcut(shortcut)
	if err != nil {
		return err
	}

	return input.addLongOption(opt.GetName(), value)
}

func (input *ArgvInput) addLongOption(name string, value InputType) error {
	if !input.definition.HasOption(name) {
		if !input.definition.HasNegation(name) {
			return err.NewRuntimeError(fmt.Sprintf("The \"--%s\" option does not exist.", name))
		}

		optName := input.definition.NegationToName(name)
		input.options[optName] = false

		if value != nil && value != "" {
			return err.NewRuntimeError(fmt.Sprintf("The \"--%s\" option does not accept a value.", name))
		}

		return nil
	}

	opt, e := input.definition.GetOption(name)
	if e != nil {
		return e
	}

	if value != nil && value != "" && !opt.AcceptValue() {
		return err.NewRuntimeError(fmt.Sprintf("The \"--%s\" option does not accept a value.", name))
	}

	if (value == nil || value == "") && opt.AcceptValue() && len(input.parsed) > 0 {
		// if option accepts an optional or mandatory argument
		// let's see if there is one provided
		next := helper.Shift(&input.parsed)
		if (next != "" && next[0:1] != "-") || next == "" {
			value = next
		} else {
			helper.Unshift(&input.parsed, next)
		}
	}

	if value == nil || value == "" {
		if opt.IsValueOptional() {
			return err.NewRuntimeError(fmt.Sprintf("The \"--%s\" option requires a value.", name))
		}

		if !opt.IsArray() && !opt.IsValueOptional() {
			value = true
		}
	}

	if opt.IsArray() {
		cur := input.options[name]
		arr, isArr := cur.([]string)
		if !isArr {

		} else {
			arr = append(arr, value.(string))
		}
		input.options[name] = arr
	} else {
		input.options[name] = value
	}

	return nil
}

func (input *ArgvInput) GetFirstArgument() InputType {
	isOption := false
	tokenCount := len(input.tokens)

	for i, token := range input.tokens {
		if token != "" && strings.HasPrefix(token, "-") {
			if strings.Contains(token, "=") || i+1 >= tokenCount {
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

			value := input.options[name]
			if (value == nil || value == "") && !input.definition.HasShortcut(name) {
				//noop
			} else if value != "" && value != nil && input.tokens[i+1] == value {
				isOption = true
			} else {
				name = input.definition.ShortcutToName(name)
				value = input.options[name]

				if value != "" && value != nil && input.tokens[i+1] == value {
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

func (input *ArgvInput) HasParameterOption(value string, onlyParams bool) bool {
	for _, token := range input.tokens {
		if onlyParams && token == "--" {
			return false
		}

		var leading string = value
		if strings.HasPrefix(value, "--") {
			leading = value + "="
		}

		if token == value || (leading != "" && strings.HasPrefix(token, leading)) {
			return true
		}
	}

	return false
}

func (input *ArgvInput) GetParameterOption(value string, defaultValue InputType, onlyParams bool) InputType {
	tokens := make([]string, 0, len(input.tokens))
	copy(tokens, input.tokens)

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
		var leading string = value
		if strings.HasPrefix(value, "--") {
			leading = value + "="
		}

		if leading != "" && strings.HasPrefix(token, leading) {
			return token[len(leading):]
		}
	}

	return nil
}

func (input *ArgvInput) ToString() string {
	re := regexp.MustCompile(`{^(-[^=]+=)(.+)}`)
	tokens := make([]string, 0, len(input.tokens))

	for _, token := range input.tokens {
		match := re.FindStringSubmatch(token)

		if match != nil {
			tokens = append(tokens, match[1]+input.EscapeToken(match[2]))
		} else {
			if token != "" && token[0] != '-' {
				tokens = append(tokens, input.EscapeToken(token))
			} else {
				tokens = append(tokens, token)
			}
		}
	}

	return strings.Join(tokens, " ")
}
