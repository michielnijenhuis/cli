package cli

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

type InputType any

type InputValidator func(value InputType) error

type InputDefinition struct {
	arguments            []Arg
	flags                []Flag
	requiredCount        uint
	firstArgument        Arg
	lastArrayArgument    *ArrayArg
	lastOptionalArgument Arg
	negations            map[string]string
	shortcuts            map[string]string
}

func (d *InputDefinition) SetDefinition(arguments []Arg, flags []Flag) error {
	if err := d.SetArguments(arguments); err != nil {
		return err
	}

	if err := d.SetFlags(flags); err != nil {
		return err
	}

	return nil
}

func (d *InputDefinition) SetArguments(arguments []Arg) error {
	d.requiredCount = 0
	d.firstArgument = nil
	d.lastOptionalArgument = nil
	d.lastArrayArgument = nil
	d.arguments = make([]Arg, 0)
	return d.AddArguments(arguments)
}

func (d *InputDefinition) AddArguments(arguments []Arg) error {
	for i, arg := range arguments {
		if i == 0 {
			d.firstArgument = arg
		}

		if err := d.AddArgument(arg); err != nil {
			return err
		}
	}

	return nil
}

func (d *InputDefinition) AddArgument(argument Arg) error {
	if d.arguments == nil {
		d.arguments = make([]Arg, 0)
	}

	if d.firstArgument == nil {
		d.firstArgument = argument
	}

	name := argument.GetName()

	if d.HasArgument(name) {
		return fmt.Errorf("an argument with name \"%s\" already exists", name)
	}

	if d.HasFlag(name) {
		return fmt.Errorf("a flag with name \"%s\" already exists", name)
	}

	if d.lastArrayArgument != nil {
		return fmt.Errorf("cannot add a required argument \"%s\" after an array argument \"%s\"", name, d.lastArrayArgument.Name)
	}

	if argument.IsRequired() && d.lastOptionalArgument != nil {
		return fmt.Errorf("cannot add a required argument \"%s\" after an flagal one \"%s\"", name, d.lastOptionalArgument.GetName())
	}

	if arr, ok := argument.(*ArrayArg); ok {
		d.lastArrayArgument = arr
	}

	if argument.IsRequired() {
		d.requiredCount += 1
	} else {
		d.lastOptionalArgument = argument
	}

	d.arguments = append(d.arguments, argument)
	return nil
}

func (d *InputDefinition) HasArgument(name string) bool {
	a, _ := d.Argument(name)
	return a != nil
}

func (d *InputDefinition) Argument(name string) (Arg, error) {
	for _, a := range d.arguments {
		if a.GetName() == name {
			return a, nil
		}
	}

	return nil, fmt.Errorf("the \"%s\" argument does not exist", name)
}

func (d *InputDefinition) GetArguments() []Arg {
	s := make([]Arg, 0, len(d.arguments))
	s = append(s, d.arguments...)
	return s
}

func (d *InputDefinition) GetFlags() []Flag {
	s := make([]Flag, 0, len(d.flags))
	s = append(s, d.flags...)
	return s
}

