package cli

import (
	"regexp"
)

var TrueAnswerRegex = regexp.MustCompile("(?i)^y")

type ConfirmationQuestion struct {
	*Question[bool]
	trueAnswerRegex *regexp.Regexp
}

func (q *ConfirmationQuestion) DefaultNormalizer() QuestionNormalizer[bool] {
	defaultValue := q.DefaultValue
	regex := q.trueAnswerRegex
	if regex == nil {
		regex = TrueAnswerRegex
	}

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
