package cli

type ErrorWithAlternatives interface {
	error
	Alternatives() []string
}

type QuestionValidator[T any] func(T) (T, error)

type QuestionNormalizer[T any] func(string) T
