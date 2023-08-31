package controller

import (
	"github.com/opensourceways/foundation-model-server/inferenceqa/app"
	"github.com/opensourceways/foundation-model-server/inferenceqa/domain/dp"
)

type qaRequest struct {
	Question          string  `json:"question"              binding:"required"`
	ModelName         string  `json:"model_name"            binding:"required"`
	TopP              float32 `json:"top_p"`
	Temperature       float32 `json:"temperature"`
	RepetitionPenalty float32 `json:"repetition_penalty"`
	Stop              string  `json:"stop"`
	StopTokenIds      string  `json:"stop_token_ids"`
	Echo              int     `json:"echo"`
	MaxNewTokens      int     `json:"max_new_tokens"`
}

func (req *qaRequest) toCmd() (cmd app.CmdToAskQuestion, err error) {
	if cmd.Question, err = dp.NewQuestion(req.Question); err != nil {
		return
	}

	if cmd.ModelName, err = dp.NewModelName(req.ModelName); err != nil {
		return
	}

	p := &cmd.Parameter

	p.TopP = req.TopP
	p.Temperature = req.Temperature
	p.RepetitionPenalty = req.RepetitionPenalty

	p.Stop = req.Stop
	p.StopTokenIds = req.StopTokenIds

	p.Echo = req.Echo
	p.MaxNewTokens = req.MaxNewTokens

	return
}
