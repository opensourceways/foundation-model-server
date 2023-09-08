package chatadapter

import (
	"bytes"
	"io"
	"net/http"

	port "github.com/opensourceways/foundation-model-server/chat/domain/chat"
	"github.com/opensourceways/foundation-model-server/utils"
)

var instance *chatAdapter

func Init(cfg *Config) error {
	w := modelWatcher{
		cfg:          cfg.modelConfig,
		cli:          utils.NewHttpClient(3, 1),
		stop:         make(chan struct{}),
		stopped:      make(chan struct{}),
		modelAddress: map[string]string{},
	}

	if err := w.refreshModels(); err != nil {
		return err
	}

	w.start()

	instance = &chatAdapter{
		w:   &w,
		cfg: cfg.chatConfig,
		cli: utils.NewHttpClient(3, cfg.Timeout),
	}

	return nil
}

func ChatAdapter() *chatAdapter {
	return instance
}

func Exit() {
	if instance != nil {
		instance.w.exit()

		instance = nil
	}
}

type chatAdapter struct {
	w   *modelWatcher
	cfg chatConfig
	cli utils.HttpClient
}

func (impl *chatAdapter) Ask(q *port.Question) error {
	v := chatRequest{
		Models:            q.ModelName.ModelName(),
		Prompt:            q.Question.Question(),
		QuestionParameter: q.Parameter,
	}

	buf := &bytes.Buffer{}
	if err := jsonMarshal(&v, buf); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(q.Ctx, http.MethodPost, impl.cfg.ChatURL, buf)
	if err != nil {
		return err
	}

	return impl.cli.SendAndHandle(req, func(h http.Header, respBody io.Reader) error {
		st := streamTransfer{
			input: respBody,
		}

		q.SteamWrite(st.readAndWriteOnce)

		return nil
	})
}

func (impl *chatAdapter) MaxLengthOfQuestion() int {
	return impl.cfg.MaxLengthOfQuestion
}

func (impl *chatAdapter) AllModels() (r []string) {
	return impl.w.getAllModels()
}

func (impl *chatAdapter) IsValidModelName(m string) (b bool) {
	return impl.w.hasModel(m)
}
