package cli

type Question[T any] struct {
	Query           string
	Attempts        int
	Hidden          bool
	HiddenFallback  bool
	Validator       QuestionValidator[T]
	Normalizer      QuestionNormalizer[T]
	PreventTrimming bool
	Multiline       bool
	DefaultValue    T
}

func (q *Question[any]) DefaultNormalizer() QuestionNormalizer[any] {
	return nil
}

func (q *Question[any]) DefaultValidator() QuestionValidator[any] {
	return nil
}

func (q *Question[any]) IsQuestion() bool {
	return true
}
