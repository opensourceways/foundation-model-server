package chat

import (
	"io"

	"github.com/opensourceways/foundation-model-server/chat/domain/dp"
)

type Question struct {
	Question   dp.Question
	Parameter  QuestionParameter
	ModelName  dp.ModelName
	SteamWrite func(doOnce func(io.Writer) (bool, error))
}

// TODO it should check for each parameters
type QuestionParameter struct {
	StopTokenIds      []int   `json:"stop_token_ids"`
	MaxNewTokens      int     `json:"max_new_tokens"`
	RepetitionPenalty float32 `json:"repetition_penalty"`
	Temperature       float32 `json:"temperature"`
	TopP              float32 `json:"top_p"`
	Stop              string  `json:"stop"`
	Echo              bool    `json:"echo"`
}

type Chat interface {
	Ask(*Question) error
	AllModels() []string
	IsValidModelName(string) bool
	MaxLengthOfQuestion() int
}
