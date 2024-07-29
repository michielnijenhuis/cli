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
	name         string
	description  string
	mode         InputArgumentMode
	defaultValue InputType
	validator    InputValidator
}

func NewInputArgument(name string, mode InputArgumentMode, description string, defaultValue InputType, validator InputValidator) *InputArgument {
	if mode == 0 {
		mode = INPUT_ARGUMENT_OPTIONAL
	} else if mode > 7 || mode < 1 {
		panic(fmt.Sprintf("Argument mode \"%d\" is not valid.", mode))
	}

	a := &InputArgument{
		name:         name,
		description:  description,
		mode:         mode,
		defaultValue: nil,
		validator:    validator,
	}

	a.SetDefaultValue(defaultValue)

	return a
}

func (a *InputArgument) Name() string {
	return a.name
}

func (a *InputArgument) IsRequired() bool {
	return (a.mode & INPUT_ARGUMENT_REQUIRED) == INPUT_ARGUMENT_REQUIRED
}

func (a *InputArgument) IsArray() bool {
	return (a.mode & INPUT_ARGUMENT_IS_ARRAY) == INPUT_ARGUMENT_IS_ARRAY
}

func (a *InputArgument) SetDefaultValue(value InputType) {
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

		a.defaultValue = value
		return
	}

	if !isStr && !isNil {
		panic("InputArgument values may be of type string, []string, or nil.")
	}

	a.defaultValue = value
}

func (a *InputArgument) DefaultValue() InputType {
	return a.defaultValue
}

func (a *InputArgument) Description() string {
	return a.description
}
