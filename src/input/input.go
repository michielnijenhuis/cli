package input

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	err "github.com/michielnijenhuis/cli/error"
)

type Input struct {
	definition  *InputDefinition
	stream      *os.File
	options     map[string]InputType
	arguments   map[string]InputType
	interactive bool
}

func NewInput(definition *InputDefinition) (*Input, error) {
	input := &Input{
		definition:  nil,
		stream:      nil,
		options:     make(map[string]InputType),
		arguments:   make(map[string]InputType),
		interactive: true,
	}

	var err error = nil

	if definition == nil {
		input.definition = NewInputDefinition(nil, nil)
	} else {
		input.Bind(definition)
		err = input.Validate()
	}

	return input, err
}

func (input *Input) SetDefinition(definition *InputDefinition) {
	if definition == nil {
		input.definition = NewInputDefinition(nil, nil)
	} else {
		input.Bind(definition)
		input.Validate()
	}
}

func (input *Input) GetFirstArgument() InputType {
	panic("Abstract method Input.GetFirstArgument() is not implemented.")
}

func (input *Input) HasParameterOption(value string, onlyParams bool) bool {
	panic("Abstract method Input.HasParameterOption() is not implemented.")
}

func (input *Input) GetParameterOption(value string, defaultValue InputType, onlyParams bool) InputType {
	panic("Abstract method Input.GetParameterOption() is not implemented.")
}

func (input *Input) ToString() string {
	panic("Abstract method Input.ToString() is not implemented.")
}

func (input *Input) Bind(definition *InputDefinition) error {
	input.arguments = make(map[string]InputType)
	input.options = make(map[string]InputType)
	input.definition = definition

	return input.Parse()
}

func (input *Input) Parse() error {
	panic("Abstract method Input.Parse() is not implemented.")
}

func (input *Input) Validate() error {
	definition := input.definition
	if definition == nil {
		return errors.New("inputDefinition not found")
	}

	givenArguments := input.arguments
	if givenArguments == nil {
		return errors.New("given arguments not found")
	}

	arguments := definition.GetArguments()
	if arguments == nil {
		return errors.New("input arguments not found")
	}

	missingArguments := make([]string, 0, len(arguments))
	for name := range arguments {
		if givenArguments[name] != nil {
			continue
		}

		argument, _ := definition.GetArgument(name)
		if argument != nil && argument.IsRequired() {
			missingArguments = append(missingArguments, name)
		}
	}

	if len(missingArguments) > 0 {
		return err.NewRuntimeError(
			fmt.Sprintf("Not enough arguments (missing: \"%s\").", strings.Join(missingArguments, ", ")),
		)
	}

	validationError := input.runOptionValidators()
	if validationError != nil {
		return validationError
	}

	return input.runArgumentValidators()
}

func (input *Input) IsInteractive() bool {
	return input.interactive
}

func (input *Input) SetInteractive(interactive bool) {
	input.interactive = interactive
}

func (input *Input) GetArguments() map[string]InputType {
	definition := input.definition
	args := make(map[string]InputType)

	if definition != nil {
		for k, v := range definition.GetArgumentDefaults() {
			args[k] = v
		}
	}

	for k, v := range input.arguments {
		args[k] = v
	}

	return args
}

func (input *Input) GetStringArgument(name string) (string, error) {
	definition := input.definition
	if definition == nil {
		return "", err.NewRuntimeError("no InputDefinition found")
	}

	if !definition.HasArgument(name) {
		return "", err.NewInvalidArgumentError(fmt.Sprintf("The \"%s\" argument does not exist.", name))
	}

	if input.arguments[name] == nil {
		arg, e := definition.GetArgument(name)
		if e != nil {
			return "", e
		}

		value := arg.GetDefaultValue()
		str, isStr := value.(string)
		if !isStr {
			return "", nil
		}

		return str, nil
	}

	value := input.arguments[name]
	str, isStr := value.(string)
	if !isStr {
		return "", nil
	}

	return str, nil
}

func (input *Input) GetArrayArgument(name string) ([]string, error) {
	definition := input.definition
	if definition == nil {
		return []string{}, err.NewRuntimeError("No InputDefinition found")
	}

	if !definition.HasArgument(name) {
		return []string{}, err.NewInvalidArgumentError(fmt.Sprintf("The \"%s\" argument does not exist.", name))
	}

	if input.arguments[name] == nil {
		arg, e := definition.GetArgument(name)
		if e != nil {
			return []string{}, e
		}

		value := arg.GetDefaultValue()
		arr, isArr := value.([]string)
		if !isArr {
			return []string{}, nil
		}

		return arr, nil
	}

	value := input.arguments[name]
	arr, isArr := value.([]string)
	if !isArr {
		return []string{}, nil
	}

	return arr, nil
}

