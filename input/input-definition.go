package input

import (
	"fmt"
	"math"
	"strings"

	Error "github.com/michielnijenhuis/cli/error"
)

type InputType interface{}

type InputValidator func(value InputType) error

type InputDefinition struct {
	arguments            map[string]*InputArgument
	requiredCount        uint
	lastArrayArgument    *InputArgument
	lastOptionalArgument *InputArgument
	options              map[string]*InputOption
	negations            map[string]string
	shortcuts            map[string]string
}

func NewInputDefinition(arguments []*InputArgument, options []*InputOption) *InputDefinition {
	definition := &InputDefinition{
		arguments:            nil,
		requiredCount:        0,
		lastArrayArgument:    nil,
		lastOptionalArgument: nil,
		options:              nil,
		negations:            nil,
		shortcuts:            nil,
	}

	definition.SetDefinition(arguments, options)

	return definition
}

func (definition *InputDefinition) SetDefinition(arguments []*InputArgument, options []*InputOption) {
	definition.SetArguments(arguments)
	definition.SetOptions(options)
}

func (input *InputDefinition) SetArguments(arguments []*InputArgument) {
	input.arguments = make(map[string]*InputArgument)
	input.requiredCount = 0
	input.lastOptionalArgument = nil
	input.lastArrayArgument = nil
	input.AddArguments(arguments)
}

func (input *InputDefinition) AddArguments(arguments []*InputArgument) {
	if arguments == nil {
		return
	}

	for _, arg := range arguments {
		input.AddArgument(arg)
	}
}

func (input *InputDefinition) AddArgument(argument *InputArgument) {
	if argument == nil {
		return
	}

	if input.arguments[argument.GetName()] != nil {
		panic(fmt.Sprintf("An argument with name \"%s\" already exists.", argument.GetName()))
	}

	if input.lastArrayArgument != nil {
		panic(fmt.Sprintf("Cannot add a required argument \"%s\" after an array argument \"%s\".", argument.GetName(), input.lastArrayArgument.GetName()))
	}

	if argument.IsRequired() && input.lastOptionalArgument != nil {
		panic(fmt.Sprintf("Cannot add a required argument \"%s\" after an optional one \"%s\".", argument.GetName(), input.lastOptionalArgument.GetName()))
	}

	if argument.IsArray() {
		input.lastArrayArgument = argument
	}

	if argument.IsRequired() {
		input.requiredCount += 1
	} else {
		input.lastOptionalArgument = argument
	}

	input.arguments[argument.GetName()] = argument
}

func (input *InputDefinition) HasArgument(name string) bool {
	return input.arguments[name] != nil
}

func (definition *InputDefinition) GetArgument(name string) (*InputArgument, error) {
	if !definition.HasArgument(name) {
		return nil, Error.NewInvalidArgumentError(
			fmt.Sprintf("The \"%s\" argument does not exist.", name),
		)
	}

	return definition.arguments[name], nil
}

func (definition *InputDefinition) GetArgumentByIndex(index uint) (*InputArgument, error) {
	args := definition.GetArguments()
	count := uint(len(args))

	if index >= count {
		return nil, Error.NewInvalidArgumentError(fmt.Sprintf("Argument index out of bounds. Received \"%d\", but only \"%d\" arguments are found.", index, count))
	}

	var i uint = 0
	for _, arg := range args {
		if i == index {
			return arg, nil
		}

		i++
	}

	panic("Unreachable.")
}

func (definition *InputDefinition) GetArguments() map[string]*InputArgument {
	return definition.arguments
}

func (definition *InputDefinition) GetArgumentsArray() []*InputArgument {
	args := make([]*InputArgument, 0, len(definition.arguments))
	for _, arg := range definition.GetArguments() {
		args = append(args, arg)
	}
	return args
}

func (definition *InputDefinition) GetArgumentCount() uint {
	if definition.lastArrayArgument != nil {
		return uint(math.Inf(1))
	}

	return definition.requiredCount
}

func (definition *InputDefinition) GetArgumentRequiredCount() uint {
	return definition.requiredCount
}

func (definition *InputDefinition) GetArgumentDefaults() map[string]InputType {
	m := make(map[string]InputType)

	for name, arg := range definition.arguments {
		m[name] = arg.GetDefaultValue()
	}

	return m
}

func (definition *InputDefinition) SetOptions(options []*InputOption) {
	definition.options = make(map[string]*InputOption)
	definition.shortcuts = make(map[string]string)
	definition.negations = make(map[string]string)
	definition.AddOptions(options)
}

