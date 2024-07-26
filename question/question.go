package question

type QuestionValidator[T any] func(T) T
type QuestionNormalizer[T any] func(string) T

type Question[T any] struct {
	attempts       uint
	hidden         bool
	hiddenFallback bool
	validator      QuestionValidator[T]
	normalizer     QuestionNormalizer[T]
	trimmable      bool
	multiline      bool
	question       string
	defaultValue   T
}

func NewQuestion[T any](question string, defaultValue T) *Question[T] {
	return &Question[T]{
		attempts:       0,
		hidden:         false,
		hiddenFallback: false,
		validator:      nil,
		normalizer:     nil,
		trimmable:      false,
		multiline:      true,
		question:       question,
		defaultValue:   defaultValue,
	}
}

func (q *Question[any]) Question() string {
	return q.question
}

func (q *Question[any]) Default() any {
	return q.defaultValue
}

func (q *Question[any]) IsMultiline() bool {
	return q.multiline
}

func (q *Question[any]) SetMultiline(multiline bool) {
	q.multiline = multiline
}

func (q *Question[any]) IsHidden() bool {
	return q.hidden
}

func (q *Question[any]) SetHidden(hidden bool) {
	q.hidden = hidden
}

func (q *Question[any]) IsHiddenFallback() bool {
	return q.hiddenFallback
}

func (q *Question[any]) SetHiddenFallback(hiddenFallback bool) {
	q.hiddenFallback = hiddenFallback
}

func (q *Question[any]) SetValidator(validator QuestionValidator[any]) {
	q.validator = validator
}

func (q *Question[any]) Validator() QuestionValidator[any] {
	return q.validator
}

func (q *Question[any]) SetMaxAttempts(maxAttempts uint) {
	q.attempts = maxAttempts
}

func (q *Question[any]) MaxAttempts() uint {
	return q.attempts
}

func (q *Question[any]) SetNormalizer(normalizer QuestionNormalizer[any]) {
	q.normalizer = normalizer
}

func (q *Question[any]) Normalizer() QuestionNormalizer[any] {
	return q.normalizer
}

func (q *Question[any]) IsTrimmable() bool {
	return q.trimmable
}

func (q *Question[any]) SetTrimmable(trimmable bool) {
	q.trimmable = trimmable
}
