package cli

import (
	"fmt"
	"strings"
)

type StringArg struct {
	Name        string
	Description string
	Value       string
	Required    bool
	Options     []string
	Validator   func(string) error
}

type ArrayArg struct {
	Name        string
	Description string
	Value       []string
	Min         uint
	Options     []string
	Validator   func([]string) error
}

type Arg interface {
	GetName() string
	GetDescription() string
	IsRequired() bool
	HasValue() bool
	Opts() []string
}

func (a *StringArg) GetName() string {
	return a.Name
}

func (a *StringArg) GetDescription() string {
	return a.Description
}

func (a *StringArg) IsRequired() bool {
	return a.Required
}

func (a *StringArg) HasValue() bool {
	return a.Value != ""
}

func (a *StringArg) Opts() []string {
	return a.Options
}

func (a *ArrayArg) GetName() string {
	return a.Name
}

func (a *ArrayArg) GetDescription() string {
	return a.Description
}

func (a *ArrayArg) IsRequired() bool {
	return a.Min > 0
}

func (a *ArrayArg) HasValue() bool {
	return len(a.Value) > 0
}

func (a *ArrayArg) Opts() []string {
	return a.Options
}

func GetArgStringValue(arg Arg) string {
	if a, ok := arg.(*StringArg); ok {
		return a.Value
	}

	return ""
}

func GetArgArrayValue(arg Arg) []string {
	if a, ok := arg.(*ArrayArg); ok {
		return a.Value
	}

	return []string{}
}

func ValidateArg(arg Arg) error {
	switch a := arg.(type) {
	case *StringArg:
		if len(a.Options) > 0 {
			isValid := false
			for _, option := range a.Options {
				if option == a.Value {
					isValid = true
					break
				}
			}

			if !isValid {
				return fmt.Errorf("invalid option \"%s\". Expected one of: %s", a.Value, strings.Join(a.Options, ", "))
			}
		}

		if a.Validator != nil {
			return a.Validator(a.Value)
		}

		return nil
	case *ArrayArg:
		if len(a.Options) > 0 {
			for _, value := range a.Value {
				isValid := false
				for _, option := range a.Options {
					if option == value {
						isValid = true
						break
					}
				}

				if !isValid {
					return fmt.Errorf("invalid option \"%s\". Possible values: %s", value, strings.Join(a.Options, ", "))
				}
			}
		}

		if a.Validator != nil {
			return a.Validator(a.Value)
		}

		return nil
	default:
		panic("invalid argument type")
	}
}
