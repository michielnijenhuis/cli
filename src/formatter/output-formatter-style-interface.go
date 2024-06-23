package formatter

type OutputFormatterStyleInterface interface {
	SetForeground(color string)
	SetBackground(color string)
	SetOption(option string)
	UnsetOption(option string)
	SetOptions(options []string)
	Apply(text string)
	Clone() OutputFormatterStyleInterface
}
