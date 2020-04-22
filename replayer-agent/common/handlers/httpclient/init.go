package httpclient

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"
)

type HttpClient struct {
}

var Handler HttpClient

func Init() {
	Handler = HttpClient{}
}

//Get http get
func (hc *HttpClient) Get(ctx context.Context, url string) (*http.Response, []byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		tlog.Handler.Errorf(ctx, tlog.DLTagUndefined, err.Error())
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	tlog.Handler.Infof(ctx, tlog.DLTagUndefined, "resp=%v", resp)
	if err != nil {
		tlog.Handler.Errorf(ctx, tlog.DLTagUndefined, err.Error())
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return resp, body, nil
}

//Post http post
func (hc *HttpClient) Post(ctx context.Context, url string, jsonBytes []byte, timeout time.Duration, headers map[string]string) (*http.Response, []byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		tlog.Handler.Errorf(ctx, tlog.DLTagUndefined, err.Error())
		return nil, nil, err
	}
	//默认 application/json
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	//headers 优先级更高
	if len(headers) > 0 {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	client := &http.Client{}
	if timeout > 0 {
		client.Timeout = timeout
	}
	resp, err := client.Do(req)
	tlog.Handler.Infof(ctx, tlog.DLTagUndefined, "resp=%v", resp)
	if err != nil {
		tlog.Handler.Errorf(ctx, tlog.DLTagUndefined, err.Error())
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return resp, body, nil
}
