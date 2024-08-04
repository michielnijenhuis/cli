package cli

import (
	"fmt"
	"math"
	"strings"
)

type InputType any

type InputValidator func(value InputType) error

type InputDefinition struct {
	Arguments            []*InputArgument
	Options              []*InputOption
	requiredCount        uint
	lastArrayArgument    *InputArgument
	lastOptionalArgument *InputArgument
	negations            map[string]string
	shortcuts            map[string]string
}

func (d *InputDefinition) SetDefinition(arguments []*InputArgument, options []*InputOption) {
	d.SetArguments(arguments)
	d.SetOptions(options)
}

func (d *InputDefinition) SetArguments(arguments []*InputArgument) {
	d.requiredCount = 0
	d.lastOptionalArgument = nil
	d.lastArrayArgument = nil
	d.Arguments = make([]*InputArgument, 0)
	d.AddArguments(arguments)
}

func (d *InputDefinition) AddArguments(arguments []*InputArgument) {
	if arguments == nil {
		return
	}

	for _, arg := range arguments {
		d.AddArgument(arg)
	}
}

func (d *InputDefinition) AddArgument(argument *InputArgument) {
	if argument == nil {
		return
	}

	if d.Arguments == nil {
		d.Arguments = make([]*InputArgument, 0)
	}

	if d.HasArgument(argument.Name) {
		panic(fmt.Sprintf("An argument with name \"%s\" already exists.", argument.Name))
	}

	if d.lastArrayArgument != nil {
		panic(fmt.Sprintf("Cannot add a required argument \"%s\" after an array argument \"%s\".", argument.Name, d.lastArrayArgument.Name))
	}

	if argument.IsRequired() && d.lastOptionalArgument != nil {
		panic(fmt.Sprintf("Cannot add a required argument \"%s\" after an optional one \"%s\".", argument.Name, d.lastOptionalArgument.Name))
	}

	if argument.IsArray() {
		d.lastArrayArgument = argument
	}

	if argument.IsRequired() {
		d.requiredCount += 1
	} else {
		d.lastOptionalArgument = argument
	}

	d.Arguments = append(d.Arguments, argument)
}

func (d *InputDefinition) HasArgument(name string) bool {
	a, _ := d.Argument(name)
	return a != nil
}

func (d *InputDefinition) Argument(name string) (*InputArgument, error) {
	if d.Arguments == nil {
		return nil, fmt.Errorf("the \"%s\" argument does not exist", name)
	}

	for _, a := range d.Arguments {
		if a.Name == name {
			return a, nil
		}
	}

	return nil, fmt.Errorf("the \"%s\" argument does not exist", name)
}

func (d *InputDefinition) GetArguments() []*InputArgument {
	s := make([]*InputArgument, 0, len(d.Arguments))
	s = append(s, d.Arguments...)
	return s
}

func (d *InputDefinition) GetOptions() []*InputOption {
	s := make([]*InputOption, 0, len(d.Options))
	s = append(s, d.Options...)
	return s
}

func (d *InputDefinition) ArgumentByIndex(index uint) (*InputArgument, error) {
	count := uint(len(d.Arguments))

	if index >= count {
		return nil, fmt.Errorf("argument index out of bounds. Received \"%d\", but only \"%d\" arguments are found", index, count)
	}

	return d.Arguments[index], nil
}

func (d *InputDefinition) ArgumentCount() uint {
	if d.lastArrayArgument != nil {
		return uint(math.Inf(1))
	}

	return d.requiredCount
}

func (d *InputDefinition) ArgumentRequiredCount() uint {
	return d.requiredCount
}

func (d *InputDefinition) ArgumentDefaults() map[string]InputType {
	if d.Arguments == nil {
		return nil
	}

	m := make(map[string]InputType)

	for _, arg := range d.Arguments {
		m[arg.Name] = arg.DefaultValue
	}

	return m
}

