package cli

import (
	"fmt"
	"strings"
)

const (
	InputOptionBool      = 1
	InputOptionRequired  = 2
	InputOptionOptional  = 4
	InputOptionIsArray   = 8
	InputOptionNegatable = 16
)

type InputOptionMode uint8

type InputOption struct {
	// Should not be prefixed with dashes (e.g. "foo" is ok, "--foo" is not)
	Name string
	// Pipe separated string of shortcuts (e.g. "f|F")
	Shortcut     string
	Description  string
	Mode         InputOptionMode
	DefaultValue InputType
	Validator    InputValidator
	validated    bool
}

func (o *InputOption) validate() {
	if o.validated {
		return
	}

	o.validated = true

	name, _ := strings.CutPrefix(o.Name, "--")
	o.Name = name

	if o.Mode == 0 {
		o.Mode = InputOptionOptional
	} else if o.Mode >= InputOptionNegatable<<1 || o.Mode < 1 {
		panic(fmt.Sprintf("option mode \"%d\" is not valid", o.Mode))
	}

	if o.IsArray() && !o.AcceptValue() {
		panic("impossible to have an option mode IS_ARRAY if the option does not accept a value")
	}

	if o.IsNegatable() && o.AcceptValue() {
		panic("impossible to have an option mode NEGATABLE if the option also accepts a value")
	}
}

func (o *InputOption) IsArray() bool {
	o.validate()
	return (o.Mode & InputOptionIsArray) == InputOptionIsArray
}

func (o *InputOption) IsNegatable() bool {
	o.validate()
	return (o.Mode & InputOptionNegatable) == InputOptionNegatable
}

func (o *InputOption) IsValueRequired() bool {
	o.validate()
	return (o.Mode & InputOptionRequired) == InputOptionRequired
}

func (o *InputOption) IsValueOptional() bool {
	o.validate()
	return (o.Mode & InputOptionOptional) == InputOptionOptional
}

func (o *InputOption) AcceptValue() bool {
	o.validate()
	return o.IsValueRequired() || o.IsValueOptional()
}

func (o *InputOption) SetDefaultValue(value InputType) *InputOption {
	o.validate()

	if (o.Mode&InputOptionBool) == InputOptionBool && value != "" && value != nil {
		panic("cannot set a default value when using InputOption.BOOLEAN mode")
	}

	if o.IsArray() {
		str, isStr := value.(string)
		_, isArr := value.([]string)

		if isStr {
			value = []string{str}
		} else if !isArr {
			panic("a default value for an array option must be an array")
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

func (o *InputOption) Clone() *InputOption {
	c := &InputOption{
		Name:        o.Name,
		Shortcut:    o.Shortcut,
		Mode:        o.Mode,
		Description: o.Description,
		validated:   o.validated,
	}
	if o.DefaultValue != nil {
		c.SetDefaultValue(o.DefaultValue)
	}
	if o.Validator != nil {
		c.SetValidator(o.Validator)
	}
	return c
}