func (definition *InputDefinition) AddOptions(options []*InputOption) {
	if options == nil {
		return
	}

	for _, option := range options {
		definition.AddOption(option)
	}
}

func (definition *InputDefinition) AddOption(option *InputOption) {
	name := option.GetName()

	if definition.options[name] != nil && !option.Equals(definition.options[name]) {
		panic(fmt.Sprintf("An option named \"%s\" already exists.", name))
	}

	if definition.negations[name] != "" {
		panic(fmt.Sprintf("An option named \"%s\" already exists.", name))
	}

	shortcut := option.GetShortcut()
	if shortcut != "" {
		shortcuts := strings.Split(shortcut, "|")
		for _, s := range shortcuts {
			if definition.shortcuts[s] != "" && !option.Equals(definition.options[definition.shortcuts[s]]) {
				panic(fmt.Sprintf("An option with shortcut \"%s\" already exists.", s))
			}
		}

		for _, s := range shortcuts {
			definition.shortcuts[s] = name
		}
	}

	definition.options[name] = option

	if option.IsNegatable() {
		negatedName := fmt.Sprintf("no-%s", name)

		if definition.options[negatedName] != nil {
			panic(fmt.Sprintf("An option named \"%s\" already exists.", negatedName))
		}

		definition.negations[negatedName] = name
	}
}

func (definition *InputDefinition) HasOption(name string) bool {
	return definition.options[name] != nil
}

func (definition *InputDefinition) GetOption(name string) (*InputOption, error) {
	if !definition.HasOption(name) {
		return nil, Error.NewInvalidArgumentError(fmt.Sprintf("The \"--%s\" option does not exist.", name))
	}

	return definition.options[name], nil
}

func (definition *InputDefinition) GetOptions() map[string]*InputOption {
	return definition.options
}

func (definition *InputDefinition) GetOptionsArray() []*InputOption {
	opts := make([]*InputOption, 0, len(definition.options))
	for _, opt := range definition.GetOptions() {
		opts = append(opts, opt)
	}
	return opts
}

func (definition *InputDefinition) HasShortcut(name string) bool {
	return definition.shortcuts[name] != ""
}

func (definition *InputDefinition) HasNegation(name string) bool {
	return definition.negations[name] != ""
}

func (definition *InputDefinition) GetOptionForShortcut(shortcut string) (*InputOption, error) {
	opt, err := definition.GetOption(definition.ShortcutToName(shortcut))
	if err != nil {
		return nil, Error.NewRuntimeError(fmt.Sprintf("The \"-%s\" option does not exist.", shortcut))
	}
	return opt, nil
}

func (definition *InputDefinition) GetOptionDefaults() map[string]InputType {
	values := make(map[string]InputType)
	options := definition.GetOptions()
	for _, option := range options {
		values[option.GetName()] = option.GetDefaultValue()
	}

	return values
}

func (definition *InputDefinition) ShortcutToName(shortcut string) string {
	return definition.shortcuts[shortcut]
}

func (definition *InputDefinition) NegationToName(negation string) string {
	return definition.negations[negation]
}

func (definition *InputDefinition) GetSynopsis(short bool) string {
	elements := make([]string, 0)
	options := definition.GetOptions()

	if short && len(options) > 0 {
		elements = append(elements, "[options]")
	} else if !short {
		for _, o := range options {
			value := ""

			if o.AcceptValue() {
				segments := make([]string, 0, 3)

				if o.IsValueOptional() {
					segments = append(segments, "[")
				} else {
					segments = append(segments, "")
				}

				segments = append(segments, strings.ToUpper(o.GetName()))

				if o.IsValueOptional() {
					segments = append(segments, "]")
				} else {
					segments = append(segments, "")
				}

				value = fmt.Sprintf(" %s", strings.Join(segments, ""))
			}

			shortcut := ""
			if o.GetShortcut() != "" {
				shortcut = fmt.Sprintf("-%s|", o.GetShortcut())
			}

			negation := ""
			if o.IsNegatable() {
				negation = fmt.Sprintf("|--no-%s", o.GetName())
			}

			elements = append(elements, fmt.Sprintf("[%s--%s%s%s]", shortcut, o.GetName(), value, negation))
		}
	}

	if len(elements) > 0 && len(definition.GetArguments()) > 0 {
		elements = append(elements, "[--]")
	}

	tail := ""
	for _, arg := range definition.GetArguments() {
		element := fmt.Sprintf("<%s>", arg.GetName())

		if arg.IsArray() {
			element += "..."
		}

		if !arg.IsRequired() {
			element = "[" + element
			tail += "]"
		}

		elements = append(elements, element)
	}

	return strings.Join(elements, " ") + tail
}
