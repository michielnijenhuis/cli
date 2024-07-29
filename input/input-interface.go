package input

type InputInterface interface {
	FirstArgument() InputType
	HasParameterOption(value string, onlyParams bool) bool
	ParameterOption(value string, defaultValue InputType, onlyParams bool) InputType
	Bind(definition *InputDefinition)
	Parse() error
	Validate() error
	Arguments() map[string]InputType
	StringArgument(name string) (string, error)
	ArrayArgument(name string) ([]string, error)
	SetArgument(name string, value InputType) error
	HasArgument(name string) bool
	Options() map[string]InputType
	StringOption(name string) (string, error)
	BoolOption(name string) (bool, error)
	ArrayOption(name string) ([]string, error)
	SetOption(name string, value InputType) error
	HasOption(name string) bool
	IsInteractive() bool
	SetInteractive(interactive bool)
	ToString() string
}