func (d *InputDefinition) ArgumentByIndex(index uint) (Arg, error) {
	count := uint(len(d.arguments))

	if index >= count {
		return nil, fmt.Errorf("argument index out of bounds. Received \"%d\", but only \"%d\" arguments are found", index, count)
	}

	return d.arguments[index], nil
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

func (d *InputDefinition) SetFlags(flags []Flag) error {
	d.shortcuts = make(map[string]string)
	d.negations = make(map[string]string)
	d.flags = make([]Flag, 0)
	return d.AddFlags(flags)
}

func (d *InputDefinition) AddFlags(flags []Flag) error {
	for _, flag := range flags {
		if err := d.AddFlag(flag); err != nil {
			return err
		}
	}

	return nil
}

func (d *InputDefinition) AddFlag(flag Flag) error {
	if flag == nil {
		return errors.New("flag cannot be nil")
	}

	if d.flags == nil {
		d.flags = make([]Flag, 0)
	}

	name := flag.GetName()

	if d.HasFlag(name) {
		return fmt.Errorf("a flag named \"%s\" already exists", name)
	}

	if d.HasArgument(name) {
		return fmt.Errorf("flag cannot be added, as an argument named \"%s\" already exists", name)
	}

	if d.negations[name] != "" {
		return fmt.Errorf("a flag named \"%s\" already exists", name)
	}

	shortcuts := flag.GetShortcuts()
	if shortcuts != nil {
		for _, s := range shortcuts {
			name, ok := d.shortcuts[s]
			if ok {
				f, _ := d.Flag(name)
				if f != nil && FlagEquals(flag, f) {
					return fmt.Errorf("a flag with shortcut \"%s\" already exists", s)
				}
			}
		}

		for _, s := range shortcuts {
			if d.shortcuts == nil {
				d.shortcuts = make(map[string]string)
			}

			d.shortcuts[s] = name
		}
	}

	d.flags = append(d.flags, flag)

	if FlagIsNegatable(flag) {
		negatedName := fmt.Sprintf("no-%s", name)

		if negated, _ := d.Flag(negatedName); negated != nil {
			return fmt.Errorf("a flag named \"%s\" already exists", negatedName)
		}

		if d.negations == nil {
			d.negations = make(map[string]string)
		}

		d.negations[negatedName] = name
	}

	return nil
}

func (d *InputDefinition) HasFlag(name string) bool {
	o, _ := d.Flag(name)
	return o != nil
}

func (d *InputDefinition) Flag(name string) (Flag, error) {
	for _, o := range d.flags {
		if o.GetName() == name {
			return o, nil
		}
	}

	return nil, fmt.Errorf("the \"--%s\" flag does not exist", name)
}

func (d *InputDefinition) HasShortcut(name string) bool {
	return d.shortcuts[name] != ""
}

func (d *InputDefinition) HasNegation(name string) bool {
	return d.negations[name] != ""
}

func (d *InputDefinition) FlagForShortcut(shortcut string) (Flag, error) {
	opt, err := d.Flag(d.ShortcutToName(shortcut))
	if err != nil {
		return nil, fmt.Errorf("the \"-%s\" flag does not exist", shortcut)
	}
	return opt, nil
}

func (d *InputDefinition) ShortcutToName(shortcut string) string {
	if d.shortcuts == nil {
		d.shortcuts = make(map[string]string)
	}

	return d.shortcuts[shortcut]
}

func (d *InputDefinition) NegationToName(negation string) string {
	if d.negations == nil {
		d.negations = make(map[string]string)
	}

	return d.negations[negation]
}

func (d *InputDefinition) Synopsis(short bool) string {
	elements := make([]string, 0)
	flags := d.flags

	if short && len(flags) > 0 {
		elements = append(elements, "[flags]")
	} else if !short {
		for _, f := range flags {
			value := ""

			if FlagAcceptsValue(f) {
				segments := make([]string, 0, 3)

				if FlagValueIsOptional(f) {
					segments = append(segments, "[")
				} else {
					segments = append(segments, "")
				}

				segments = append(segments, strings.ToUpper(f.GetName()))

				if FlagValueIsOptional(f) {
					segments = append(segments, "]")
				} else {
					segments = append(segments, "")
				}

				value = fmt.Sprintf(" %s", strings.Join(segments, ""))
			}

			shortcut := ""
			if s := f.GetShortcutString(); s != "" {
				shortcut = fmt.Sprintf("-%s|", s)
			}

			negation := ""
			if FlagIsNegatable(f) {
				negation = fmt.Sprintf("|--no-%s", f.GetName())
			}

			elements = append(elements, fmt.Sprintf("[%s--%s%s%s]", shortcut, f.GetName(), value, negation))
		}
	}

	if len(elements) > 0 && len(d.arguments) > 0 {
		elements = append(elements, "[--]")
	}

	tail := ""
	for _, arg := range d.arguments {
		element := fmt.Sprintf("<%s>", arg.GetName())

		if _, ok := arg.(*ArrayArg); ok {
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
