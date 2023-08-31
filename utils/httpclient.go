package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpClient struct {
	client     http.Client
	maxRetries int
}

func newClient(timeout int) http.Client {
	return http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
}

func NewHttpClient(n, timeout int) HttpClient {
	return HttpClient{
		maxRetries: n,
		client:     newClient(timeout),
	}
}

func (hc *HttpClient) SendAndHandle(req *http.Request, handle func(http.Header, io.Reader) error) error {
	resp, err := hc.do(req)
	if err != nil || resp == nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		rb, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("response has status:%s and body:%q", resp.Status, rb)
	}

	if handle != nil {
		return handle(resp.Header, resp.Body)
	}

	return nil
}

func (hc *HttpClient) ForwardTo(req *http.Request, jsonResp interface{}) (statusCode int, err error) {
	resp, err := hc.do(req)
	if err != nil || resp == nil {
		return
	}

	defer resp.Body.Close()

	if code := resp.StatusCode; code < 200 || code > 299 {
		statusCode = code

		var rb []byte
		if rb, err = ioutil.ReadAll(resp.Body); err == nil {
			err = fmt.Errorf("response has status:%s and body:%q", resp.Status, rb)
		}

		return
	}

	if jsonResp != nil {
		err = json.NewDecoder(resp.Body).Decode(jsonResp)
	}

	return
}

func (hc *HttpClient) do(req *http.Request) (resp *http.Response, err error) {
	if resp, err = hc.client.Do(req); err == nil {
		return
	}

	maxRetries := hc.maxRetries
	backoff := 10 * time.Millisecond

	for retries := 1; retries < maxRetries; retries++ {
		time.Sleep(backoff)
		backoff *= 2

		if resp, err = hc.client.Do(req); err == nil {
			break
		}
	}
	return
}
