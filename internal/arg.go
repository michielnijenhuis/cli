package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ArgValidator[T any] func(T) error

type StringArg struct {
	Name        string
	Description string
	Value       string
	Required    bool
	Options     []string
	Validator   ArgValidator[string]
}

type ArrayArg struct {
	Name        string
	Description string
	Value       []string
	Min         uint
	Max         uint
	Options     []string
	Validator   ArgValidator[[]string]
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
		if a.Min > 0 && len(a.Value) < int(a.Min) {
			return fmt.Errorf("expected at least %d values for argument \"%s\"", a.Min, a.Name)
		}

		if a.Max > 0 && len(a.Value) > int(a.Max) {
			return fmt.Errorf("expected at most %d values for argument \"%s\"", a.Max, a.Name)
		}

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
		return errors.New("invalid argument type")
	}
}

func IsFile(s string) error {
	if strings.HasPrefix(s, ".") {
		abs, err := filepath.Abs(s)
		if err != nil {
			return fmt.Errorf("failed to determine absolute path: %w", err)
		}

		s = abs
	}

	fileInfo, err := os.Stat(s)
	if err != nil {
		return fmt.Errorf("value [%s] is not a valid path:\n\t- %w", s, err)
	}

	if fileInfo.IsDir() {
		return errors.New("path [%s] is a directory, not a file path")
	}

	return nil
}

func IsInt(s string) error {
	_, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("value [%s] is not a valid integer:\n\t- %w", s, err)
	}

	return nil
}

func IsPositiveOrZeroInt(s string) error {
	i, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("value [%s] is not a valid integer:\n\t- %w", s, err)
	}

	if i < 0 {
		return fmt.Errorf("value [%s] is not a valid positive or zero integer.", s)
	}

	return nil
}

func IsPositiveInt(s string) error {
	i, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("value [%s] is not a valid integer:\n\t- %w", s, err)
	}

	if i <= 0 {
		return fmt.Errorf("value [%s] is not a valid positive integer.", s)
	}

	return nil
}

func ValidateMany[T any](fns ...ArgValidator[T]) ArgValidator[T] {
	return func(v T) error {
		for _, fn := range fns {
			err := fn(v)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
