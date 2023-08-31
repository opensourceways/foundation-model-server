package dp

import (
	"errors"

	"github.com/opensourceways/foundation-model-server/utils"
)

var maxLengthOfQuestion int

func Init(n int, v ModelNameValidator) {
	modelNameValidator = v
	maxLengthOfQuestion = n
}

type Question interface {
	Question() string
}

func NewQuestion(v string) (Question, error) {
	if v == "" || utils.StrLen(v) > maxLengthOfQuestion {
		return nil, errors.New("invalid question")
	}

	return nil, nil
}

type question string

func (v question) Question() string {
	return string(v)
}