func (d *InputDefinition) SetOptions(options []*InputOption) {
	d.shortcuts = make(map[string]string)
	d.negations = make(map[string]string)
	d.Options = make([]*InputOption, 0)
	d.AddOptions(options)
}

func (d *InputDefinition) AddOptions(options []*InputOption) {
	if options == nil {
		return
	}

	for _, option := range options {
		d.AddOption(option)
	}
}

func (d *InputDefinition) AddOption(option *InputOption) {
	name := option.Name

	if o, _ := d.Option(name); o != nil && !option.Equals(o) {
		panic(fmt.Sprintf("an option named \"%s\" already exists", name))
	}

	if d.negations[name] != "" {
		panic(fmt.Sprintf("an option named \"%s\" already exists", name))
	}

	shortcut := option.Shortcut
	if shortcut != "" {
		shortcuts := strings.Split(shortcut, "|")
		for _, s := range shortcuts {
			optName, ok := d.shortcuts[s]
			if ok {
				opt, _ := d.Option(optName)
				if opt != nil && !option.Equals(opt) {
					panic(fmt.Sprintf("An option with shortcut \"%s\" already exists.", s))
				}
			}
		}

		for _, s := range shortcuts {
			d.shortcuts[s] = name
		}
	}

	d.Options = append(d.Options, option)

	if option.IsNegatable() {
		negatedName := fmt.Sprintf("no-%s", name)

		if negated, _ := d.Option(negatedName); negated != nil {
			panic(fmt.Sprintf("An option named \"%s\" already exists.", negatedName))
		}

		d.negations[negatedName] = name
	}
}

func (d *InputDefinition) HasOption(name string) bool {
	o, _ := d.Option(name)
	return o != nil
}

func (d *InputDefinition) Option(name string) (*InputOption, error) {
	if d.Options == nil {
		return nil, fmt.Errorf("the \"--%s\" option does not exist", name)
	}

	for _, o := range d.Options {
		if o.Name == name {
			return o, nil
		}
	}

	return nil, fmt.Errorf("the \"--%s\" option does not exist", name)
}

func (d *InputDefinition) HasShortcut(name string) bool {
	return d.shortcuts[name] != ""
}

func (d *InputDefinition) HasNegation(name string) bool {
	return d.negations[name] != ""
}

func (d *InputDefinition) OptionForShortcut(shortcut string) (*InputOption, error) {
	opt, err := d.Option(d.ShortcutToName(shortcut))
	if err != nil {
		return nil, fmt.Errorf("the \"-%s\" option does not exist", shortcut)
	}
	return opt, nil
}

func (d *InputDefinition) OptionDefaults() map[string]InputType {
	values := make(map[string]InputType)
	for _, option := range d.Options {
		values[option.Name] = option.DefaultValue
	}

	return values
}

func (d *InputDefinition) ShortcutToName(shortcut string) string {
	return d.shortcuts[shortcut]
}

func (d *InputDefinition) NegationToName(negation string) string {
	return d.negations[negation]
}

func (d *InputDefinition) Synopsis(short bool) string {
	elements := make([]string, 0)
	options := d.Options

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

				segments = append(segments, strings.ToUpper(o.Name))

				if o.IsValueOptional() {
					segments = append(segments, "]")
				} else {
					segments = append(segments, "")
				}

				value = fmt.Sprintf(" %s", strings.Join(segments, ""))
			}

			shortcut := ""
			if o.Shortcut != "" {
				shortcut = fmt.Sprintf("-%s|", o.Shortcut)
			}

			negation := ""
			if o.IsNegatable() {
				negation = fmt.Sprintf("|--no-%s", o.Name)
			}

			elements = append(elements, fmt.Sprintf("[%s--%s%s%s]", shortcut, o.Name, value, negation))
		}
	}

	if len(elements) > 0 && len(d.Arguments) > 0 {
		elements = append(elements, "[--]")
	}

	tail := ""
	for _, arg := range d.Arguments {
		element := fmt.Sprintf("<%s>", arg.Name)

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
