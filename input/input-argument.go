package input

import (
	"fmt"
)

const (
	INPUT_ARGUMENT_REQUIRED = 1
	INPUT_ARGUMENT_OPTIONAL = 2
	INPUT_ARGUMENT_IS_ARRAY = 4
)

type InputArgumentMode uint8

type InputArgument struct {
	Name         string
	Description  string
	Mode         InputArgumentMode
	DefaultValue InputType
	Validator    InputValidator
	constructed  bool
}

func NewInputArgument(name string, mode InputArgumentMode, description string) *InputArgument {
	if mode == 0 {
		mode = INPUT_ARGUMENT_OPTIONAL
	} else if mode > 7 || mode < 1 {
		panic(fmt.Sprintf("Argument mode \"%d\" is not valid.", mode))
	}

	a := &InputArgument{
		Name:         name,
		Description:  description,
		Mode:         mode,
		DefaultValue: nil,
		Validator:    nil,
		constructed:  true,
	}

	return a
}

func (a *InputArgument) IsRequired() bool {
	return (a.Mode & INPUT_ARGUMENT_REQUIRED) == INPUT_ARGUMENT_REQUIRED
}

func (a *InputArgument) IsArray() bool {
	return (a.Mode & INPUT_ARGUMENT_IS_ARRAY) == INPUT_ARGUMENT_IS_ARRAY
}

func (a *InputArgument) SetDefaultValue(value InputType) *InputArgument {
	if a.IsRequired() && value != "" && value != nil {
		panic("Cannot set a default value except for OPTIONAL mode.")
	}

	isNil := value == nil
	str, isStr := value.(string)

	if a.IsArray() {
		_, isArr := value.([]string)

		if isNil {
			value = make([]string, 0)
		} else if isStr {
			arr := make([]string, 0)
			value = append(arr, str)
		} else if !isArr {
			panic("A default value for an array argument must be an array.")
		}

		a.DefaultValue = value
		return a
	}

	if !isStr && !isNil {
		panic("InputArgument values may be of type string, []string, or nil.")
	}

	a.DefaultValue = value
	return a
}

func (a *InputArgument) SetValidator(validator InputValidator) *InputArgument {
	a.Validator = validator
	return a
}

func (a *InputArgument) WasConstructed() bool {
	return a.constructed
}

func (a *InputArgument) Clone() *InputArgument {
	c := NewInputArgument(a.Name, a.Mode, a.Description)
	if a.DefaultValue != nil {
		c.SetDefaultValue(a.DefaultValue)
	}
	if a.Validator != nil {
		c.SetValidator(a.Validator)
	}
	return c
}
