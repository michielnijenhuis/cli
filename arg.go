package cli

type StringArg struct {
	Name        string
	Description string
	Value       string
	Required    bool
	Validator   func(string) error
}

type ArrayArg struct {
	Name        string
	Description string
	Value       []string
	Min         uint
	Validator   func([]string) error
}

type Arg interface {
	GetName() string
	GetDescription() string
	IsRequired() bool
	HasValue() bool
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
		if a.Validator != nil {
			return a.Validator(a.Value)
		}
		return nil
	case *ArrayArg:
		if a.Validator != nil {
			return a.Validator(a.Value)
		}
		return nil
	default:
		panic("invalid argument type")
	}
}
