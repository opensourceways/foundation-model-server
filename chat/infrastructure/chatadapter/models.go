package chatadapter

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	port "github.com/opensourceways/foundation-model-server/chat/domain/chat"
	"github.com/opensourceways/foundation-model-server/utils"
)

type modelWatcher struct {
	cfg          modelConfig
	cli          utils.HttpClient
	mutex        sync.RWMutex
	stop         chan struct{}
	stopped      chan struct{}
	allModels    []string
	modelAddress map[string]string
}

func (impl *modelWatcher) getAllModels() (r []string) {
	impl.mutex.RLock()
	r = append(r, impl.allModels...)
	impl.mutex.Unlock()

	return
}

func (impl *modelWatcher) hasModel(m string) (b bool) {
	impl.mutex.RLock()
	_, b = impl.modelAddress[m]
	impl.mutex.Unlock()

	return
}

func (impl *modelWatcher) start() {
	go impl.watch()
}

func (impl *modelWatcher) exit() {
	close(impl.stop)

	<-impl.stopped
}

func (impl *modelWatcher) watch() {
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
				logrus.Errorf("refesh models failed, err:%s", err.Error())
			}

			timer.Reset(interval)
		}
	}
}

func (impl *modelWatcher) refreshModels() error {
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

func (impl *modelWatcher) refreshWorkers() error {
	req, err := http.NewRequest(http.MethodPost, impl.cfg.RefreshAllWorkersURL, nil)
	if err != nil {
		return err
	}

	_, err = impl.cli.ForwardTo(req, nil)

	return err
}

func (impl *modelWatcher) listModels() ([]string, error) {
	req, err := http.NewRequest(http.MethodPost, impl.cfg.ListModelsURL, nil)
	if err != nil {
		return nil, err
	}

	v := &listModelsResp{}

	_, err = impl.cli.ForwardTo(req, v)

	return v.Models, err
}

func (impl *modelWatcher) getModelAdress(m string) (string, error) {
	buf := &bytes.Buffer{}
	if err := jsonMarshal(&getModelAdressReq{m}, buf); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, impl.cfg.GetWorkerAddressURL, buf)
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
