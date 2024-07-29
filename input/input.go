package input

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type InputParser func(self interface{}) error

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

	if definition == nil {
		input.definition = NewInputDefinition(nil, nil)
	} else {
		input.definition = definition
		input.Bind(definition)
		err := input.Parse()
		if err != nil {
			return input, err
		}

		err = input.Validate()
		if err != nil {
			return input, err
		}
	}

	return input, nil
}

func (input *Input) SetDefinition(definition *InputDefinition) error {
	if definition == nil {
		input.definition = NewInputDefinition(nil, nil)
		return nil
	} else {
		input.Bind(definition)
		err := input.Parse()
		if err != nil {
			return err
		}

		return input.Validate()
	}
}

func (input *Input) Bind(definition *InputDefinition) {
	input.arguments = make(map[string]InputType)
	input.options = make(map[string]InputType)
	input.definition = definition
}

func (input *Input) Parse() error {
	panic("Abstract method Input.Parse() not implemented")
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

	arguments := definition.Arguments()
	if arguments == nil {
		return errors.New("input arguments not found")
	}

	missingArguments := make([]string, 0, len(arguments))
	for name := range arguments {
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

func (input *Input) Arguments() map[string]InputType {
	definition := input.definition
	args := make(map[string]InputType)

	if definition != nil {
		for k, v := range definition.ArgumentDefaults() {
			args[k] = v
		}
	}

	for k, v := range input.arguments {
		args[k] = v
	}

	return args
}

func (input *Input) StringArgument(name string) (string, error) {
	definition := input.definition
	if definition == nil {
		return "", errors.New("no InputDefinition found")
	}

	if !definition.HasArgument(name) {
		return "", fmt.Errorf("the \"%s\" argument does not exist", name)
	}

	if input.arguments[name] == nil {
		arg, e := definition.Argument(name)
		if e != nil {
			return "", e
		}

		value := arg.DefaultValue()
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

func (input *Input) ArrayArgument(name string) ([]string, error) {
	definition := input.definition
	if definition == nil {
		return []string{}, errors.New("no InputDefinition found")
	}

	if !definition.HasArgument(name) {
		return []string{}, fmt.Errorf("the \"%s\" argument does not exist", name)
	}

	if input.arguments[name] == nil {
		arg, e := definition.Argument(name)
		if e != nil {
			return []string{}, e
		}

		value := arg.DefaultValue()
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
		return errors.New("no InputDefinition found")
	}

	if !definition.HasArgument(name) {
		return fmt.Errorf("the \"%s\" argument does not exist", name)
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

func (input *Input) Options() map[string]InputType {
	definition := input.definition
	opts := make(map[string]InputType)

	if definition != nil {
		for k, v := range definition.OptionDefaults() {
			opts[k] = v
		}
	}

	for k, v := range input.options {
		opts[k] = v
	}

	return opts
}

func (input *Input) BoolOption(name string) (bool, error) {
	definition := input.definition
	if definition == nil {
		return false, errors.New("no InputDefinition found")
	}

	if !definition.HasOption(name) {
		return false, fmt.Errorf("the \"%s\" option does not exist", name)
	}

	if input.options[name] == nil {
		opt, e := definition.Option(name)
		if e != nil {
			return false, e
		}

		value := opt.DefaultValue()
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

func (input *Input) StringOption(name string) (string, error) {
	definition := input.definition
	if definition == nil {
		return "", errors.New("no InputDefinition found")
	}

	if !definition.HasOption(name) {
		return "", fmt.Errorf("the \"%s\" option does not exist", name)
	}

	if input.options[name] == nil {
		opt, e := definition.Option(name)
		if e != nil {
			return "", e
		}

		value := opt.DefaultValue()
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

func (input *Input) ArrayOption(name string) ([]string, error) {
	definition := input.definition
	if definition == nil {
		return []string{}, errors.New("no InputDefinition found")
	}

	if !definition.HasOption(name) {
		return []string{}, fmt.Errorf("the \"%s\" option does not exist", name)
	}

	if input.options[name] == nil {
		opt, e := definition.Option(name)
		if e != nil {
			return []string{}, e
		}

		value := opt.DefaultValue()
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
		return errors.New("no InputDefinition found")
	}

	if definition.HasNegation(name) {
		input.options[definition.NegationToName(name)] = value
		return nil
	}

	if !definition.HasOption(name) {
		return fmt.Errorf("the \"%s\" option does not exist", name)
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

func (input *Input) Stream() *os.File {
	return input.stream
}

func (input *Input) runArgumentValidators() error {
	definition := input.definition
	if definition == nil {
		return errors.New("no InputDefinition found")
	}

	args := definition.Arguments()
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
		return errors.New("no InputDefinition found")
	}

	opts := definition.Options()
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
