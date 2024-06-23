package input

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	INPUT_OPTION_BOOLEAN   = 1
	INPUT_OPTION_REQUIRED  = 2
	INPUT_OPTION_OPTIONAL  = 4
	INPUT_OPTION_IS_ARRAY  = 8
	INPUT_OPTION_NEGATABLE = 16
)

type InputOptionMode uint8

type InputOption struct {
	name         string
	shortcut     string
	description  string
	mode         InputOptionMode
	defaultValue InputType
	validator    InputValidator
}

func NewInputOption(name string, shortcut string, mode InputOptionMode, description string, defaultValue InputType, validator InputValidator) *InputOption {
	name, _ = strings.CutPrefix(name, "--")

	if name == "" {
		panic("An option name cannot be empty.")
	}

	if shortcut != "" {
		re := regexp.MustCompile(`{(\})-?}`) // VERIFY: maybe simply split by '|'?

		shortcuts := re.Split(shortcut, -1)
		for i, s := range shortcuts {
			shortcuts[i] = strings.TrimSpace(strings.TrimLeft(s, "-"))
		}

		set := make([]string, 0, len(shortcuts))
		for _, s := range shortcuts {
			if s != "" {
				set = append(set, s)
			}
		}

		shortcut = strings.Join(set, "|")
	}

	if mode == 0 {
		mode = INPUT_OPTION_OPTIONAL
	} else if mode >= INPUT_OPTION_NEGATABLE<<1 || mode < 1 {
		panic(fmt.Sprintf("Option mode \"%d\" is not valid.", mode))
	}

	o := &InputOption{
		name:         name,
		shortcut:     shortcut,
		description:  description,
		mode:         mode,
		defaultValue: nil,
		validator:    validator,
	}

	if o.IsArray() && !o.AcceptValue() {
		panic("Impossible to have an option mode IS_ARRAY if the option does not accept a value.")
	}

	if o.IsNegatable() && o.AcceptValue() {
		panic("Impossible to have an option mode NEGATABLE if the option also accepts a value.")
	}

	o.SetDefaultValue(defaultValue)

	return o
}

func (o *InputOption) IsArray() bool {
	return (o.mode & INPUT_OPTION_IS_ARRAY) == INPUT_OPTION_IS_ARRAY
}

func (o *InputOption) IsNegatable() bool {
	return (o.mode & INPUT_OPTION_NEGATABLE) == INPUT_OPTION_NEGATABLE
}

func (o *InputOption) IsValueRequired() bool {
	return (o.mode & INPUT_OPTION_REQUIRED) == INPUT_OPTION_REQUIRED
}

func (o *InputOption) IsValueOptional() bool {
	return (o.mode & INPUT_OPTION_OPTIONAL) == INPUT_OPTION_OPTIONAL
}

func (o *InputOption) AcceptValue() bool {
	return o.IsValueRequired() || o.IsValueOptional()
}

func (o *InputOption) SetDefaultValue(value InputType) {
	if (o.mode&INPUT_OPTION_BOOLEAN) == INPUT_OPTION_BOOLEAN {
		panic("Cannot set a default value when using InputOption.BOOLEAN mode.")
	}

	if o.IsArray() {
		str, isStr := value.(string)
		_, isArr := value.([]string)

		if isStr {
			value = []string{str}
		} else if !isArr {
			panic("A default value for an array option must be an array.")
		}
	}

	if o.AcceptValue() || o.IsNegatable() {
		o.defaultValue = value
	} else {
		o.defaultValue = false
	}
}

func (o *InputOption) GetName() string {
	return o.name
}

func (o *InputOption) GetShortcut() string {
	return o.shortcut
}

func (o *InputOption) GetDefaultValue() InputType {
	return o.defaultValue
}

func (o *InputOption) GetDescription() string {
	return o.description
}

func (o *InputOption) Equals(opt *InputOption) bool {
	if opt == nil {
		return false
	}

	return opt.GetName() == o.GetName() &&
		opt.GetShortcut() == o.GetShortcut() &&
		opt.GetDefaultValue() == o.GetDefaultValue() &&
		opt.IsNegatable() == o.IsNegatable() &&
		opt.IsArray() == o.IsArray() &&
		opt.IsValueRequired() == o.IsValueRequired() &&
		opt.IsValueOptional() == o.IsValueOptional()
}

func InputTypeToArray(value InputType) []string {
	arr, isArr := value.([]string)
	if isArr {
		return arr
	}

	str, isStr := value.(string)
	if isStr {
		return []string{str}
	}

	return make([]string, 0)
}

func InputTypeToString(value InputType) string {
	str, isStr := value.(string)
	if isStr {
		return str
	}

	return ""
}
