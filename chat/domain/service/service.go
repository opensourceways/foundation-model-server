package service

import (
	"github.com/opensourceways/foundation-model-server/chat/domain/chat"
	"github.com/opensourceways/foundation-model-server/common/domain/flowcontroller"
	"github.com/opensourceways/foundation-model-server/common/domain/moderation"
)

type Question = chat.Question

type ChatService interface {
	Ask(*Question) error
	Models() []string
}

func NewChatService(
	m moderation.Moderation,
	fc flowcontroller.FlowController,
	s chat.Chat,
) ChatService {
	return &chatService{
		m:  m,
		fc: fc,
		s:  s,
	}
}

type chatService struct {
	m  moderation.Moderation
	fc flowcontroller.FlowController
	s  chat.Chat
}

func (impl *chatService) Ask(q *Question) error {
	if err := impl.m.CheckText(q.Question.Question()); err != nil {
		return err
	}

	f := func() error {
		return impl.s.Ask(q)
	}

	err := impl.fc.Do(f)

	return err
}

func (impl *chatService) Models() []string {
	return impl.s.AllModels()
}
