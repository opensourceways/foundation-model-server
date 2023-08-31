package service

import (
	"github.com/opensourceways/foundation-model-server/inferenceqa/domain/flowcontroller"
	"github.com/opensourceways/foundation-model-server/inferenceqa/domain/moderation"
	"github.com/opensourceways/foundation-model-server/inferenceqa/domain/qa"
)

type Question = qa.Question

type QAService interface {
	Ask(*Question) error
	Models() []string
}

func NewQAService(
	m moderation.Moderation,
	fc flowcontroller.FlowController,
	qaObj qa.QA,
) QAService {
	return &qaService{
		m:     m,
		fc:    fc,
		qaObj: qaObj,
	}
}

type qaService struct {
	m     moderation.Moderation
	fc    flowcontroller.FlowController
	qaObj qa.QA
}

func (impl *qaService) Ask(q *Question) error {
	if err := impl.m.CheckText(q.Question.Question()); err != nil {
		return err
	}

	f := func() error {
		return impl.qaObj.Ask(q)
	}

	err := impl.fc.Do(f)

	return err
}

func (impl *qaService) Models() []string {
	return impl.qaObj.AllModels()
}
