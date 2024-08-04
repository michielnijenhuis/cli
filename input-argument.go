package cli

import (
	"fmt"
)

const (
	InputArgumentRequired = 1
	InputArgumentOptional = 2
	InputArgumentIsArray  = 4
)

type InputArgumentMode uint8

type InputArgument struct {
	Name         string
	Description  string
	Mode         InputArgumentMode
	DefaultValue InputType
	Validator    InputValidator
	initialized  bool
}

func (a *InputArgument) init() {
	if a.initialized {
		return
	}

	a.initialized = true

	if a.Mode == 0 {
		a.Mode = InputArgumentOptional
	} else if a.Mode > 7 || a.Mode < 1 {
		panic(fmt.Sprintf("Argument mode \"%d\" is not valid.", a.Mode))
	}

	if a.DefaultValue != nil {
		a.SetDefaultValue(a.DefaultValue)
	}
}

func (a *InputArgument) IsRequired() bool {
	a.init()

	return (a.Mode & InputArgumentRequired) == InputArgumentRequired
}

func (a *InputArgument) IsArray() bool {
	a.init()

	return (a.Mode & InputArgumentIsArray) == InputArgumentIsArray
}

func (a *InputArgument) SetDefaultValue(value InputType) *InputArgument {
	if a.IsRequired() && value != "" && value != nil {
		panic("Cannot set a default value except for OPTIONAL mode.")
	}

	isNil := value == nil
	str, isStr := value.(string)

	if a.IsArray() {
		arr, isArr := value.([]string)

		if isNil {
			arr = make([]string, 0)
		} else if isStr {
			arr = []string{str}
		} else if !isArr {
			panic("A default value for an array argument must be an array.")
		}

		a.DefaultValue = arr
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

func (a *InputArgument) Clone() *InputArgument {
	c := &InputArgument{
		Name:        a.Name,
		Mode:        a.Mode,
		Description: a.Description,
		initialized: a.initialized,
	}
	if a.DefaultValue != nil {
		c.SetDefaultValue(a.DefaultValue)
	}
	if a.Validator != nil {
		c.SetValidator(a.Validator)
	}
	return c
}