func (input *Input) SetArgument(name string, value InputType) error {
	definition := input.definition
	if definition == nil {
		return err.NewRuntimeError("No InputDefinition found.")
	}

	if !definition.HasArgument(name) {
		return err.NewInvalidArgumentError(fmt.Sprintf("The \"%s\" argument does not exist.", name))
	}

	input.arguments[name] = value
	return nil
}

func (input *Input) HasArgument(name string) bool {
	definition := input.definition
	if definition == nil {
		return false
	}

	return definition.HasArgument(name)
}

func (input *Input) GetOptions() map[string]InputType {
	definition := input.definition
	opts := make(map[string]InputType)

	if definition != nil {
		for k, v := range definition.GetOptionDefaults() {
			opts[k] = v
		}
	}

	for k, v := range input.options {
		opts[k] = v
	}

	return opts
}

func (input *Input) GetBoolOption(name string) (bool, error) {
	definition := input.definition
	if definition == nil {
		return false, err.NewRuntimeError("No InputDefinition found")
	}

	if !definition.HasOption(name) {
		return false, err.NewInvalidArgumentError(fmt.Sprintf("The \"%s\" option does not exist.", name))
	}

	if input.options[name] == nil {
		opt, e := definition.GetOption(name)
		if e != nil {
			return false, e
		}

		value := opt.GetDefaultValue()
		boolval, isBool := value.(bool)
		if !isBool {
			return false, nil
		}

		if definition.HasNegation(name) {
			return !boolval, nil
		}

		return boolval, nil
	}

	value := input.options[name]
	boolval, isBool := value.(bool)
	if !isBool {
		return false, nil
	}

	if definition.HasNegation(name) {
		return !boolval, nil
	}

	return boolval, nil
}

func (input *Input) GetStringOption(name string) (string, error) {
	definition := input.definition
	if definition == nil {
		return "", err.NewRuntimeError("No InputDefinition found")
	}

	if !definition.HasOption(name) {
		return "", err.NewInvalidArgumentError(fmt.Sprintf("The \"%s\" option does not exist.", name))
	}

	if input.options[name] == nil {
		opt, e := definition.GetOption(name)
		if e != nil {
			return "", e
		}

		value := opt.GetDefaultValue()
		str, isStr := value.(string)
		if !isStr {
			return "", nil
		}

		return str, nil
	}

	value := input.options[name]
	str, isStr := value.(string)
	if !isStr {
		return "", nil
	}

	return str, nil
}

func (input *Input) GetArrayOption(name string) ([]string, error) {
	definition := input.definition
	if definition == nil {
		return []string{}, err.NewRuntimeError("No InputDefinition found")
	}

	if !definition.HasOption(name) {
		return []string{}, err.NewInvalidArgumentError(fmt.Sprintf("The \"%s\" option does not exist.", name))
	}

	if input.options[name] == nil {
		opt, e := definition.GetOption(name)
		if e != nil {
			return []string{}, e
		}

		value := opt.GetDefaultValue()
		arr, isArr := value.([]string)
		if !isArr {
			return []string{}, nil
		}

		return arr, nil
	}

	value := input.options[name]
	arr, isArr := value.([]string)
	if !isArr {
		return []string{}, nil
	}

	return arr, nil
}

func (input *Input) SetOption(name string, value InputType) error {
	definition := input.definition
	if definition == nil {
		return err.NewRuntimeError("No InputDefinition found.")
	}

	if definition.HasNegation(name) {
		input.options[definition.NegationToName(name)] = value
		return nil
	}

	if !definition.HasOption(name) {
		return err.NewInvalidArgumentError(fmt.Sprintf("The \"%s\" option does not exist.", name))
	}

	input.options[name] = value
	return nil
}

func (input *Input) HasOption(name string) bool {
	definition := input.definition
	if definition == nil {
		return false
	}

	return definition.HasOption(name) || definition.HasNegation(name)
}

func (input *Input) SetStream(stream *os.File) {
	input.stream = stream
}

func (input *Input) GetStream() *os.File {
	return input.stream
}

func (input *Input) runArgumentValidators() error {
	definition := input.definition
	if definition == nil {
		return err.NewRuntimeError("No InputDefinition found.")
	}

	args := definition.GetArguments()
	for name, arg := range args {
		value := input.arguments[name]

		validator := arg.validator
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

func (input *Input) runOptionValidators() error {
	definition := input.definition
	if definition == nil {
		return err.NewRuntimeError("No InputDefinition found.")
	}

	opts := definition.GetOptions()
	for name, opt := range opts {
		value := input.options[name]

		validator := opt.validator
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

func (input *Input) EscapeToken(token string) string {
	re := regexp.MustCompile(`{^[\w-]+}`)
	if re.MatchString(token) {
		return token
	}
	re2 := regexp.MustCompile(`'`)
	return re2.ReplaceAllString(token, "'\\''")
}
