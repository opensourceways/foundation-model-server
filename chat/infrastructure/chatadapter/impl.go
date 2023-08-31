package chatadapter

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	port "github.com/opensourceways/foundation-model-server/chat/domain/chat"
	"github.com/opensourceways/foundation-model-server/utils"
)

var instance *chatAdapter

func Init(cfg *Config) (*chatAdapter, error) {
	v := &chatAdapter{
		cfg:          *cfg,
		cli:          utils.NewHttpClient(3, 1),
		stop:         make(chan struct{}),
		stopped:      make(chan struct{}),
		modelAddress: map[string]string{},
	}

	if err := v.refreshModels(); err != nil {
		return nil, err
	}

	v.start()

	instance = v

	return v, nil
}

func ChatAdapter() *chatAdapter {
	return instance
}

type chatAdapter struct {
	cfg          Config
	cli          utils.HttpClient
	mutex        sync.RWMutex
	stop         chan struct{}
	stopped      chan struct{}
	allModels    []string
	modelAddress map[string]string
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

	req, err := http.NewRequest(http.MethodPost, impl.cfg.ChatURL, buf)
	if err != nil {
		return err
	}

	cli := utils.NewHttpClient(3, 180)

	return cli.SendAndHandle(req, func(h http.Header, respBody io.Reader) error {
		st := streamTransfer{
			input: respBody,
		}

		q.SteamWrite(st.readAndWriteOnce)

		return nil
	})
}

func (impl *chatAdapter) AllModels() (r []string) {
	impl.mutex.RLock()
	r = append(r, impl.allModels...)
	impl.mutex.Unlock()

	return
}

func (impl *chatAdapter) IsValidModelName(m string) (b bool) {
	impl.mutex.RLock()
	_, b = impl.modelAddress[m]
	impl.mutex.Unlock()

	return
}

func (impl *chatAdapter) MaxLengthOfQuestion() int {
	return impl.cfg.MaxLengthOfQuestion
}

func (impl *chatAdapter) start() {
	go impl.watch()
}

func (impl *chatAdapter) Exit() {
	close(impl.stop)

	<-impl.stopped
}

func (impl *chatAdapter) watch() {
	interval := time.Minute
	timer := time.NewTimer(interval)

	defer func() {
		timer.Stop()

		close(impl.stopped)
	}()

	for {
		select {
		case <-impl.stop:
			return

		case <-timer.C:
			if err := impl.refreshModels(); err != nil {
				// TODO log it
			}

			timer.Reset(interval)
		}
	}
}

func (impl *chatAdapter) refreshModels() error {
	if err := impl.refreshWorkers(); err != nil {
		return err
	}

	v, err := impl.listModels()
	if err != nil {
		return err
	}
	if len(v) == 0 {
		return errors.New("no models")
	}

	r := map[string]string{}
	for _, m := range v {
		addr, err := impl.getModelAdress(m)
		if err != nil {
			return err
		}

		r[m] = addr
	}

	impl.mutex.Lock()

	impl.allModels = v
	impl.modelAddress = r

	impl.mutex.Unlock()

	return nil
}

func (impl *chatAdapter) refreshWorkers() error {
	req, err := http.NewRequest(http.MethodPost, impl.cfg.RefreshAllWorkersURL, nil)
	if err != nil {
		return err
	}

	_, err = impl.cli.ForwardTo(req, nil)

	return err
}

func (impl *chatAdapter) listModels() ([]string, error) {
	req, err := http.NewRequest(http.MethodPost, impl.cfg.ListModelsURL, nil)
	if err != nil {
		return nil, err
	}

	v := &listModelsResp{}

	_, err = impl.cli.ForwardTo(req, v)

	return v.Models, err
}

func (impl *chatAdapter) getModelAdress(m string) (string, error) {
	buf := &bytes.Buffer{}
	if err := jsonMarshal(&getModelAdressReq{m}, buf); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, impl.cfg.ListModelsURL, buf)
	if err != nil {
		return "", err
	}

	v := &getModelAdressResp{}

	if _, err = impl.cli.ForwardTo(req, v); err != nil {
		return "", err
	}

	if v.Address == "" {
		return "", errors.New("no address")
	}

	return v.Address, nil
}

type listModelsResp struct {
	Models []string `json:"models"`
}

type getModelAdressReq struct {
	Model string `json:"model"`
}

type getModelAdressResp struct {
	Address string `json:"address"`
}

type chatRequest struct {
	Models string `json:"models"`
	Prompt string `json:"prompt"`

	port.QuestionParameter
}

func jsonMarshal(t interface{}, buffer *bytes.Buffer) error {
	enc := json.NewEncoder(buffer)
	enc.SetEscapeHTML(false)

	return enc.Encode(t)
}
