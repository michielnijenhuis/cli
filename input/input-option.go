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
	Name         string
	Shortcut     string
	Description  string
	Mode         InputOptionMode
	DefaultValue InputType
	Validator    InputValidator
	constructed  bool
}

func NewInputOption(name string, shortcut string, mode InputOptionMode, description string) *InputOption {
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
		Name:         name,
		Shortcut:     shortcut,
		Description:  description,
		Mode:         mode,
		DefaultValue: nil,
		Validator:    nil,
		constructed:  true,
	}

	if o.IsArray() && !o.AcceptValue() {
		panic("Impossible to have an option mode IS_ARRAY if the option does not accept a value.")
	}

	if o.IsNegatable() && o.AcceptValue() {
		panic("Impossible to have an option mode NEGATABLE if the option also accepts a value.")
	}

	return o
}

func (o *InputOption) IsArray() bool {
	return (o.Mode & INPUT_OPTION_IS_ARRAY) == INPUT_OPTION_IS_ARRAY
}

func (o *InputOption) IsNegatable() bool {
	return (o.Mode & INPUT_OPTION_NEGATABLE) == INPUT_OPTION_NEGATABLE
}

func (o *InputOption) IsValueRequired() bool {
	return (o.Mode & INPUT_OPTION_REQUIRED) == INPUT_OPTION_REQUIRED
}

func (o *InputOption) IsValueOptional() bool {
	return (o.Mode & INPUT_OPTION_OPTIONAL) == INPUT_OPTION_OPTIONAL
}

func (o *InputOption) AcceptValue() bool {
	return o.IsValueRequired() || o.IsValueOptional()
}

func (o *InputOption) SetDefaultValue(value InputType) *InputOption {
	if (o.Mode&INPUT_OPTION_BOOLEAN) == INPUT_OPTION_BOOLEAN && value != "" && value != nil {
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
		o.DefaultValue = value
	} else {
		o.DefaultValue = false
	}

	return o
}

func (o *InputOption) SetValidator(validator InputValidator) *InputOption {
	o.Validator = validator
	return o
}

func (o *InputOption) Equals(opt *InputOption) bool {
	if opt == nil {
		return false
	}

	return opt.Name == o.Name &&
		opt.Shortcut == o.Shortcut &&
		opt.DefaultValue == o.DefaultValue &&
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

func (o *InputOption) WasConstructed() bool {
	return o.constructed
}

func (o *InputOption) Clone() *InputOption {
	c := NewInputOption(o.Name, o.Shortcut, o.Mode, o.Description)
	if o.DefaultValue != nil {
		c.SetDefaultValue(o.DefaultValue)
	}
	if o.Validator != nil {
		c.SetValidator(o.Validator)
	}
	return c
}
