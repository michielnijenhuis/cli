package input

import (
	"fmt"
	"strings"
)

type ObjectInput struct {
	Input
	parameters map[string]InputType
}

func NewObjectInput(parameters map[string]InputType, definition *InputDefinition) (*ObjectInput, error) {
	baseInput, err := NewInput(definition)
	objectInput := &ObjectInput{
		Input:      *baseInput,
		parameters: parameters,
	}

	return objectInput, err
}

func (input *ObjectInput) GetFirstArgument() InputType {
	for param, value := range input.parameters {
		if strings.HasPrefix(param, "-") {
			continue
		}

		return value
	}

	return ""
}

func (input *ObjectInput) HasParameterOption(value string, onlyParams bool) bool {
	for k := range input.parameters {
		if onlyParams && k == "--" {
			return false
		}

		if k == value {
			return true
		}
	}

	return false
}

func (input *ObjectInput) GetParameterOption(value string, defaultValue InputType, onlyParams bool) InputType {
	for k, v := range input.parameters {
		if onlyParams && k == "--" {
			return defaultValue
		}

		if k == value {
			return v
		}
	}

	return defaultValue
}

func (input *ObjectInput) ToString() string {
	params := make([]string, 0)

	for param, val := range input.parameters {
		if param != "" && strings.HasPrefix(param, "-") {
			glue := " "
			if strings.HasPrefix(param, "--") {
				glue = "="
			}

			arr, isArr := val.([]string)
			if isArr {
				for _, v := range arr {
					res := param
					if v != "" {
						res += glue + input.EscapeToken(v)
					}
					params = append(params, res)
				}
			} else {
				str := val.(string)
				res := param
				if str != "" {
					res += glue + input.EscapeToken(str)
				}
				params = append(params, res)
			}
		} else {
			arr, isArr := val.([]string)
			str, isStr := val.(string)

			if isArr {
				mapped := make([]string, 0, len(arr))
				for _, v := range arr {
					mapped = append(mapped, input.EscapeToken(v))
				}
				params = append(params, strings.Join(mapped, " "))
			} else if isStr {
				params = append(params, input.EscapeToken(str))
			} else {
				params = append(params, input.EscapeToken(""))
			}
		}
	}

	return strings.Join(params, " ")
}

func (input *ObjectInput) Parse() error {
	for key, value := range input.parameters {
		if key == "--" {
			return nil
		}

		if strings.HasPrefix(key, "--") {
			err := input.addLongOption(key[2:], value)
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(key, "-") {
			err := input.addShortOption(key[1:], value)
			if err != nil {
				return err
			}
		} else {
			err := input.addArgument(key, value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (input *ObjectInput) addShortOption(shortcut string, value InputType) error {
	if !input.definition.HasShortcut(shortcut) {
		return fmt.Errorf("the \"-%s\" option does not exist", shortcut)
	}

	opt, e := input.definition.GetOptionForShortcut(shortcut)
	if e != nil {
		return e
	}

	return input.addLongOption(opt.GetName(), value)
}

func (input *ObjectInput) addLongOption(name string, value InputType) error {
	if !input.definition.HasOption(name) {
		if !input.definition.HasNegation(name) {
			return fmt.Errorf("the \"--%s\" option does not exist", name)
		}

		optName := input.definition.NegationToName(name)
		input.options[optName] = false

		return nil
	}

	opt, e := input.definition.GetOption(name)
	if e != nil {
		return e
	}

	if value == "" || value == nil {
		if opt.IsValueRequired() {
			return fmt.Errorf("the \"--%s\" option requires a value", name)
		}

		if !opt.IsValueOptional() {
			value = true
		}
	}

	input.options[name] = value
	return nil
}

func (input *ObjectInput) addArgument(name string, value InputType) error {
	if !input.definition.HasArgument(name) {
		return fmt.Errorf("the \"%s\" argument does not exist", name)
	}

	input.arguments[name] = value
	return nil
}
