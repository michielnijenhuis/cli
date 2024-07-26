package question

import (
	"regexp"
)

type ConfirmationQuestion struct {
	*Question[bool]
	trueAnswerRegex *regexp.Regexp
}

func NewConfirmationQuestion(question string, defaultValue bool, trueAnswerRegex *regexp.Regexp) *ConfirmationQuestion {
	if trueAnswerRegex == nil {
		trueAnswerRegex = regexp.MustCompile("(?i)^y")
	}

	q := NewQuestion[bool](question, defaultValue)
	cq := &ConfirmationQuestion{Question: q, trueAnswerRegex: trueAnswerRegex}
	cq.SetNormalizer(cq.defaultNormalizer())

	return cq
}

func (q *ConfirmationQuestion) defaultNormalizer() QuestionNormalizer[bool] {
	defaultValue := q.Default()
	regex := q.trueAnswerRegex

	normalizer := func(answer string) bool {
		if answer == "" {
			return false
		}

		answerIsTrue := regex.MatchString(answer)
		if !defaultValue {
			return answer != "" && answerIsTrue
		}

		return answer == "" || answerIsTrue
	}

	return normalizer
}
