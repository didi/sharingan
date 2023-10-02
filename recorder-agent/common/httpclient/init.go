package httpclient

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/didi/sharingan/recorder-agent/common/zap"
)

type HttpClient struct {
}

var Handler HttpClient

func Init() {
	Handler = HttpClient{}
}

//Post http post
func (hc *HttpClient) Post(ctx context.Context, url string, jsonBytes []byte, timeout time.Duration) (*http.Response, []byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		zap.Logger.Error(zap.Format(ctx, "ERROR", err.Error()))
		return nil, nil, err
	}
	req.SetBasicAuth("username", "password")
	//默认 application/json
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	client := &http.Client{}
	if timeout > 0 {
		client.Timeout = timeout
	}
	resp, err := client.Do(req)
	zap.Logger.Info(zap.Format(ctx, "INFO", "resp=%v", resp))
	if err != nil {
		zap.Logger.Error(zap.Format(ctx, "ERROR", err.Error()))
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return resp, body, nil
}
