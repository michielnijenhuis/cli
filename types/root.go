package types

type ErrorWithAlternatives interface {
	error
	Alternatives() []string
}

type QuestionValidator[T any] func(T) (T, error)
type QuestionNormalizer[T any] func(string) T
type QuestionInterface[T any] interface {
	GetQuestion() string
	Default() T
	Normalizer() QuestionNormalizer[T]
	Validator() QuestionValidator[T]
	MaxAttempts() uint
	IsTrimmable() bool
	IsMultiline() bool
}

type StyleInterface interface {
	Title(message string)
	Section(message string)
	Listing(elements []string, prefix string)
	Text(messages []string)
	Success(messages []string)
	Err(messages []string)
	Warning(messages []string)
	Info(messages []string)
	Note(messages []string)
	Caution(messages []string)
	Table(headers []string, rows map[string]string)
	Ask(questions string, defaultValue string, validator func(string) bool) string
	AskHidden(question string, validator func(string) bool) string
	Confirm(question string, defaultValue bool) bool
	Choice(question string, choices map[string]string, defaultValue string) string
	NewLine(count int)
	ProgressStart(max uint)
	ProgressAdvance(step uint)
	ProgressFinish()
	Box(title string, body string, footer string, color string, info string)
}
