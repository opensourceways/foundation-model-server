package app

import "github.com/opensourceways/foundation-model-server/inferenceqa/domain/service"

type CmdToAskQuestion = service.Question

type QAService interface {
	Ask(*CmdToAskQuestion) error
	Models() []string
}

func NewQAService(s service.QAService) QAService {
	return &qaService{s}
}

type qaService struct {
	qa service.QAService
}

func (s *qaService) Ask(cmd *CmdToAskQuestion) error {
	return s.qa.Ask(cmd)
}

func (s *qaService) Models() []string {
	return s.qa.Models()
}
