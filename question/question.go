package question

import "github.com/michielnijenhuis/cli/types"

type Question[T any] struct {
	attempts       int
	hidden         bool
	hiddenFallback bool
	validator      types.QuestionValidator[T]
	normalizer     types.QuestionNormalizer[T]
	trimmable      bool
	multiline      bool
	question       string
	defaultValue   T
}

func NewQuestion[T any](question string, defaultValue T) *Question[T] {
	return &Question[T]{
		attempts:       -1,
		hidden:         false,
		hiddenFallback: true,
		validator:      nil,
		normalizer:     nil,
		trimmable:      true,
		multiline:      false,
		question:       question,
		defaultValue:   defaultValue,
	}
}

func (q *Question[any]) GetQuestion() string {
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

func (q *Question[any]) SetValidator(validator types.QuestionValidator[any]) {
	q.validator = validator
}

func (q *Question[any]) Validator() types.QuestionValidator[any] {
	return q.validator
}

func (q *Question[any]) SetMaxAttempts(maxAttempts int) {
	q.attempts = maxAttempts
}

func (q *Question[any]) MaxAttempts() int {
	return q.attempts
}

func (q *Question[any]) SetNormalizer(normalizer types.QuestionNormalizer[any]) {
	q.normalizer = normalizer
}

func (q *Question[any]) Normalizer() types.QuestionNormalizer[any] {
	return q.normalizer
}

func (q *Question[any]) IsTrimmable() bool {
	return q.trimmable
}

func (q *Question[any]) SetTrimmable(trimmable bool) {
	q.trimmable = trimmable
}
