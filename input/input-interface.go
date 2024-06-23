package input

type InputInterface interface {
	GetFirstArgument() InputType
	HasParameterOption(value string, onlyParams bool) bool
	GetParameterOption(value string, defaultValue InputType, onlyParams bool) InputType
	Bind(definition *InputDefinition) error
	Validate() error
	GetArguments() map[string]InputType
	GetStringArgument(name string) (string, error)
	GetArrayArgument(name string) ([]string, error)
	SetArgument(name string, value InputType) error
	HasArgument(name string) bool
	GetOptions() map[string]InputType
	GetStringOption(name string) (string, error)
	GetBoolOption(name string) (bool, error)
	GetArrayOption(name string) ([]string, error)
	SetOption(name string, value InputType) error
	HasOption(name string) bool
	IsInteractive() bool
	SetInteractive(interactive bool)
	ToString() string
}
