package app

import "github.com/opensourceways/foundation-model-server/chat/domain/service"

type CmdToAskQuestion = service.Question

type ChatAppService interface {
	Ask(*CmdToAskQuestion) error
	Models() []string
}

func NewChatAppService(s service.ChatService) ChatAppService {
	return &chatAppService{s}
}

type chatAppService struct {
	qa service.ChatService
}

func (s *chatAppService) Ask(cmd *CmdToAskQuestion) error {
	return s.qa.Ask(cmd)
}

func (s *chatAppService) Models() []string {
	return s.qa.Models()
}
