package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/didi/sharingan/recorder-agent/common/conf"
	"github.com/didi/sharingan/recorder-agent/common/httpclient"
	"github.com/didi/sharingan/recorder-agent/common/zap"
	"github.com/didi/sharingan/recorder-agent/record"
	"github.com/didi/sharingan/recorder-agent/server"
)

const (
	timeOut = 10 * time.Second
)

var (
	svr = server.New()
)

func init() {
	httpclient.Init()
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[panic] %s\n%s", r, debug.Stack())
		}
	}()

	svr.SetHTTPAddr(conf.Handler.GetString("http.http_addr"))
	svr.SetHTTPHandlerTimeout(conf.Handler.GetInt("http.handlerTimeout"))
	svr.SetHTTPReadTimeout(conf.Handler.GetInt("http.readTimeout"))
	svr.SetHTTPWriteTimeout(conf.Handler.GetInt("http.writeTimeout"))
	svr.SetHTTPIdleTimeout(conf.Handler.GetInt("http.idleTimeout"))
	svr.AddHTTPHandleFunc("/", indexHandler)

	log.Printf("[http] Server Running! addr:%s.\n", conf.Handler.GetString("http.http_addr"))
	if err := svr.Run(); err != nil {
		log.Fatalf("[http] Server failed! error:%s.", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("READ ERROR, err:" + err.Error()))
		return
	}

	isFilter, err := record.Fliter(string(buf))
	if err != nil {
		w.Write([]byte("FILTER ERROR, err:" + err.Error()))
		return
	}

	// Filter
	if isFilter {
		w.Write([]byte("FILTER"))
		return
	}

	// 日志收集，最终入ES
	url := conf.Handler.GetString("es_url.default")
	if _, _, err := httpclient.Handler.Post(r.Context(), url, buf, timeOut); err != nil {
		zap.Logger.Error(zap.Format(r.Context(), "ERROR", "send data to es err: %v", err))

		// TO log
		w.Write([]byte("OK"))
		zap.Logger.Info(string(buf))
		return
	}

	return
}
